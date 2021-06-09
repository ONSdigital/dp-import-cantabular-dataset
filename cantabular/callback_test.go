package cantabular
import (
	"net/http"
	"errors"
	"fmt"
	"testing"

	"github.com/ONSdigital/log.go/v2/log"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCallbackHappy(t *testing.T) {

	Convey("Given an error with embedded status code", t, func() {
		err := &Error{
			statusCode: http.StatusBadRequest,
		}

		Convey("When StatusCode(err) is called", func() {
			statusCode := StatusCode(err)
			So(statusCode, ShouldEqual, http.StatusBadRequest)
		})
	})

	Convey("Given an error with embedded logData", t, func() {
		err := &Error{
			logData: log.Data{
				"log":"data",
			},
		}

		Convey("When LogData(err) is called", func() {
			logData := LogData(err)
			So(logData, ShouldResemble, log.Data{"log":"data"})
		})
	})

	Convey("Given an error chain with wrapped logData", t, func() {
		err1 := &Error{
			err: errors.New("original error"),
			logData: log.Data{
				"log":"data",
			},
		}

		err2 := &Error{
			err: fmt.Errorf("err1: %w", err1),
			logData: log.Data{
				"additional": "data",
			},
		}

		err3 := &Error{
			err: fmt.Errorf("err2: %w", err2),
			logData: log.Data{
				"final": "data",
			},
		}


		Convey("When UnwrapLogData(err) is called", func() {
			logData := UnwrapLogData(err3)
			expected := []log.Data{
				log.Data{"final":"data"},
				log.Data{"additional":"data"},
				log.Data{"log":"data"},
			}
			
			So(logData, ShouldResemble,expected)

		})
	})
}