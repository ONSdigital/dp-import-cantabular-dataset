package service

import (
	"context"
	"net/http"

	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-import-cantabular-dataset/processor"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-api-clients-go/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/recipe"
	"github.com/ONSdigital/dp-api-clients-go/dataset"

)

//go:generate moq -out mock/server.go -pkg mock . HTTPServer
//go:generate moq -out mock/health-check.go -pkg mock . HealthChecker
//go:generate moq -out mock/canabular-client.go -pkg mock . CantabularClient
//go:generate moq -out mock/dataset-api-client.go -pkg mock . DatasetAPIClient
//go:generate moq -out mock/recipe-api-client.go -pkg mock . RecipeAPIClient
//go:generate moq -out mock/processor.go -pkg mock . Processor

// HTTPServer defines the required methods from the HTTP server
type HTTPServer interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

// HealthChecker defines the required methods from Healthcheck
type HealthChecker interface {
	Handler(w http.ResponseWriter, req *http.Request)
	Start(ctx context.Context)
	Stop()
	AddCheck(name string, checker healthcheck.Checker) (err error)
}

type CantabularClient interface{
	GetCodebook(context.Context, cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error)
	Checker(context.Context, *healthcheck.CheckState) error
}

type DatasetAPIClient interface{
	PutInstance(context.Context, string, string, string, string, dataset.UpdateInstance) error
	PutInstanceState(context.Context, string, string, dataset.State) error
	Checker(context.Context, *healthcheck.CheckState) error
}

type RecipeAPIClient interface{
	GetRecipe(context.Context, string, string, string) (*recipe.Recipe, error)
	Checker(context.Context, *healthcheck.CheckState) error
}

type ImportAPIClient interface{
	UpdateImportJobState(context.Context, string, string, string) error
}

type Processor interface{
	Consume (context.Context, kafka.IConsumerGroup, processor.Handler)
}
