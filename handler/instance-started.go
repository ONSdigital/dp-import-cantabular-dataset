package handler

import (
	"context"
	"fmt"

	"github.com/ONSdigital/dp-import-cantabular-dataset/config"
	"github.com/ONSdigital/dp-import-cantabular-dataset/event"
	// "github.com/ONSdigital/dp-import-cantabular-dataset/transform"
	"github.com/ONSdigital/dp-import-cantabular-dataset/cantabular"
)

// InstanceStarted ...
type InstanceStarted struct {
	ctblr      CantabularClient // *cantabular.Client
	datasets   DatasetAPIClient // *dp-api-clients-go/dataset.Client
	authToken string
	// Kafka Producer
}

func NewInstanceStarted(ctblrClient CantabularClient, datasetAPIClient DatasetAPIClient, authToken string) *InstanceStarted {
	return &InstanceStarted{
		ctblr: ctblrClient,
		datasets: datasetAPIClient,
		authToken: authToken,
	}
}

// Handle takes a single event.
func (h *InstanceStarted) Handle(ctx context.Context, cfg *config.Config, e *event.InstanceStarted) error {
	req := cantabular.GetCodebookRequest{
		DatasetName: e.DatablobName,
		Variables: []string{"city", "siblings"},//e.Variables,
		Categories: true,
	}

	resp, err := h.ctblr.GetCodebook(ctx, req)
	if err != nil{
		return fmt.Errorf("failed to get codebook from Cantabular: %w", err)
	}

	fmt.Println(resp)
	/*
	ds, err := transform.CodebookToDataset(resp.Codebook)
	if err != nil{
		return fmt.Errorf("failed to transform codebook to dataset: %w", err)
	}

	if err := h.datasets.PutDataset(ctx, "", h.authToken, e.CollectionID, e.DatablobName, *ds); err != nil{
		return fmt.Errorf("failed to save dataset: %w", err)
	}*/

	if err := h.triggerImportDimensionOptions(); err != nil{
		return fmt.Errorf("failed to import dimension options: %w", err)
	}

	return nil
}

func (h *InstanceStarted) triggerImportDimensionOptions() error {
	// Kafka producer loop
	return nil
}