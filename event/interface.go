package event

import (
	"context"

	"github.com/ONSdigital/dp-import-cantabular-dataset/config"
)

//go:generate moq -out mock/handler.go -pkg mock . Handler

// Handler represents a handler for processing a single event.
type Handler interface {
	Handle(context.Context, *config.Config, *InstanceStarted) error
}