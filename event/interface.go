package event

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
)

//go:generate moq -out mock/handler.go -pkg mock . Handler
//go:generate moq -out mock/import-api-client.go -pkg mock . ImportAPIClient
//go:generate moq -out mock/dataset-api-client.go -pkg mock . DatasetAPIClient

// Handler represents a handler for processing a single event.
type Handler interface {
	Handle(context.Context, *InstanceStarted) error
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
	UpdateImportJobState(context.Context, string, string, string) error
}

type DatasetAPIClient interface {
	PutInstanceState(context.Context, string, string, dataset.State) error
}
