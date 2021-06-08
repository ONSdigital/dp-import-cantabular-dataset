package cantabular_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/ONSdigital/dp-import-cantabular-dataset/cantabular"
	dphttp "github.com/ONSdigital/dp-net/http"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetCodebook(t *testing.T) {

	Convey("Given a 404 response from the /Codebook endpoint", t, func() {
		testCtx := context.Background()

		mockHttpClient := &dphttp.ClienterMock{
			GetFunc: func(ctx context.Context, url string) (*http.Response, error) {
				return Response(make([]byte, 0), 404), nil
			},
		}

		cantabularClient := cantabular.NewClient(
			mockHttpClient,
			cantabular.Config{},
		)

		Convey("When the GetCodebook method is called", func() {
			req := cantabular.GetCodebookRequest{}
			cb, err := cantabularClient.GetCodebook(testCtx, req)

			Convey("Then the expected status code should be recoverable from the error", func() {
				So(cb, ShouldBeNil)
				So(cantabular.StatusCode(err), ShouldEqual, http.StatusNotFound)
			})
		})
	})
}

func Response(body []byte, statusCode int) *http.Response {
	reader := bytes.NewBuffer(body)
	readCloser := ioutil.NopCloser(reader)

	return &http.Response{
		StatusCode: statusCode,
		Body:       readCloser,
	}
}