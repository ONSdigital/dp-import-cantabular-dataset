package handler_test
/*
import (
	"os"
	"testing"
	"context"

	"github.com/ONSdigital/dp-import-cantabular-dataset/config"
	"github.com/ONSdigital/dp-import-cantabular-dataset/handler"
	"github.com/ONSdigital/dp-import-cantabular-dataset/event"

	. "github.com/smartystreets/goconvey/convey"
)
var (
	testEvent = event.InstanceStarted{
		RecipientName: "World",
	}
)

func TestInstanceStartedHandler_Handle(t *testing.T) {
	Convey("Given a successful event handler, when Handle is triggered", t, func() {
		ctx := context.Background()
		h := &handler.InstanceStarted{}

		err := h.Handle(ctx, &config.Config{OutputFilePath: filePath}, &testEvent)
		So(err, ShouldBeNil)
	})

	Convey("handler returns an error when cannot write to file", t, func() {
		ctx := context.Background()
		h := &handler.InstanceStarted{}
		filePath := ""

		err := h.Handle(ctx, &config.Config{OutputFilePath: filePath}, &testEvent)
		So(err, ShouldNotBeNil)
	})
}
*/