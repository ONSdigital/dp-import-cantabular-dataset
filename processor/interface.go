package processor

import (
	"context"

	"github.com/ONSdigital/dp-import-cantabular-dataset/event"
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
