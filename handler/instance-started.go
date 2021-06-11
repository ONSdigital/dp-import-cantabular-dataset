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
	ctblr      cantabularClient
	datasets   datasetAPIClient
	recipes    recipeAPIClient
	authToken string
	// Kafka Producer
}

func NewInstanceStarted(c cantabularClient, r recipeAPIClient, d datasetAPIClient, t string) *InstanceStarted {
	return &InstanceStarted{
		ctblr: c,
		recipes: r,
		datasets: d,
		authToken: t,
	}
}

// Handle takes a single event.
func (h *InstanceStarted) Handle(ctx context.Context, cfg *config.Config, e *event.InstanceStarted) error {
	r, err := h.recipes.GetRecipe(ctx, "", h.authToken, e.RecipeID)
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
		DatasetName: r.CantabularBlob, // "e.g. 'Example' or 'Teaching-Dataset'"
		Variables: codelists,
		Categories: true,
	}

	resp, err := h.ctblr.GetCodebook(ctx, req)
	if err != nil{
		return fmt.Errorf("failed to get codebook from Cantabular: %w", err)
	}

	log.Info(ctx, "Successfully got Codebook", log.Data{
		"datablob": resp.Dataset,
		"num_variables": len(resp.Codebook),
	})

	// Should be implemented correctly up to here, beyond this is WIP

	iReq, err := h.createInstanceRequest(resp.Codebook)
	if err != nil{
		return fmt.Errorf("failed to transform codebook to dataset: %w", err)
	}

	fmt.Println(iReq)

	/*
	if err := h.datasets.PutDataset(ctx, "", h.authToken, e.CollectionID, e.DatablobName, *ds); err != nil{
		return fmt.Errorf("failed to save dataset: %w", err)
	}*/

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
		codelists = append(codelists, cl.Name)
	}

	return codelists, nil
}

func (h *InstanceStarted) createInstanceRequest(cb cantabular.Codebook) (*dataset.UpdateInstance, error){
	req := dataset.UpdateInstance{
		Edition: "2021",
		CSVHeader: []string{"ftb_table"},
	}

	for _, v := range cb{
		d := dataset.VersionDimension{
			ID: v.Name,
			URL: fmt.Sprintf("$DATASET_API_HOST:$DATASET_API_PORT/code-lists/%s", v.Name),
			Label: v.Label,
			Name:v.Label,
		}
		req.Dimensions = append(req.Dimensions, d)
		req.CSVHeader  = append(req.CSVHeader, v.Name)
	}

	return &req, nil
}

func (h *InstanceStarted) triggerImportDimensionOptions() error {
	// Kafka producer loop
	return nil
}