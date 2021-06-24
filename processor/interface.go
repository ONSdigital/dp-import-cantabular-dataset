package processor

import (
	"context"

	"github.com/ONSdigital/dp-import-cantabular-dataset/event"
	"github.com/ONSdigital/dp-api-clients-go/dataset"
)

//go:generate moq -out mock/handler.go -pkg mock . Handler

// Handler represents a handler for processing a single event.
type Handler interface {
	Handle(context.Context, *event.InstanceStarted) error
}

type coder interface{
	Code() int
}

type dataLogger interface{
	LogData() map[string]interface{}
}

type ImportAPIClient interface{
	UpdateImportJobState(context.Context, string, string, string) error
}

type DatasetAPIClient interface{
	PutInstanceState(context.Context, string, string, dataset.State) error
}