package handler_test

import (
	"testing"
//	"github.com/ONSdigital/log.go/v2/log"
	"context"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/recipe"
	"github.com/ONSdigital/dp-api-clients-go/cantabular"

	"github.com/ONSdigital/dp-import-cantabular-dataset/config"
	"github.com/ONSdigital/dp-import-cantabular-dataset/handler"
	"github.com/ONSdigital/dp-import-cantabular-dataset/event"
	"github.com/ONSdigital/dp-import-cantabular-dataset/schema"

	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	"github.com/ONSdigital/dp-import-cantabular-dataset/handler/mock"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInstanceStartedHandler_HandleHappy(t *testing.T) {
	cfg := config.Config{}

	ctblrClient := mock.CantabularClientMock{
		GetCodebookFunc: func(ctx context.Context, req cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error){
			return &cantabular.GetCodebookResponse{
				Codebook: testCodebook(),
			}, nil
		},
	}

	recipeAPIClient := mock.RecipeAPIClientMock{
		GetRecipeFunc: func(ctx context.Context, uaToken, saToken, recipeID string) (*recipe.Recipe, error){
			return testRecipe(), nil
		},
	}

	datasetAPIClient := mock.DatasetAPIClientMock{
		PutInstanceFunc: func(ctx context.Context, uaToken, saToken, collectionID, instanceID string, i dataset.UpdateInstance) error {
			return nil
		},
	}

	producer := kafkatest.NewMessageProducer(true)

	h := handler.NewInstanceStarted(
		cfg,
		&ctblrClient,
		&recipeAPIClient,
		&datasetAPIClient,
		producer,
	)
	ctx := context.Background()

	Convey("Given a successful event handler, when Handle is triggered", t, func() {
		input := make(chan []byte, 2)
		go func(){
			for {
				select{
					case b := <-producer.Channels().Output:
						input <- b
					case <-time.After(time.Second * 1):
						return
				}
			}
		}()

		err := h.Handle(ctx, &event.InstanceStarted{
			RecipeID: "test-recipe-id",
			InstanceID: "test-instance-id",
			JobID: "test-job-id",
			CantabularType: "cantabular-table",
		})
		So(err, ShouldBeNil)

		time.Sleep(time.Second * 1)
		So(len(input), ShouldEqual, 2)

		err = producer.Close(ctx)
		So(err, ShouldBeNil)

		got := make([]event.CategoryDimensionImport, 2)
		for i := 0; i < 2; i++{
			b := <- input
			s := schema.CategoryDimensionImport
			err := s.Unmarshal(b, &got[i])
			So(err, ShouldBeNil)
		}

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

		So(got, ShouldResemble, expected)
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