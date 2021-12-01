package handler_test

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/importapi"
	"github.com/ONSdigital/dp-api-clients-go/v2/recipe"

	"github.com/ONSdigital/dp-import-cantabular-dataset/config"
	"github.com/ONSdigital/dp-import-cantabular-dataset/event"
	"github.com/ONSdigital/dp-import-cantabular-dataset/handler"
	"github.com/ONSdigital/dp-import-cantabular-dataset/schema"

	"github.com/ONSdigital/dp-import-cantabular-dataset/handler/mock"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	workerID   = 1
	msgTimeout = time.Second
)

type testError struct {
	statusCode int
}

func (e *testError) Error() string {
	return "I am a test error"
}

func (e *testError) StatusCode() int {
	return e.statusCode
}

type statusCoder interface {
	StatusCode() int
}

func TestInstanceStartedHandler_HandleHappy(t *testing.T) {
	cfg := config.Config{}

	Convey("Given a successful event handler", t, func() {
		ctblrClient := cantabularClientHappy()
		recipeAPIClient := recipeAPIClientHappy()
		importAPIClient := importAPIClientHappy()
		datasetAPIClient := datasetAPIClientHappy()
		producer := kafkatest.NewMessageProducer(true)

		h := handler.NewInstanceStarted(
			cfg,
			ctblrClient,
			recipeAPIClient,
			importAPIClient,
			datasetAPIClient,
			producer,
		)

		ctx := context.Background()
		wg := &sync.WaitGroup{}

		Convey("When Handle is triggered", func(c C) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				msg := kafkaMessage(c, &event.InstanceStarted{
					RecipeID:       "test-recipe-id",
					InstanceID:     "test-instance-id",
					JobID:          "test-job-id",
					CantabularType: "cantabular_table",
				})
				err := h.Handle(ctx, workerID, msg)
				c.So(err, ShouldBeNil)
			}()

			Convey("Then the expected category dimension import events are produced", func() {
				expected := []event.CategoryDimensionImport{
					{
						JobID:          "test-job-id",
						InstanceID:     "test-instance-id",
						DimensionID:    "test-variable",
						CantabularBlob: "test-cantabular-blob",
					},
					{
						JobID:          "test-job-id",
						InstanceID:     "test-instance-id",
						DimensionID:    "test-mapped-variable",
						CantabularBlob: "test-cantabular-blob",
					},
				}

				for _, e := range expected {
					validateMessage(t, producer, c, e)
				}

				Convey("And the expected dimensions are updated in dataset api instance, with ID and Href values comming from the corresponding codelist", func() {
					So(datasetAPIClient.PutInstanceCalls(), ShouldHaveLength, 1)
					So(datasetAPIClient.PutInstanceCalls()[0].InstanceID, ShouldEqual, "test-instance-id")
					So(datasetAPIClient.PutInstanceCalls()[0].I, ShouldResemble, dataset.UpdateInstance{
						Edition:    "2021",
						CSVHeader:  []string{"cantabular_table", "test-variable", "test-mapped-variable"},
						InstanceID: "test-instance-id",
						Type:       "cantabular_table",
						IsBasedOn: &dataset.IsBasedOn{
							ID:   "test-cantabular-blob",
							Type: "cantabular_table",
						},
						Dimensions: []dataset.VersionDimension{
							{
								ID:              "test-variable",
								URL:             "http://recipe-defined-host/code-lists/test-variable",
								Label:           "Test Variable",
								Name:            "Test Variable",
								Variable:        "test-variable",
								NumberOfOptions: 3,
							},
							{
								ID:              "test-mapped-variable",
								URL:             "http://recipe-defined-host/code-lists/test-mapped-variable",
								Label:           "Test Mapped Variable",
								Name:            "Test Mapped Variable",
								Variable:        "test-variable",
								NumberOfOptions: 2,
							},
						},
					})
				})

				wg.Wait()
				err := producer.Close(ctx)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestInstanceStartedHandler_HandleUnhappy(t *testing.T) {
	cfg := config.Config{}
	ctx := context.Background()

	Convey("Given an event handler with a recipe that cannot be found", t, func() {
		ctblrClient := cantabularClientHappy()
		recipeAPIClient := recipeAPIClientHappy()
		importAPIClient := importAPIClientHappy()
		datasetAPIClient := datasetAPIClientHappy()
		producer := kafkatest.NewMessageProducer(true)

		recipeAPIClient.GetRecipeFunc = func(ctx context.Context, uaToken, saToken, recipeID string) (*recipe.Recipe, error) {
			return nil, &testError{http.StatusNotFound}
		}

		h := handler.NewInstanceStarted(
			cfg,
			ctblrClient,
			recipeAPIClient,
			importAPIClient,
			datasetAPIClient,
			producer,
		)

		Convey("When Handle is triggered", func(c C) {
			msg := kafkaMessage(c, &event.InstanceStarted{
				RecipeID:       "test-recipe-id",
				InstanceID:     "test-instance-id",
				JobID:          "test-job-id",
				CantabularType: "cantabular_table",
			})
			err := h.Handle(ctx, workerID, msg)

			Convey("Then the expected error is returned, with Status NotFound", func() {
				So(err, ShouldNotBeNil)
				var cerr statusCoder
				So(errors.As(err, &cerr), ShouldBeTrue)
				So(cerr.StatusCode(), ShouldEqual, http.StatusNotFound)
			})

			Convey("Then the import job and instance are set to failed state", func() {
				validateFailure(importAPIClient, datasetAPIClient, "test-job-id", "test-instance-id")
			})
		})
	})

	Convey("Given an event handler with a dataset api client that fails to update an instance", t, func() {
		ctblrClient := cantabularClientHappy()
		recipeAPIClient := recipeAPIClientHappy()
		importAPIClient := importAPIClientHappy()
		datasetAPIClient := datasetAPIClientUnhappy()
		producer := kafkatest.NewMessageProducer(true)

		h := handler.NewInstanceStarted(
			cfg,
			ctblrClient,
			recipeAPIClient,
			importAPIClient,
			datasetAPIClient,
			producer,
		)

		Convey("When Handle is triggered", func(c C) {
			msg := kafkaMessage(c, &event.InstanceStarted{
				RecipeID:       "test-recipe-id",
				InstanceID:     "test-instance-id",
				JobID:          "test-job-id",
				CantabularType: "cantabular_table",
			})
			err := h.Handle(ctx, workerID, msg)

			Convey("Then the expected error is returned, with Status InternalServerError", func() {
				So(err, ShouldNotBeNil)
				var cerr statusCoder
				So(errors.As(err, &cerr), ShouldBeTrue)
				So(cerr.StatusCode(), ShouldEqual, http.StatusInternalServerError)
			})

			Convey("Then the import job and instance are set to failed state", func() {
				validateFailure(importAPIClient, datasetAPIClient, "test-job-id", "test-instance-id")
			})
		})
	})
}

func testCodebook() cantabular.Codebook {
	return cantabular.Codebook{
		cantabular.Variable{
			VariableBase: cantabular.VariableBase{
				Name:  "test-variable",
				Label: "Test Variable",
			},
			Len: 3,
			Codes: []string{
				"code1",
				"code2",
				"code3",
			},
			Labels: []string{
				"Code 1",
				"Code 2",
				"Code 3",
			},
		},
		cantabular.Variable{
			VariableBase: cantabular.VariableBase{
				Name:  "test-mapped-variable",
				Label: "Test Mapped Variable",
			},
			Len: 2,
			Codes: []string{
				"code-values-1-2",
				"code-value-3",
			},
			Labels: []string{
				"Codes 1-2",
				"Code 3",
			},
			MapFrom: []cantabular.MapFrom{
				{
					SourceNames: []string{
						"test-variable",
					},
					Code: []string{
						"code-values 1-2",
						"",
						"code-value 3",
					},
				},
			},
		},
	}
}

func testRecipe() *recipe.Recipe {
	return &recipe.Recipe{
		ID:             "test-recipe-id",
		CantabularBlob: "test-cantabular-blob",
		OutputInstances: []recipe.Instance{
			{
				DatasetID: "test-dataset-id",
				Title:     "Test Instance",
				CodeLists: []recipe.CodeList{
					{
						ID:   "test-variable",
						HRef: "http://recipe-defined-host/code-lists/test-variable",
						Name: "Test Variable",
					},
					{
						ID:   "test-mapped-variable",
						HRef: "http://recipe-defined-host/code-lists/test-mapped-variable",
						Name: "Test Mapped Variable",
					},
				},
			},
		},
	}
}

func cantabularClientHappy() *mock.CantabularClientMock {
	return &mock.CantabularClientMock{
		GetCodebookFunc: func(ctx context.Context, req cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error) {
			return &cantabular.GetCodebookResponse{
				Codebook: testCodebook(),
			}, nil
		},
	}
}

func recipeAPIClientHappy() *mock.RecipeAPIClientMock {
	return &mock.RecipeAPIClientMock{
		GetRecipeFunc: func(ctx context.Context, uaToken, saToken, recipeID string) (*recipe.Recipe, error) {
			return testRecipe(), nil
		},
	}
}

func importAPIClientHappy() *mock.ImportAPIClientMock {
	return &mock.ImportAPIClientMock{
		UpdateImportJobStateFunc: func(ctx context.Context, uaToken string, jobID string, state importapi.State) error {
			return nil
		},
	}
}

func datasetAPIClientHappy() *mock.DatasetAPIClientMock {
	return &mock.DatasetAPIClientMock{
		PutInstanceFunc: func(ctx context.Context, uaToken, saToken, collectionID, instanceID string, i dataset.UpdateInstance, mi string) (string, error) {
			return "", nil
		},
		PutInstanceStateFunc: func(ctx context.Context, uaToken, instanceID string, s dataset.State, mi string) (string, error) {
			return "", nil
		},
	}
}

func datasetAPIClientUnhappy() *mock.DatasetAPIClientMock {
	return &mock.DatasetAPIClientMock{
		PutInstanceFunc: func(ctx context.Context, uaToken, saToken, collectionID, instanceID string, i dataset.UpdateInstance, mi string) (string, error) {
			return "", &testError{http.StatusInternalServerError}
		},
		PutInstanceStateFunc: func(ctx context.Context, uaToken, instanceID string, s dataset.State, mi string) (string, error) {
			return "", nil
		},
	}
}

// kafkaMessage creates a mocked kafka message with the provided event as data
func kafkaMessage(c C, e *event.InstanceStarted) *kafkatest.Message {
	b, err := schema.InstanceStarted.Marshal(e)
	c.So(err, ShouldBeNil)
	return kafkatest.NewMessage(b, 0)
}

// validateMessage waits on the producer output channel,
// un-marshals the received message with the CategoryDimensionImport schema
// and checks that it resembles the provided expected event.
// If no message is received after a timeout (const) then it forces a test failure
func validateMessage(t *testing.T, producer kafka.IProducer, c C, expected event.CategoryDimensionImport) {
	select {
	case b := <-producer.Channels().Output:
		var got event.CategoryDimensionImport
		s := schema.CategoryDimensionImport
		err := s.Unmarshal(b, &got)
		c.So(err, ShouldBeNil)
		c.So(got, ShouldResemble, expected)
	case <-time.After(msgTimeout):
		t.Fail()
	}
}

func validateFailure(i *mock.ImportAPIClientMock, d *mock.DatasetAPIClientMock, jobID, instanceID string) {
	So(i.UpdateImportJobStateCalls(), ShouldHaveLength, 1)
	So(i.UpdateImportJobStateCalls()[0].JobID, ShouldEqual, jobID)
	So(i.UpdateImportJobStateCalls()[0].NewState, ShouldEqual, importapi.StateFailed)
	So(d.PutInstanceStateCalls(), ShouldHaveLength, 1)
	So(d.PutInstanceStateCalls()[0].InstanceID, ShouldEqual, instanceID)
	So(d.PutInstanceStateCalls()[0].State, ShouldEqual, dataset.StateFailed)
}
