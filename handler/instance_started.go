package handler

import (
	"context"
	"errors"
	"fmt"

	"github.com/ONSdigital/dp-import-cantabular-dataset/config"
	"github.com/ONSdigital/dp-import-cantabular-dataset/event"
	"github.com/ONSdigital/dp-import-cantabular-dataset/schema"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/recipe"
	kafka "github.com/ONSdigital/dp-kafka/v2"

	"github.com/ONSdigital/log.go/v2/log"
)

const (
	cantabularTable = "cantabular_table"
)

// InstanceStarted is the handler for the InstanceStarted event
type InstanceStarted struct {
	cfg      config.Config
	ctblr    CantabularClient
	datasets DatasetAPIClient
	recipes  RecipeAPIClient
	producer kafka.IProducer
}

func NewInstanceStarted(cfg config.Config, c CantabularClient, r RecipeAPIClient, d DatasetAPIClient, p kafka.IProducer) *InstanceStarted {
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
	ld := log.Data{
		"job_id":      e.JobID,
		"instance_id": e.InstanceID,
	}

	r, err := h.recipes.GetRecipe(ctx, "", h.cfg.ServiceAuthToken, e.RecipeID)
	if err != nil {
		return &Error{
			err:     fmt.Errorf("failed to get recipe: %w", err),
			logData: ld,
		}
	}

	log.Info(ctx, "Successfully got Recipe", log.Data{"recipe_alias": r.Alias})

	i, err := h.getInstanceFromRecipe(ctx, r)
	if err != nil {
		return &Error{
			err:     fmt.Errorf("failed to get instance from recipe: %s", err),
			logData: ld,
		}
	}

	log.Info(ctx, "Successfully got instance", log.Data{"instance_title": i.Title})

	codelists, err := h.getCodeListsFromInstance(i)
	if err != nil {
		return &Error{
			err:     fmt.Errorf("failed to get code-lists (dimensions) from recipe instance: %s", err),
			logData: ld,
		}
	}

	log.Info(ctx, "Successfully got codelists", log.Data{"num_codelists": len(codelists)})

	req := cantabular.GetCodebookRequest{
		DatasetName: r.CantabularBlob,
		Variables:   codelists,
		Categories:  false,
	}

	// Validation happens here, if any variables are incorrect, will throw an error
	resp, err := h.ctblr.GetCodebook(ctx, req)
	if err != nil {
		return &Error{
			err:     fmt.Errorf("failed to get codebook from Cantabular: %w", err),
			logData: ld,
		}
	}

	log.Info(ctx, "Successfully got Codebook", log.Data{
		"datablob":      resp.Dataset,
		"num_variables": len(resp.Codebook),
		"codebook":      resp.Codebook,
	})

	ireq := h.createUpdateInstanceRequest(resp.Codebook, e, r.CantabularBlob)

	log.Info(ctx, "Updating instance", log.Data{
		"instance_id":    ireq.InstanceID,
		"csv_headers":    ireq.CSVHeader,
		"edition":        ireq.Edition,
		"num_dimensions": len(ireq.Dimensions),
	})

	if _, err := h.datasets.PutInstance(ctx, "", h.cfg.ServiceAuthToken, "", e.InstanceID, ireq, headers.IfMatchAnyETag); err != nil {
		return &Error{
			err:     fmt.Errorf("failed to update instance: %w", err),
			logData: ld,
		}
	}

	if _, err := h.datasets.PutInstanceState(ctx, h.cfg.ServiceAuthToken, e.InstanceID, dataset.StateCompleted, headers.IfMatchAnyETag); err != nil {
		return &Error{
			err:     fmt.Errorf("failed to update instance state: %w", err),
			logData: ld,
		}
	}

	log.Info(ctx, "Triggering dimension options import", log.Data{
		"num_dimensions": len(codelists),
	})

	if errs := h.triggerImportDimensionOptions(ctx, r.CantabularBlob, codelists, e); len(errs) != 0 {
		var errdata []map[string]interface{}

		for _, err := range errs {
			errdata = append(errdata, map[string]interface{}{
				"error":    err.Error(),
				"log_data": logData(err),
			})
		}

		ld["errors"] = errdata

		return &Error{
			err:               errors.New("failed to successfully trigger options import for all dimensions"),
			logData:           ld,
			instanceCompleted: true,
		}
	}

	log.Info(ctx, "Successfully triggered options import for all dimensions")

	return nil
}

func (h *InstanceStarted) getInstanceFromRecipe(ctx context.Context, r *recipe.Recipe) (*recipe.Instance, error) {
	if len(r.OutputInstances) < 1 {
		return nil, errors.New("no instances found in recipe")
	}

	if len(r.OutputInstances) > 1 {
		log.Warn(ctx, "more than one instance found in recipe. Defaulting to instances[0]", log.Data{
			"num_instances": len(r.OutputInstances),
		})
	}

	return &r.OutputInstances[0], nil
}

func (h *InstanceStarted) getCodeListsFromInstance(i *recipe.Instance) ([]string, error) {
	if len(i.CodeLists) < 1 {
		return nil, fmt.Errorf("no code-lists (dimensions) found in instance")
	}

	var codelists []string
	for _, cl := range i.CodeLists {
		codelists = append(codelists, cl.ID)
	}

	return codelists, nil
}

func (h *InstanceStarted) createUpdateInstanceRequest(cb cantabular.Codebook, e *event.InstanceStarted, ctblrBlob string) dataset.UpdateInstance {
	req := dataset.UpdateInstance{
		Edition:    "2021",
		CSVHeader:  []string{cantabularTable},
		InstanceID: e.InstanceID,
		Type:       cantabularTable,
		IsBasedOn:  &dataset.IsBasedOn{
			ID:   ctblrBlob,
			Type: cantabularTable,
		},
	}

	for _, v := range cb {
		sourceName := v.Name

		if len(v.MapFrom) > 0 {
			if len(v.MapFrom[0].SourceNames) > 0 {
				sourceName = v.MapFrom[0].SourceNames[0]
			}
		}

		d := dataset.VersionDimension{
			ID:              sourceName,
			URL:             fmt.Sprintf("%s/code-lists/%s", h.cfg.RecipeAPIURL, sourceName),
			Label:           v.Label,
			Name:            v.Label,
			Variable:        sourceName,
			NumberOfOptions: v.Len,
		}
		req.Dimensions = append(req.Dimensions, d)
		req.CSVHeader = append(req.CSVHeader, v.Name)
	}

	return req
}

func (h *InstanceStarted) triggerImportDimensionOptions(ctx context.Context, blob string, dimensions []string, e *event.InstanceStarted) []error {
	var errs []error

	for _, d := range dimensions {
		ie := event.CategoryDimensionImport{
			DimensionID:    d,
			JobID:          e.JobID,
			InstanceID:     e.InstanceID,
			CantabularBlob: blob,
		}

		s := schema.CategoryDimensionImport

		b, err := s.Marshal(ie)
		if err != nil {
			errs = append(errs, &Error{
				err: fmt.Errorf("avro: failed to marshal dimension: %w", err),
				logData: log.Data{
					"dimension_id": d,
				},
			})
			continue
		}

		h.producer.Channels().Output <- b
	}

	return errs
}
