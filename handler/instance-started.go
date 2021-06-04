package handler

import (
	"context"
	"fmt"

	"github.com/ONSdigital/dp-import-cantabular-dataset/config"
	"github.com/ONSdigital/dp-import-cantabular-dataset/event"
	"github.com/ONSdigital/dp-import-cantabular-dataset/transform"
	"github.com/ONSdigital/dp-import-cantabular-dataset/cantabular"
)

// InstanceStarted ...
type InstanceStarted struct {
	ctblr      CantabularClient // *cantabular.Client
	datasets   DatasetAPIClient // *dataset.Client
	authToken string
	// Kafka Producer
}

// Handle takes a single event.
func (h *InstanceStarted) Handle(ctx context.Context, cfg *config.Config, e *event.InstanceStarted) error {
	req := cantabular.GetCodebookRequest{
		DatasetName: e.DatablobName,
		//Variables: e.Variables,
		Categories: true,
	}

	resp, err := h.ctblr.GetCodebook(req)
	if err != nil{
		return fmt.Errorf("failed to get codebook from Cantabular: %w", err)
	}

	ds, err := transform.CodebookToDataset(resp.Codebook)
	if err != nil{
		return fmt.Errorf("failed to transform codebook to dataset: %w", err)
	}

	if err := h.datasets.PutDataset(ctx, "", h.authToken, e.CollectionID, e.DatablobName, *ds); err != nil{
		return fmt.Errorf("failed to save dataset: %w", err)
	}

	if err := h.triggerImportDimensionOptions(); err != nil{
		return fmt.Errorf("failed to import dimension options: %w", err)
	}

	return nil
}

func (h *InstanceStarted) triggerImportDimensionOptions() error {
	// Kafka producer loop
	return nil
}