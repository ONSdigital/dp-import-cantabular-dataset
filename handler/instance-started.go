package handler

import (
	"context"
	"fmt"
	"errors"

	"github.com/ONSdigital/dp-import-cantabular-dataset/config"
	"github.com/ONSdigital/dp-import-cantabular-dataset/models"
	"github.com/ONSdigital/dp-import-cantabular-dataset/event"
	"github.com/ONSdigital/dp-import-cantabular-dataset/transform"
	"github.com/ONSdigital/dp-import-cantabular-dataset/cantabular"
	"github.com/ONSdigital/log.go/v2/log"
)

// InstanceStarted ...
type InstanceStarted struct {
	ctblr      cantabularClient // *cantabular.Client
	datasets   datasetAPIClient // *dp-api-clients-go/dataset.Client
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
	r, err := h.recipes.Get(ctx, e.RecipeID)
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
		Variables: codelists,
		Categories: false,
	}

	resp, err := h.ctblr.GetCodebook(ctx, req)
	if err != nil{
		return fmt.Errorf("failed to get codebook from Cantabular: %w", err)
	}

	log.Info(ctx, "Successfully got Codebook", log.Data{"codebook": resp})

	ds, err := transform.CodebookToDataset(resp.Codebook)
	if err != nil{
		return fmt.Errorf("failed to transform codebook to dataset: %w", err)
	}

	fmt.Println(ds)

	/*
	if err := h.datasets.PutDataset(ctx, "", h.authToken, e.CollectionID, e.DatablobName, *ds); err != nil{
		return fmt.Errorf("failed to save dataset: %w", err)
	}*/

	if err := h.triggerImportDimensionOptions(); err != nil{
		return fmt.Errorf("failed to import dimension options: %w", err)
	}

	return nil
}

func (h *InstanceStarted) getInstanceFromRecipe(ctx context.Context, r *models.Recipe) (*models.Instance, error){
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

func (h *InstanceStarted) getCodeListsFromInstance(i *models.Instance) ([]string, error){
	if len(i.CodeLists) < 1{
		return nil, fmt.Errorf("no code-lists (dimensions) found in instance")
	}

	var codelists []string
	for _, cl := range i.CodeLists{
		codelists = append(codelists, cl.Name)
	}

	return codelists, nil
}

func (h *InstanceStarted) triggerImportDimensionOptions() error {
	// Kafka producer loop
	return nil
}