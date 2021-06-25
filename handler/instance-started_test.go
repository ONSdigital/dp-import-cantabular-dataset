package handler_test

import (
	"testing"
	"context"
	"time"
	"net/http"
	"sync"
	"errors"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/recipe"
	"github.com/ONSdigital/dp-api-clients-go/cantabular"

	"github.com/ONSdigital/dp-import-cantabular-dataset/config"
	"github.com/ONSdigital/dp-import-cantabular-dataset/handler"
	"github.com/ONSdigital/dp-import-cantabular-dataset/event"
	"github.com/ONSdigital/dp-import-cantabular-dataset/schema"

	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-import-cantabular-dataset/handler/mock"

	. "github.com/smartystreets/goconvey/convey"
)

type testError struct{
	statusCode int
}

func (e *testError) Error() string {
	return "I am a test error"
}

func (e *testError) StatusCode() int {
	return e.statusCode
}

type statusCoder interface{
	StatusCode() int
}

func TestInstanceStartedHandler_HandleHappy(t *testing.T) {
	cfg := config.Config{}

	ctblrClient := cantabularClientHappy()
	recipeAPIClient := recipeAPIClientHappy()
	datasetAPIClient := datasetAPIClientHappy()
	producer := kafkatest.NewMessageProducer(true)

	h := handler.NewInstanceStarted(
		cfg,
		&ctblrClient,
		&recipeAPIClient,
		&datasetAPIClient,
		producer,
	)
	ctx := context.Background()

	Convey("Given a successful event handler, when Handle is triggered", t, func(c C) {
		wg := &sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer wg.Done()

			expected := []event.CategoryDimensionImport{
				event.CategoryDimensionImport{
					JobID:       "test-job-id",
					InstanceID:  "test-instance-id",
					DimensionID: "test-variable",
				},
				event.CategoryDimensionImport{
					JobID:       "test-job-id",
					InstanceID:  "test-instance-id",
					DimensionID: "test-mapped-variable",
				},
			}

			for _, e := range expected{
				validateMessage(t, producer, c, e)
			}
		}()

		err := h.Handle(ctx, &event.InstanceStarted{
			RecipeID: "test-recipe-id",
			InstanceID: "test-instance-id",
			JobID: "test-job-id",
			CantabularType: "cantabular-table",
		})
		So(err, ShouldBeNil)


		err = producer.Close(ctx)
		So(err, ShouldBeNil)
	})
}

func validateMessage(t *testing.T, producer kafka.IProducer, c C, expected event.CategoryDimensionImport) {
	timeout := time.Second * 1

	select {
	case b := <-producer.Channels().Output:
		var got event.CategoryDimensionImport
		s := schema.CategoryDimensionImport
		err := s.Unmarshal(b, &got)
		c.So(err, ShouldBeNil)
		c.So(got, ShouldResemble, expected)
	case <-time.After(timeout):
		t.Fail()
	}
}

func TestInstanceStartedHandler_HandleUnhappy(t *testing.T) {
	cfg := config.Config{}

	ctblrClient := cantabularClientHappy()
	recipeAPIClient := recipeAPIClientHappy()
	datasetAPIClient := datasetAPIClientHappy()
	producer := kafkatest.NewMessageProducer(true)

	ctx := context.Background()

	Convey("Given an event handler, when Handle is triggered with but recipe cannot be found", t, func() {
		recipeAPIClient.GetRecipeFunc = func(ctx context.Context, uaToken, saToken, recipeID string)(*recipe.Recipe, error){
				return nil, &testError{http.StatusNotFound}
		}

		h := handler.NewInstanceStarted(
			cfg,
			&ctblrClient,
			&recipeAPIClient,
			&datasetAPIClient,
			producer,
		)

		err := h.Handle(ctx, &event.InstanceStarted{
			RecipeID: "test-recipe-id",
			InstanceID: "test-instance-id",
			JobID: "test-job-id",
			CantabularType: "cantabular-table",
		})
		So(err, ShouldNotBeNil)
		var cerr statusCoder

		So(errors.As(err, &cerr), ShouldBeTrue)
		So(cerr.StatusCode(), ShouldEqual, http.StatusNotFound)
	})

	recipeAPIClient = recipeAPIClientHappy()

	Convey("Given an event handler, when Handle is triggered with but attempting to update instance fails", t, func() {
		datasetAPIClient = datasetAPIClientUnhappy()

		h := handler.NewInstanceStarted(
			cfg,
			&ctblrClient,
			&recipeAPIClient,
			&datasetAPIClient,
			producer,
		)

		err := h.Handle(ctx, &event.InstanceStarted{
			RecipeID: "test-recipe-id",
			InstanceID: "test-instance-id",
			JobID: "test-job-id",
			CantabularType: "cantabular-table",
		})
		So(err, ShouldNotBeNil)
		var cerr statusCoder

		So(errors.As(err, &cerr), ShouldBeTrue)
		So(cerr.StatusCode(), ShouldEqual, http.StatusInternalServerError)
	})
}

func testCodebook() cantabular.Codebook{
	return cantabular.Codebook{
		cantabular.Variable{
			Name: "test-variable",
			Label: "Test Variable",
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
			Name: "test-mapped-variable",
			Label: "Test Mapped Variable",
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
				cantabular.MapFrom{
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

func testRecipe() *recipe.Recipe{
	return &recipe.Recipe{
		ID: "test-recipe-id",
		OutputInstances: []recipe.Instance{
			recipe.Instance{
				DatasetID: "test-dataset-id",
				Title: "Test Instance",
				CodeLists: []recipe.CodeList{
					recipe.CodeList{
						ID: "test-variable",
						Name: "Test Variable",
					},
					recipe.CodeList{
						ID: "test-mapped-variable",
						Name: "Test Mapped Variable",
					},
				},
			},
		},
	}
}

func cantabularClientHappy() mock.CantabularClientMock{
	return mock.CantabularClientMock{
		GetCodebookFunc: func(ctx context.Context, req cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error){
			return &cantabular.GetCodebookResponse{
				Codebook: testCodebook(),
			}, nil
		},
	}
}

func recipeAPIClientHappy() mock.RecipeAPIClientMock{
	return mock.RecipeAPIClientMock{
		GetRecipeFunc: func(ctx context.Context, uaToken, saToken, recipeID string) (*recipe.Recipe, error){
			return testRecipe(), nil
		},
	}
}

func datasetAPIClientHappy() mock.DatasetAPIClientMock{
	return mock.DatasetAPIClientMock{
		PutInstanceFunc: func(ctx context.Context, uaToken, saToken, collectionID, instanceID string, i dataset.UpdateInstance) error {
			return nil
		},
		PutInstanceStateFunc: func(ctx context.Context, uaToken, instanceID string, s dataset.State) error {
			return nil
		},
	}
}

func datasetAPIClientUnhappy() mock.DatasetAPIClientMock{
	return mock.DatasetAPIClientMock{
		PutInstanceFunc: func(ctx context.Context, uaToken, saToken, collectionID, instanceID string, i dataset.UpdateInstance) error {
			return &testError{http.StatusInternalServerError}
		},
	}
}