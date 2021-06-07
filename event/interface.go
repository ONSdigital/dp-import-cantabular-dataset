package event

import (
	"context"
	"errors"

	"github.com/ONSdigital/dp-import-cantabular-dataset/config"
)

//go:generate moq -out mock/handler.go -pkg mock . Handler

// Handler represents a handler for processing a single event.
type Handler interface {
	Handle(context.Context, *config.Config, *InstanceStarted) error
}

type coder interface{
	Code() int
}

type dataLogger interface{
	LogData() map[string]interface{}
}

func statusCode(err error) int{
	var cerr coder
	if errors.As(err, &cerr){
		return cerr.Code()
	}

	return 0
}

func logData(err error) map[string]interface{}{
	var lderr dataLogger
	if errors.As(err, &lderr){
		return lderr.LogData()
	}

	return nil
}