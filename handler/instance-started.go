package handler

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/ONSdigital/dp-import-cantabular-dataset/config"
	"github.com/ONSdigital/dp-import-cantabular-dataset/event"
	"github.com/ONSdigital/log.go/log"
)

// InstanceStarted ...
type InstanceStarted struct {}

// Handle takes a single event.
func (h *InstanceStarted) Handle(ctx context.Context, cfg *config.Config, e *event.InstanceStarted) error {
	logData := log.Data{"event": e}

	log.Event(ctx, "event handler called", log.INFO, logData)

	greeting := fmt.Sprintf("Hello, %s!", e.RecipientName)

	if err := ioutil.WriteFile(cfg.OutputFilePath, []byte(greeting), 0644); err != nil {
		return err
	}

	logData["greeting"] = greeting
	log.Event(ctx, "hello world example handler called successfully", log.INFO, logData)
	log.Event(ctx, "event successfully handled", log.INFO, logData)

	return nil
}
