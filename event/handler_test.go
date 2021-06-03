package event_test

import (
	"os"
	"testing"

	"github.com/ONSdigital/dp-import-cantabular-dataset/config"
	"github.com/ONSdigital/dp-import-cantabular-dataset/event"
	. "github.com/smartystreets/goconvey/convey"
)

func TestInstanceStartedHandler_Handle(t *testing.T) {
	Convey("Given a successful event handler, when Handle is triggered", t, func() {
		eventHandler := &event.InstanceStartedHandler{}
		filePath := "/tmp/helloworld.txt"
		os.Remove(filePath)
		err := eventHandler.Handle(testCtx, &config.Config{OutputFilePath: filePath}, &testEvent)
		So(err, ShouldBeNil)
	})

	Convey("handler returns an error when cannot write to file", t, func() {
		eventHandler := &event.InstanceStartedHandler{}
		filePath := ""
		err := eventHandler.Handle(testCtx, &config.Config{OutputFilePath: filePath}, &testEvent)
		So(err, ShouldNotBeNil)
	})
}
