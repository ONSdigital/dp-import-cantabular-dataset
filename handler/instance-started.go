package handler

import (
	"context"
	"fmt"
	"errors"

	"github.com/ONSdigital/dp-import-cantabular-dataset/config"
	"github.com/ONSdigital/dp-import-cantabular-dataset/event"
	"github.com/ONSdigital/dp-api-clients-go/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/recipe"
	"github.com/ONSdigital/log.go/v2/log"
)

// InstanceStarted is the handler for the InstanceStarted event
type InstanceStarted struct {
	cfg        config.Config
	ctblr      cantabularClient
	datasets   datasetAPIClient
	recipes    recipeAPIClient
	producer   kafkaProducer
}

func NewInstanceStarted(cfg config.Config, c cantabularClient, r recipeAPIClient, d datasetAPIClient, p kafkaProducer) *InstanceStarted {
	return &InstanceStarted{
		cfg:      cfg,
		ctblr:    c,
		recipes:  r,
		datasets: d,
		producer: p,
	}
}

// Note to self: why pass cfg rathe than have as part of struct?
// Handle takes a single event.
func (h *InstanceStarted) Handle(ctx context.Context, e *event.InstanceStarted) error {
	r, err := h.recipes.GetRecipe(ctx, "", h.cfg.ServiceAuthToken, e.RecipeID)
	if err != nil{
		return fmt.Errorf("failed to get recipe: %w", err)
	}

	log.Info(ctx, "Successfully got Recipe", log.Data{"recipe_alias": r.Alias})

	i, err := h.getInstanceFromRecipe(ctx, r)
	if err != nil{
		return fmt.Errorf("failed to get instance from recipe: %s", err)
	}

	log.Info(ctx, "Successfully got instance", log.Data{"instance_title": i.Title})

	codelists, err := h.getCodeListsFromInstance(i)
	if err != nil{
		return fmt.Errorf("failed to get code-lists (dimensions) from recipe instance: %s", err)
	}

	log.Info(ctx, "Successfully got codelists", log.Data{"num_codelists": len(codelists)})

	req := cantabular.GetCodebookRequest{
		DatasetName: r.CantabularBlob,
		Variables:   codelists,
		Categories:  false,
	}

	// Validation happens here, if any variables are incorrect, will throw an error
	resp, err := h.ctblr.GetCodebook(ctx, req)
	if err != nil{
		return fmt.Errorf("failed to get codebook from Cantabular: %w", err)
	}

	log.Info(ctx, "Successfully got Codebook", log.Data{
		"datablob":      resp.Dataset,
		"num_variables": len(resp.Codebook),
		"codebook":      resp.Codebook,
	})

	ireq := h.createUpdateInstanceRequest(resp.Codebook, e)
	
	log.Info(ctx, "Updating instance", log.Data{
		"instance_id":    ireq.InstanceID,
		"csv_headers":    ireq.CSVHeader,
		"edition":        ireq.Edition,
		"num_dimensions": len(ireq.Dimensions),
	})

	if err := h.datasets.PutInstance(ctx, "", h.cfg.ServiceAuthToken, "", e.InstanceID, ireq); err != nil{
		return fmt.Errorf("failed to update instance: %w", err)
	}

	log.Info(ctx, "Triggering dimension options import")

	if err := h.triggerImportDimensionOptions(); err != nil{
		return fmt.Errorf("failed to import dimension options: %w", err)
	}

	return nil
}

func (h *InstanceStarted) getInstanceFromRecipe(ctx context.Context, r *recipe.Recipe) (*recipe.Instance, error){
	if len(r.OutputInstances) < 1{
		return nil, errors.New("no instances found in recipe")
	}

	if len(r.OutputInstances) > 1{
		log.Warn(ctx, "more than one instance found in recipe. Defaulting to instances[0]", log.Data{
			"num_instances": len(r.OutputInstances),
		})
	}

	return &r.OutputInstances[0], nil
}

func (h *InstanceStarted) getCodeListsFromInstance(i *recipe.Instance) ([]string, error){
	if len(i.CodeLists) < 1{
		return nil, fmt.Errorf("no code-lists (dimensions) found in instance")
	}

	var codelists []string
	for _, cl := range i.CodeLists{
		codelists = append(codelists, cl.ID)
	}

	return codelists, nil
}

func (h *InstanceStarted) createUpdateInstanceRequest(cb cantabular.Codebook, e *event.InstanceStarted) dataset.UpdateInstance{
	req := dataset.UpdateInstance{
		Edition: "2021",
		CSVHeader: []string{"ftb_table"},
		InstanceID: e.InstanceID,
	}

	for _, v := range cb{
		sourceName := v.Name

		if len(v.MapFrom) > 0{
			if len(v.MapFrom[0].SourceNames) > 0{
				sourceName = v.MapFrom[0].SourceNames[0]
			}
		}

		d := dataset.VersionDimension{
			ID: sourceName,
			URL: fmt.Sprintf("%s/code-lists/%s", h.cfg.RecipeAPIURL, sourceName),
			Label: v.Label,
			Name:v.Label,
		}
		req.Dimensions = append(req.Dimensions, d)
		req.CSVHeader  = append(req.CSVHeader, v.Name)
	}

	return req
}

func (h *InstanceStarted) triggerImportDimensionOptions() error {
	// Kafka producer loop
	return nil
}
