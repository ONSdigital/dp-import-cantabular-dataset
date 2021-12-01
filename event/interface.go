package event

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/importapi"
)

//go:generate moq -out mock/handler.go -pkg mock . Handler
//go:generate moq -out mock/import_api_client.go -pkg mock . ImportAPIClient
//go:generate moq -out mock/dataset_api_client.go -pkg mock . DatasetAPIClient

// Handler represents a handler for processing a single event.
type Handler interface {
	Handle(ctx context.Context, instanceStarted *InstanceStarted) error
}

type coder interface {
	Code() int
}

type dataLogger interface {
	LogData() map[string]interface{}
}

type instanceCompleteder interface {
	InstanceCompleted() bool
}

type ImportAPIClient interface {
	UpdateImportJobState(ctx context.Context, jobID, serviceToken string, newState importapi.State) error
}

type DatasetAPIClient interface {
	PutInstanceState(ctx context.Context, serviceAuthToken, instanceID string, state dataset.State, ifMatch string) (eTag string, err error)
}
