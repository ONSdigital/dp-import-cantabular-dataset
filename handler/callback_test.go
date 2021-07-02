package handler

import (
	"testing"

	"github.com/ONSdigital/log.go/v2/log"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCallbackHappy(t *testing.T) {

	Convey("Given an error with embedded logData", t, func() {
		err := &Error{
			logData: log.Data{
				"log": "data",
			},
		}

		Convey("When logData(err) is called", func() {
			ld := logData(err)
			So(ld, ShouldResemble, log.Data{"log": "data"})
		})
	})
}
