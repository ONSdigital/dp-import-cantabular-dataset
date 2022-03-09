package handler

import (
	"context"
	"errors"
	"fmt"

	"github.com/ONSdigital/dp-import-cantabular-dataset/config"
	"github.com/ONSdigital/dp-import-cantabular-dataset/event"
	"github.com/ONSdigital/dp-import-cantabular-dataset/schema"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular/gql"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	"github.com/ONSdigital/dp-api-clients-go/v2/importapi"
	"github.com/ONSdigital/dp-api-clients-go/v2/recipe"
	kafka "github.com/ONSdigital/dp-kafka/v3"

	"github.com/ONSdigital/log.go/v2/log"
)

const (
	cantabularTable = "cantabular_table"
)

// InstanceStarted is the handler for the InstanceStarted event
type InstanceStarted struct {
	cfg       config.Config
	ctblr     CantabularClient
	recipes   RecipeAPIClient
	importAPI ImportAPIClient
	datasets  DatasetAPIClient
	producer  kafka.IProducer
}

func NewInstanceStarted(cfg config.Config, c CantabularClient, r RecipeAPIClient, i ImportAPIClient, d DatasetAPIClient, p kafka.IProducer) *InstanceStarted {
	return &InstanceStarted{
		cfg:       cfg,
		ctblr:     c,
		recipes:   r,
		importAPI: i,
		datasets:  d,
		producer:  p,
	}
}

// Handle takes a single event.
func (h *InstanceStarted) Handle(ctx context.Context, workerID int, msg kafka.Message) error {
	e := &event.InstanceStarted{}
	s := schema.InstanceStarted

	if err := s.Unmarshal(msg.GetData(), e); err != nil {
		return h.handleError(ctx, e, &Error{
			err: fmt.Errorf("failed to unmarshal event: %w", err),
			logData: map[string]interface{}{
				"msg_data": msg.GetData(),
			},
		})
	}

	ld := log.Data{"event": e}
	log.Info(ctx, "event received", ld)

	r, err := h.recipes.GetRecipe(ctx, "", h.cfg.ServiceAuthToken, e.RecipeID)
	if err != nil {
		return h.handleError(ctx, e, &Error{
			err:     fmt.Errorf("failed to get recipe: %w", err),
			logData: ld,
		})
	}

	log.Info(ctx, "Successfully got Recipe", log.Data{"recipe_alias": r.Alias})

	i, err := h.getInstanceFromRecipe(ctx, r)
	if err != nil {
		return h.handleError(ctx, e, &Error{
			err:     fmt.Errorf("failed to get instance from recipe: %s", err),
			logData: ld,
		})
	}

	log.Info(ctx, "Successfully got instance", log.Data{"instance_title": i.Title})

	codelists, err := h.getCodeListsFromInstance(i)
	if err != nil {
		return h.handleError(ctx, e, &Error{
			err:     fmt.Errorf("failed to get code-lists (dimensions) from recipe instance: %s", err),
			logData: ld,
		})
	}

	log.Info(ctx, "Successfully got codelists", log.Data{"num_codelists": len(codelists)})

	req := cantabular.GetDimensionsByNameRequest{
		Dataset:        r.CantabularBlob,
		DimensionNames: codelists,
	}

	// Validation happens here, if any variables are incorrect, will throw an error
	resp, err := h.ctblr.GetDimensionsByName(ctx, req)
	if err != nil {
		return h.handleError(ctx, e, &Error{
			err:     fmt.Errorf("failed to get codebook from Cantabular: %w", err),
			logData: ld,
		})
	}

	if len(resp.Dataset.Variables.Edges) == 0 {
		return h.handleError(ctx, e, &Error{
			err:     fmt.Errorf("failed to get codebook from Cantabular: %w", err),
			logData: ld,
		})
	}

	log.Info(ctx, "Successfully got Codebook", log.Data{
		"datablob":      resp.Dataset,
		"num_variables": len(resp.Dataset.Variables.Edges),
	})

	if len(i.Editions) == 0 {
		return errors.New("no editions found in instance")
	}
	edition := i.Editions[0] // for census, we assume there will only ever be one

	ireq := h.CreateUpdateInstanceRequest(ctx, resp.Dataset.Variables, e, r, i.CodeLists, edition)

	log.Info(ctx, "Updating instance", log.Data{
		"instance_id":    ireq.InstanceID,
		"csv_headers":    ireq.CSVHeader,
		"edition":        ireq.Edition,
		"num_dimensions": len(ireq.Dimensions),
	})

	if _, err := h.datasets.PutInstance(ctx, "", h.cfg.ServiceAuthToken, "", e.InstanceID, ireq, headers.IfMatchAnyETag); err != nil {
		return h.handleError(ctx, e, &Error{
			err:     fmt.Errorf("failed to update instance: %w", err),
			logData: ld,
		})
	}

	if _, err := h.datasets.PutInstanceState(ctx, h.cfg.ServiceAuthToken, e.InstanceID, dataset.StateCompleted, headers.IfMatchAnyETag); err != nil {
		return h.handleError(ctx, e, &Error{
			err:     fmt.Errorf("failed to update instance state: %w", err),
			logData: ld,
		})
	}

	log.Info(ctx, "Triggering dimension options import")

	if errs := h.TriggerImportDimensionOptions(r, i.CodeLists, e); len(errs) != 0 {
		var errdata []map[string]interface{}

		for _, err := range errs {
			errdata = append(errdata, map[string]interface{}{
				"error":    err.Error(),
				"log_data": logData(err),
			})
		}

		ld["errors"] = errdata

		return h.handleError(ctx, e, &Error{
			err:               errors.New("failed to successfully trigger options import for all dimensions"),
			logData:           ld,
			instanceCompleted: true,
		})
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

func (h *InstanceStarted) CreateUpdateInstanceRequest(ctx context.Context, mf gql.Variables, e *event.InstanceStarted, r *recipe.Recipe, codelists []recipe.CodeList, edition string) dataset.UpdateInstance {

	req := dataset.UpdateInstance{
		Edition:    edition,
		CSVHeader:  []string{cantabularTable},
		InstanceID: e.InstanceID,
		Type:       r.Format,
		IsBasedOn: &dataset.IsBasedOn{
			ID:   r.CantabularBlob,
			Type: r.Format,
		},
	}

	for i, edge := range mf.Edges {
		sourceName := edge.Node.Name
		if sourceName == "" {
			log.Warn(ctx, "ignoring empty name for node", log.Data{"node": edge.Node})
			continue
		}

		if r.Format == "cantabular_flexible_table" {
			if codelists[i].IsCantabularGeography != nil {
				if *codelists[i].IsCantabularGeography {
					continue
				}
			}
		}

		id := sourceName
		url := fmt.Sprintf("%s/code-lists/%s", h.cfg.RecipeAPIURL, sourceName)

		// id and url values overwritten by codelist values.
		// Note that cantabular codebook is sorted with exactly the same order as codelists array.
		if codelists[i].ID != "" {
			id = codelists[i].ID
		}
		if codelists[i].ID != "" {
			url = codelists[i].HRef
		}

		d := dataset.VersionDimension{
			ID:              id,
			URL:             url,
			Label:           edge.Node.Label,
			Name:            edge.Node.Label,
			Variable:        sourceName,
			NumberOfOptions: edge.Node.Categories.TotalCount,
		}
		req.Dimensions = append(req.Dimensions, d)
		req.CSVHeader = append(req.CSVHeader, edge.Node.Name)
	}

	return req
}

func (h *InstanceStarted) TriggerImportDimensionOptions(r *recipe.Recipe, codelists []recipe.CodeList, e *event.InstanceStarted) []error {
	var errs []error
	var geographyCount int
	tc := len(codelists)

	for _, cl := range codelists {
		ie := event.CategoryDimensionImport{
			DimensionID:    cl.ID,
			JobID:          e.JobID,
			InstanceID:     e.InstanceID,
			CantabularBlob: r.CantabularBlob,
		}

		if r.Format == "cantabular_flexible_table" {
			if cl.IsCantabularGeography != nil {
				if *cl.IsCantabularGeography {
					geographyCount++
					// To indicate `dp-import-cantabular-dimension-options` when consuming the kafka message to not update the instance
					// The message still needs to be consumed to determine that all dimensions have been processed
					// so that it can mark the job as finished and complete
					ie.IsGeography = true
				}
			}
		}

		s := schema.CategoryDimensionImport

		b, err := s.Marshal(ie)
		if err != nil {
			errs = append(errs, &Error{
				err: fmt.Errorf("avro: failed to marshal dimension: %w", err),
				logData: log.Data{
					"dimension_id": cl.ID,
				},
			})
			continue
		}

		h.producer.Channels().Output <- b
	}
	if geographyCount == tc {
		var errs []error
		errs = append(errs, &Error{
			err:     fmt.Errorf("only geography codelists exist in this instance, there must be at least one non-geography in the codelists"),
			logData: log.Data{},
		})
		return errs
	}

	return errs
}

// handleError updates the import job and instance to failed state, after an error during Handle
func (h *InstanceStarted) handleError(ctx context.Context, e *event.InstanceStarted, err *Error) error {
	var errs []error

	if err := h.importAPI.UpdateImportJobState(ctx, e.JobID, h.cfg.ServiceAuthToken, importapi.StateFailed); err != nil {
		errs = append(errs, &Error{
			err: fmt.Errorf("failed to update job state: %w", err),
			logData: log.Data{
				"job_id": e.JobID,
			},
		})
	}

	if !instanceCompleted(err) {
		if _, err := h.datasets.PutInstanceState(ctx, h.cfg.ServiceAuthToken, e.InstanceID, dataset.StateFailed, headers.IfMatchAnyETag); err != nil {
			errs = append(errs, &Error{
				err: fmt.Errorf("failed to update instance state: %w", err),
				logData: log.Data{
					"instance_id": e.InstanceID,
					"job_id":      e.JobID,
				},
			})
		}
	}

	if len(errs) > 0 {
		err.logData["errors"] = errs
	}

	return err
}
