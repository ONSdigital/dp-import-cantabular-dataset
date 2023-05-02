package handler_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular/gql"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/importapi"
	"github.com/ONSdigital/dp-api-clients-go/v2/recipe"
	"github.com/ONSdigital/log.go/v2/log"

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
	workerID                    = 1
	msgTimeout                  = time.Second
	testRecipeID                = "test-recipe-id"
	testInstanceID              = "test-instance-id"
	testJobID                   = "test-job-id"
	testBlob                    = "cantabular_blob"
	cantabularTable             = "cantabular_table"
	cantabularFlexibleTable     = "cantabular_flexible_table"
	cantabularMultivariateTable = "cantabular_multivariate_table"
)

var ctx = context.Background()

var trueValue = true
var falseValue = false

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

func testCfg() config.Config {
	return config.Config{
		KafkaConfig: config.KafkaConfig{
			Addr:                         []string{"localhost:9092", "localhost:9093"},
			InstanceStartedGroup:         "dp-import-cantabular-dataset",
			InstanceStartedTopic:         "cantabular-dataset-instance-started",
			CategoryDimensionImportTopic: "cantabular-dataset-category-dimension-import",
		},
	}
}

func TestInstanceStartedHandler_HandleHappy(t *testing.T) {
	testCfg := testCfg()

	Convey("Given a successful event handler", t, func() {
		ctblrClient := cantabularClientHappy()
		recipeAPIClient := recipeAPIClientHappy()
		importAPIClient := importAPIClientHappy()
		datasetAPIClient := datasetAPIClientHappy()
		producer, _ := kafkatest.NewProducer(ctx, &kafka.ProducerConfig{
			BrokerAddrs: testCfg.KafkaConfig.Addr,
			Topic:       testCfg.KafkaConfig.InstanceStartedTopic,
		},
			nil)

		h := handler.NewInstanceStarted(
			testCfg,
			ctblrClient,
			recipeAPIClient,
			importAPIClient,
			datasetAPIClient,
			producer.Mock,
		)

		wg := &sync.WaitGroup{}

		Convey("When Handle is triggered", func(c C) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				msg := kafkaMessage(c, &event.InstanceStarted{
					RecipeID:       testRecipeID,
					InstanceID:     testInstanceID,
					JobID:          testJobID,
					CantabularType: cantabularTable,
				})
				err := h.Handle(ctx, workerID, msg)
				c.So(err, ShouldBeNil)
			}()

			Convey("Then the expected category dimension import events are produced", func() {
				expected := []event.CategoryDimensionImport{
					{
						JobID:          testJobID,
						InstanceID:     testInstanceID,
						DimensionID:    "test-variable",
						CantabularBlob: "test-cantabular-blob",
					},
					{
						JobID:          testJobID,
						InstanceID:     testInstanceID,
						DimensionID:    "test-mapped-variable",
						CantabularBlob: "test-cantabular-blob",
					},
				}

				for _, e := range expected {
					err := producer.WaitForMessageSent(schema.CategoryDimensionImport, &e, 5*time.Second)
					So(err, ShouldBeNil)
				}

				Convey("And the expected dimensions are updated in dataset api instance, with ID and Href values comming from the corresponding codelist", func() {
					So(datasetAPIClient.PutInstanceCalls(), ShouldHaveLength, 1)
					So(datasetAPIClient.PutInstanceCalls()[0].InstanceID, ShouldEqual, testInstanceID)
					So(datasetAPIClient.PutInstanceCalls()[0].I, ShouldResemble, dataset.UpdateInstance{
						Edition:    "Test Editions 2021",
						CSVHeader:  []string{cantabularTable, "NameVar1", "NameVar2"},
						InstanceID: testInstanceID,
						Type:       cantabularTable,
						IsBasedOn: &dataset.IsBasedOn{
							ID:   "test-cantabular-blob",
							Type: cantabularTable,
						},
						Dimensions: []dataset.VersionDimension{
							{
								ID:              "test-variable",
								URL:             "http://recipe-defined-host/code-lists/test-variable",
								Label:           "LabelVar1",
								Name:            "Test Variable",
								Variable:        "NameVar1",
								NumberOfOptions: 100,
								IsAreaType:      &falseValue,
							},
							{
								ID:              "test-mapped-variable",
								URL:             "http://recipe-defined-host/code-lists/test-mapped-variable",
								Label:           "LabelVar2",
								Name:            "Test Mapped Variable",
								Variable:        "NameVar2",
								NumberOfOptions: 123,
								IsAreaType:      &falseValue,
							},
						},
					})
				})

				wg.Wait()
				err := producer.Mock.CloseFunc(ctx)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestInstanceStartedHandler_HandleUnhappy(t *testing.T) {
	testCfg := testCfg()

	Convey("Given an event handler with a recipe that cannot be found", t, func() {
		ctblrClient := cantabularClientHappy()
		recipeAPIClient := recipeAPIClientHappy()
		importAPIClient := importAPIClientHappy()
		datasetAPIClient := datasetAPIClientHappy()
		producer, _ := kafkatest.NewProducer(ctx, &kafka.ProducerConfig{
			BrokerAddrs: testCfg.KafkaConfig.Addr,
			Topic:       testCfg.KafkaConfig.InstanceStartedTopic,
		},
			nil)

		recipeAPIClient.GetRecipeFunc = func(ctx context.Context, uaToken, saToken, recipeID string) (*recipe.Recipe, error) {
			return nil, &testError{http.StatusNotFound}
		}

		h := handler.NewInstanceStarted(
			testCfg,
			ctblrClient,
			recipeAPIClient,
			importAPIClient,
			datasetAPIClient,
			producer.Mock,
		)

		Convey("When Handle is triggered", func(c C) {
			msg := kafkaMessage(c, &event.InstanceStarted{
				RecipeID:       testRecipeID,
				InstanceID:     testInstanceID,
				JobID:          testJobID,
				CantabularType: cantabularTable,
			})
			err := h.Handle(ctx, workerID, msg)

			Convey("Then the expected error is returned, with Status NotFound", func() {
				So(err, ShouldNotBeNil)
				var cerr statusCoder
				So(errors.As(err, &cerr), ShouldBeTrue)
				So(cerr.StatusCode(), ShouldEqual, http.StatusNotFound)
			})

			Convey("Then the import job and instance are set to failed state", func() {
				validateFailure(importAPIClient, datasetAPIClient, testJobID, testInstanceID)
			})
		})
	})

	Convey("Given an event handler with a dataset api client that fails to update an instance", t, func() {
		ctblrClient := cantabularClientHappy()
		recipeAPIClient := recipeAPIClientHappy()
		importAPIClient := importAPIClientHappy()
		datasetAPIClient := datasetAPIClientUnhappy()
		producer, _ := kafkatest.NewProducer(ctx, &kafka.ProducerConfig{
			BrokerAddrs: testCfg.KafkaConfig.Addr,
			Topic:       testCfg.KafkaConfig.InstanceStartedTopic,
		},
			nil)

		h := handler.NewInstanceStarted(
			testCfg,
			ctblrClient,
			recipeAPIClient,
			importAPIClient,
			datasetAPIClient,
			producer.Mock,
		)

		Convey("When Handle is triggered", func(c C) {
			msg := kafkaMessage(c, &event.InstanceStarted{
				RecipeID:       testRecipeID,
				InstanceID:     testInstanceID,
				JobID:          testJobID,
				CantabularType: cantabularTable,
			})
			err := h.Handle(ctx, workerID, msg)

			Convey("Then the expected error is returned, with Status InternalServerError", func() {
				So(err, ShouldNotBeNil)
				var cerr statusCoder
				So(errors.As(err, &cerr), ShouldBeTrue)
				So(cerr.StatusCode(), ShouldEqual, http.StatusInternalServerError)
			})

			Convey("Then the import job and instance are set to failed state", func() {
				validateFailure(importAPIClient, datasetAPIClient, testJobID, testInstanceID)
			})
		})
	})
}

func testDimensions() *cantabular.GetDimensionsResponse {
	return &cantabular.GetDimensionsResponse{
		Dataset: gql.Dataset{
			Variables: gql.Variables{
				Edges: []gql.Edge{
					{
						Node: gql.Node{
							Name:       "NameVar1",
							Label:      "LabelVar1",
							Categories: gql.Categories{TotalCount: 100},
							MapFrom:    []gql.Variables{},
						},
					},
					{
						Node: gql.Node{
							Name:       "NameVar2",
							Label:      "LabelVar2",
							Categories: gql.Categories{TotalCount: 123},
							MapFrom:    []gql.Variables{},
						},
					},
				},
			},
		},
	}
}

func testRecipe() *recipe.Recipe {
	return &recipe.Recipe{
		ID:             testRecipeID,
		CantabularBlob: "test-cantabular-blob",
		Format:         cantabularTable,
		OutputInstances: []recipe.Instance{
			{
				DatasetID: "test-dataset-id",
				Title:     "Test Instance",
				Editions:  []string{"Test Editions 2021"},
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

func testRecipeOnlyGeography() *recipe.Recipe {
	trueValue := true
	falseValue := false
	return &recipe.Recipe{
		Alias:          "Cantabular Example 2",
		Format:         cantabularTable,
		CantabularBlob: testBlob,
		OutputInstances: []recipe.Instance{
			{
				CodeLists: []recipe.CodeList{
					{
						IsCantabularGeography:        &trueValue,
						IsCantabularDefaultGeography: &falseValue,
					},
				},
			},
		},
	}
}

func testRecipeOnlyGeographyWithEdition() *recipe.Recipe {
	trueValue := true
	falseValue := false
	return &recipe.Recipe{
		Alias:          "Cantabular Example 2",
		Format:         cantabularTable,
		CantabularBlob: testBlob,
		OutputInstances: []recipe.Instance{
			{
				CodeLists: []recipe.CodeList{
					{
						IsCantabularGeography:        &trueValue,
						IsCantabularDefaultGeography: &falseValue,
					},
					{
						IsCantabularGeography:        &trueValue,
						IsCantabularDefaultGeography: &falseValue,
					},
				},
				Editions: []string{"2021"},
			},
		},
	}
}

func testRecipeGeographyWithEdition() *recipe.Recipe {
	trueValue := true
	return &recipe.Recipe{
		Alias:          "Cantabular Example 2",
		Format:         cantabularTable,
		CantabularBlob: testBlob,
		OutputInstances: []recipe.Instance{
			{
				CodeLists: []recipe.CodeList{
					{
						IsCantabularGeography: &trueValue,
					},
					{
						IsCantabularGeography: &trueValue,
					},
				},
				Editions: []string{"2021"},
			},
		},
	}
}

func cantabularClientHappy() *mock.CantabularClientMock {
	return &mock.CantabularClientMock{
		GetDimensionsByNameFunc: func(ctx context.Context, req cantabular.GetDimensionsByNameRequest) (*cantabular.GetDimensionsResponse, error) {
			return testDimensions(), nil
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

func recipeAPIClientOnlyGeography() *mock.RecipeAPIClientMock {
	return &mock.RecipeAPIClientMock{
		GetRecipeFunc: func(ctx context.Context, uaToken, saToken, recipeID string) (*recipe.Recipe, error) {
			return testRecipeOnlyGeography(), nil
		},
	}
}

func recipeAPIClientOnlyGeographyWithEdition() *mock.RecipeAPIClientMock {
	return &mock.RecipeAPIClientMock{
		GetRecipeFunc: func(ctx context.Context, uaToken, saToken, recipeID string) (*recipe.Recipe, error) {
			return testRecipeOnlyGeographyWithEdition(), nil
		},
	}
}

func recipeAPIClientGeographyWithEdition() *mock.RecipeAPIClientMock {
	return &mock.RecipeAPIClientMock{
		GetRecipeFunc: func(ctx context.Context, uaToken, saToken, recipeID string) (*recipe.Recipe, error) {
			return testRecipeGeographyWithEdition(), nil
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
	message, err := kafkatest.NewMessage(b, 0)
	c.So(err, ShouldBeNil)
	return message
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
		fmt.Println("TESTING")
		fmt.Println(err)
		c.So(err, ShouldBeNil)
		c.So(got, ShouldResemble, expected)
	case <-time.After(msgTimeout):
		fmt.Println("IN TEST FAILURE")
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

func validateMessagesCount(t *testing.T, producer kafka.IProducer, c C, expected event.CategoryDimensionImport) bool {
	var consumed bool

	select {
	case b := <-producer.Channels().Output:
		var got event.CategoryDimensionImport
		s := schema.CategoryDimensionImport
		err := s.Unmarshal(b, &got)
		c.So(err, ShouldBeNil)
		c.So(got, ShouldResemble, expected)
		consumed = true
	case <-time.After(msgTimeout):
		// a timeout is expected, so the following is a dummy check that passes
		c.So(1, ShouldEqual, 1)
	}
	return consumed
}

func TestInstanceStartedHandler_HandleUnhappyNoEdition(t *testing.T) {
	testCfg := testCfg()

	Convey("Given a successful event handler, with no Editions", t, func() {
		ctblrClient := cantabularClientHappy()
		recipeAPIClient := recipeAPIClientOnlyGeography()
		importAPIClient := importAPIClientHappy()
		datasetAPIClient := datasetAPIClientHappy()
		producer, _ := kafkatest.NewProducer(ctx, &kafka.ProducerConfig{
			BrokerAddrs: testCfg.KafkaConfig.Addr,
			Topic:       testCfg.KafkaConfig.InstanceStartedTopic,
		},
			nil)

		h := handler.NewInstanceStarted(
			testCfg,
			ctblrClient,
			recipeAPIClient,
			importAPIClient,
			datasetAPIClient,
			producer.Mock,
		)

		wg := &sync.WaitGroup{}

		Convey("When Handle is triggered", func(c C) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				msg := kafkaMessage(c, &event.InstanceStarted{
					RecipeID:       testRecipeID,
					InstanceID:     testInstanceID,
					JobID:          testJobID,
					CantabularType: cantabularTable,
				})
				err := h.Handle(ctx, workerID, msg)
				Convey("Then the expected error is seen", t, func() {
					c.So(err, ShouldResemble, errors.New("no editions found in instance"))
				})
			}()

			Convey("Then the validateMessagesCount consumed no messages", func() {
				expected := []event.CategoryDimensionImport{
					{
						JobID:          testJobID,
						InstanceID:     testInstanceID,
						DimensionID:    "test-variable",
						CantabularBlob: "test-cantabular-blob",
					},
				}

				var result int
				for _, e := range expected {
					err := producer.WaitForMessageSent(schema.CategoryDimensionImport, &e, 5*time.Second)
					if err == nil {
						result++
					}
					// if validateMessagesCount(t, producer.Mock, c, e) == true {
					// 	result++
					// }
				}

				wg.Wait()
				So(result, ShouldEqual, 0)

				err := producer.Mock.CloseFunc(ctx)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestCreateUpdateInstanceRequest_Happy(t *testing.T) {
	testCfg := testCfg()

	Convey("Given CreateUpdateInstanceRequest() is called with two Edges and the format is of type: cantabular_table", t, func() {
		ctblrClient := cantabularClientHappy()
		recipeAPIClient := recipeAPIClientOnlyGeographyWithEdition()
		importAPIClient := importAPIClientHappy()
		datasetAPIClient := datasetAPIClientHappy()
		producer, _ := kafkatest.NewProducer(ctx, &kafka.ProducerConfig{
			BrokerAddrs: testCfg.KafkaConfig.Addr,
			Topic:       testCfg.KafkaConfig.InstanceStartedTopic,
		},
			nil)

		h := handler.NewInstanceStarted(
			testCfg,
			ctblrClient,
			recipeAPIClient,
			importAPIClient,
			datasetAPIClient,
			producer.Mock,
		)

		var mfVariables gql.Variables

		mfVariables.Edges = []gql.Edge{
			{
				Node: gql.Node{
					Name:  "NameVar1",
					Label: "LabelVar1",
					Categories: gql.Categories{
						TotalCount: 100,
					},
				},
			},
			{
				Node: gql.Node{
					Name:  "NameVar2",
					Label: "LabelVar2",
					Categories: gql.Categories{
						TotalCount: 123,
					},
				},
			},
		}

		e := &event.InstanceStarted{
			RecipeID:       testRecipeID,
			InstanceID:     testInstanceID,
			JobID:          testJobID,
			CantabularType: cantabularTable,
		}

		codelists := []recipe.CodeList{
			{
				IsCantabularGeography:        &trueValue,
				IsCantabularDefaultGeography: &falseValue,
			},
			{
				IsCantabularGeography:        &trueValue,
				IsCantabularDefaultGeography: &falseValue,
			},
		}

		r := &recipe.Recipe{
			Format:         cantabularTable,
			CantabularBlob: testBlob,
		}

		req := h.CreateUpdateInstanceRequest(ctx, mfVariables, e, r, codelists, "2021")

		Convey("Then we get the expected result with two Dimensions", func() {

			expected := dataset.UpdateInstance{
				CollectionID: "",
				Downloads:    dataset.DownloadList{},
				Edition:      "2021",
				Dimensions: []dataset.VersionDimension{
					{
						ID:              "NameVar1",
						Name:            "NameVar1",
						Label:           "LabelVar1",
						URL:             "/code-lists/NameVar1",
						Variable:        "NameVar1",
						NumberOfOptions: 100,
						IsAreaType:      &trueValue,
					},
					{
						ID:              "NameVar2",
						Name:            "NameVar2",
						Label:           "LabelVar2",
						URL:             "/code-lists/NameVar2",
						Variable:        "NameVar2",
						NumberOfOptions: 123,
						IsAreaType:      &trueValue,
					},
				},
				ID:         "",
				InstanceID: "test-instance-id",
				CSVHeader: []string{
					cantabularTable,
					"NameVar1",
					"NameVar2"},
				Type: cantabularTable,
				IsBasedOn: &dataset.IsBasedOn{
					Type: cantabularTable,
					ID:   testBlob,
				},
			}
			// in the following we use ShouldResemble because 'IsBasedOn' exists in a different memory location
			So(req, ShouldResemble, expected)
		})
	})
}

func TestCreateUpdateInstanceRequest_Flexible_NoGeography(t *testing.T) {
	testCfg := testCfg()
	trueValue := true
	falseValue := false

	Convey("Given CreateUpdateInstanceRequest() is called with two Edges and the format is of type: cantabular_flexible_table, and neither edges are Geography", t, func() {
		ctblrClient := cantabularClientHappy()
		recipeAPIClient := recipeAPIClientOnlyGeographyWithEdition()
		importAPIClient := importAPIClientHappy()
		datasetAPIClient := datasetAPIClientHappy()
		producer, _ := kafkatest.NewProducer(ctx, &kafka.ProducerConfig{
			BrokerAddrs: testCfg.KafkaConfig.Addr,
			Topic:       testCfg.KafkaConfig.InstanceStartedTopic,
		},
			nil)

		h := handler.NewInstanceStarted(
			testCfg,
			ctblrClient,
			recipeAPIClient,
			importAPIClient,
			datasetAPIClient,
			producer.Mock,
		)

		var mfVariables gql.Variables

		mfVariables.Edges = []gql.Edge{
			{
				Node: gql.Node{
					Name:  "NameVar1",
					Label: "LabelVar1",
					Categories: gql.Categories{
						TotalCount: 100,
					},
				},
			},
			{
				Node: gql.Node{
					Name:  "NameVar2",
					Label: "LabelVar2",
					Categories: gql.Categories{
						TotalCount: 123,
					},
				},
			},
		}

		e := &event.InstanceStarted{
			RecipeID:       testRecipeID,
			InstanceID:     testInstanceID,
			JobID:          testJobID,
			CantabularType: cantabularTable,
		}

		codelists := []recipe.CodeList{
			{
				IsCantabularGeography:        &falseValue,
				IsCantabularDefaultGeography: &trueValue,
			},
			{
				IsCantabularGeography:        &falseValue,
				IsCantabularDefaultGeography: &trueValue,
			},
		}

		r := &recipe.Recipe{
			Format:         cantabularFlexibleTable,
			CantabularBlob: testBlob,
		}

		req := h.CreateUpdateInstanceRequest(ctx, mfVariables, e, r, codelists, "2021")

		Convey("Then we get the expected result with two Dimensions", func() {

			expected := dataset.UpdateInstance{
				CollectionID: "",
				Downloads:    dataset.DownloadList{},
				Edition:      "2021",
				Dimensions: []dataset.VersionDimension{
					{
						ID:              "NameVar1",
						Name:            "NameVar1",
						Label:           "LabelVar1",
						URL:             "/code-lists/NameVar1",
						Variable:        "NameVar1",
						NumberOfOptions: 100,
						IsAreaType:      &falseValue,
					},
					{
						ID:              "NameVar2",
						Name:            "NameVar2",
						Label:           "LabelVar2",
						URL:             "/code-lists/NameVar2",
						Variable:        "NameVar2",
						NumberOfOptions: 123,
						IsAreaType:      &falseValue,
					},
				},
				ID:         "",
				InstanceID: "test-instance-id",
				CSVHeader: []string{
					cantabularTable,
					"NameVar1",
					"NameVar2"},
				Type: cantabularFlexibleTable,
				IsBasedOn: &dataset.IsBasedOn{
					Type: cantabularFlexibleTable,
					ID:   testBlob,
				},
			}
			// in the following we use ShouldResemble because 'IsBasedOn' exists in a different memory location
			So(req, ShouldResemble, expected)
		})
	})
}

func TestCreateUpdateInstanceRequest_Flexible_OneGeography(t *testing.T) {
	testCfg := testCfg()
	falseValue := false
	trueValue := true

	Convey("Given CreateUpdateInstanceRequest() is called with two Edges and the format is of type: cantabular_flexible_table, and one edge is Geography", t, func() {
		ctblrClient := cantabularClientHappy()
		recipeAPIClient := recipeAPIClientOnlyGeographyWithEdition()
		importAPIClient := importAPIClientHappy()
		datasetAPIClient := datasetAPIClientHappy()
		producer, _ := kafkatest.NewProducer(ctx, &kafka.ProducerConfig{
			BrokerAddrs: testCfg.KafkaConfig.Addr,
			Topic:       testCfg.KafkaConfig.InstanceStartedTopic,
		},
			nil)

		h := handler.NewInstanceStarted(
			testCfg,
			ctblrClient,
			recipeAPIClient,
			importAPIClient,
			datasetAPIClient,
			producer.Mock,
		)

		var mfVariables gql.Variables

		mfVariables.Edges = []gql.Edge{
			{
				Node: gql.Node{
					Name:  "NameVar1",
					Label: "LabelVar1",
					Categories: gql.Categories{
						TotalCount: 100,
					},
				},
			},
			{
				Node: gql.Node{
					Name:  "NameVar2",
					Label: "LabelVar2",
					Categories: gql.Categories{
						TotalCount: 123,
					},
				},
			},
		}

		e := &event.InstanceStarted{
			RecipeID:       testRecipeID,
			InstanceID:     testInstanceID,
			JobID:          testJobID,
			CantabularType: cantabularTable,
		}

		codelists := []recipe.CodeList{
			{
				IsCantabularGeography:        &trueValue,
				IsCantabularDefaultGeography: &falseValue,
			},
			{
				IsCantabularGeography:        &falseValue,
				IsCantabularDefaultGeography: &trueValue,
			},
		}

		r := &recipe.Recipe{
			Format:         cantabularFlexibleTable,
			CantabularBlob: testBlob,
		}

		req := h.CreateUpdateInstanceRequest(ctx, mfVariables, e, r, codelists, "2021")

		Convey("Then we get the expected single non-Geography Dimension", func() {

			expected := dataset.UpdateInstance{
				CollectionID: "",
				Downloads:    dataset.DownloadList{},
				Edition:      "2021",
				Dimensions: []dataset.VersionDimension{
					{
						ID:              "NameVar2",
						Name:            "NameVar2",
						Label:           "LabelVar2",
						URL:             "/code-lists/NameVar2",
						Variable:        "NameVar2",
						NumberOfOptions: 123,
						IsAreaType:      &falseValue,
					},
				},
				ID:         "",
				InstanceID: "test-instance-id",
				CSVHeader: []string{
					cantabularTable,
					"NameVar2"},
				Type: cantabularFlexibleTable,
				IsBasedOn: &dataset.IsBasedOn{
					Type: cantabularFlexibleTable,
					ID:   testBlob,
				},
			}
			// in the following we use ShouldResemble because 'IsBasedOn' exists in a different memory location
			So(req, ShouldResemble, expected)
		})
	})

}

func TestCreateUpdateInstanceRequest_Flexible_BothGeography(t *testing.T) {
	testCfg := testCfg()
	trueValue := true
	falseValue := false

	Convey("Given CreateUpdateInstanceRequest() is called with two Edges and the format is of type: cantabular_flexible_table, and both edges are Geography", t, func() {
		ctblrClient := cantabularClientHappy()
		recipeAPIClient := recipeAPIClientOnlyGeographyWithEdition()
		importAPIClient := importAPIClientHappy()
		datasetAPIClient := datasetAPIClientHappy()
		producer, _ := kafkatest.NewProducer(ctx, &kafka.ProducerConfig{
			BrokerAddrs: testCfg.KafkaConfig.Addr,
			Topic:       testCfg.KafkaConfig.InstanceStartedTopic,
		},
			nil)

		h := handler.NewInstanceStarted(
			testCfg,
			ctblrClient,
			recipeAPIClient,
			importAPIClient,
			datasetAPIClient,
			producer.Mock,
		)

		var mfVariables gql.Variables

		mfVariables.Edges = []gql.Edge{
			{
				Node: gql.Node{
					Name:  "NameVar1",
					Label: "LabelVar1",
					Categories: gql.Categories{
						TotalCount: 100,
					},
				},
			},
			{
				Node: gql.Node{
					Name:  "NameVar2",
					Label: "LabelVar2",
					Categories: gql.Categories{
						TotalCount: 123,
					},
				},
			},
		}

		e := &event.InstanceStarted{
			RecipeID:       testRecipeID,
			InstanceID:     testInstanceID,
			JobID:          testJobID,
			CantabularType: cantabularTable,
		}

		codelists := []recipe.CodeList{
			{
				IsCantabularGeography:        &trueValue,
				IsCantabularDefaultGeography: &falseValue,
			},
			{
				IsCantabularGeography:        &trueValue,
				IsCantabularDefaultGeography: &falseValue,
			},
		}

		r := &recipe.Recipe{
			Format:         cantabularFlexibleTable,
			CantabularBlob: testBlob,
		}

		req := h.CreateUpdateInstanceRequest(ctx, mfVariables, e, r, codelists, "2021")

		Convey("Then we get the expected result with no Dimensions in it", func() {

			expected := dataset.UpdateInstance{
				CollectionID: "",
				Downloads:    dataset.DownloadList{},
				Edition:      "2021",
				ID:           "",
				InstanceID:   "test-instance-id",
				CSVHeader: []string{
					cantabularTable,
				},
				Type: cantabularFlexibleTable,
				IsBasedOn: &dataset.IsBasedOn{
					Type: cantabularFlexibleTable,
					ID:   testBlob,
				},
			}
			// in the following we use ShouldResemble because 'IsBasedOn' exists in a different memory location
			So(req, ShouldResemble, expected)
		})
	})
}

func TestTriggerImportDimensionOptions(t *testing.T) {
	testCfg := testCfg()
	trueValue := true
	falseValue := false

	Convey("Given a successful event handler, with two Edges and the format is of type: cantabular_flexible_table, with one non-geography", t, func() {
		ctblrClient := cantabularClientHappy()
		recipeAPIClient := recipeAPIClientGeographyWithEdition()
		importAPIClient := importAPIClientHappy()
		datasetAPIClient := datasetAPIClientHappy()
		producer, _ := kafkatest.NewProducer(ctx, &kafka.ProducerConfig{
			BrokerAddrs: testCfg.KafkaConfig.Addr,
			Topic:       testCfg.KafkaConfig.InstanceStartedTopic,
		},
			nil)

		h := handler.NewInstanceStarted(
			testCfg,
			ctblrClient,
			recipeAPIClient,
			importAPIClient,
			datasetAPIClient,
			producer.Mock,
		)

		wg := &sync.WaitGroup{}

		Convey("When TriggerImportDimensionOptions is called", func(c C) {
			codelists := []recipe.CodeList{
				{
					IsCantabularGeography:        &trueValue,
					IsCantabularDefaultGeography: &falseValue,
				},
				{
					IsCantabularGeography:        &falseValue, // this indicates the non-geography item
					IsCantabularDefaultGeography: &falseValue,
				},
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				r := &recipe.Recipe{
					Format:         cantabularFlexibleTable,
					CantabularBlob: testBlob,
				}

				err := h.TriggerImportDimensionOptions(r, codelists, &event.InstanceStarted{
					RecipeID:       testRecipeID,
					InstanceID:     testInstanceID,
					JobID:          testJobID,
					CantabularType: cantabularTable,
				})
				Convey("Then no error is seen", t, func() {
					c.So(err, ShouldBeNil)
				})
			}()

			Convey("Then the validateMessagesCount consumes one message", func() {
				expected := make([]event.CategoryDimensionImport, 0)

				for _, cl := range codelists {
					event := event.CategoryDimensionImport{
						JobID:          testJobID,
						InstanceID:     testInstanceID,
						DimensionID:    "",
						CantabularBlob: testBlob,
						IsGeography:    *cl.IsCantabularGeography,
					}
					expected = append(expected, event)
				}

				var result int
				for _, e := range expected {
					err := producer.WaitForMessageSent(schema.CategoryDimensionImport, &e, 5*time.Second)
					if err == nil {
						result++
					}
					// if validateMessagesCount(t, producer.Mock, c, e) == true {
					// 	result++
					// }
				}

				wg.Wait()
				So(result, ShouldEqual, 2)

				err := producer.Mock.CloseFunc(ctx)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestTriggerImportDimensionOptionsNonGeography(t *testing.T) {
	testCfg := testCfg()
	trueValue := true
	falseValue := false

	Convey("Given a successful event handler, with two Edges and the format is of type: cantabular_flexible_table, where both edges are geography", t, func() {
		ctblrClient := cantabularClientHappy()
		recipeAPIClient := recipeAPIClientGeographyWithEdition()
		importAPIClient := importAPIClientHappy()
		datasetAPIClient := datasetAPIClientHappy()
		producer, _ := kafkatest.NewProducer(ctx, &kafka.ProducerConfig{
			BrokerAddrs: testCfg.KafkaConfig.Addr,
			Topic:       testCfg.KafkaConfig.InstanceStartedTopic,
		},
			nil)

		h := handler.NewInstanceStarted(
			testCfg,
			ctblrClient,
			recipeAPIClient,
			importAPIClient,
			datasetAPIClient,
			producer.Mock,
		)

		wg := &sync.WaitGroup{}
		Convey("When TriggerImportDimensionOptions is called", func(c C) {
			codelists := []recipe.CodeList{
				{
					IsCantabularGeography:        &trueValue,
					IsCantabularDefaultGeography: &falseValue,
				},
				{
					IsCantabularGeography:        &trueValue,
					IsCantabularDefaultGeography: &falseValue,
				},
			}

			wg.Add(1)
			go func() {
				defer wg.Done()

				errs := handler.NewError(
					fmt.Errorf("only geography codelists exist in this instance, there must be at least one non-geography in the codelists"),
					log.Data{},
					false)

				r := &recipe.Recipe{
					Format:         cantabularFlexibleTable,
					CantabularBlob: testBlob,
				}

				err := h.TriggerImportDimensionOptions(r, codelists, &event.InstanceStarted{
					RecipeID:       testRecipeID,
					InstanceID:     testInstanceID,
					JobID:          testJobID,
					CantabularType: cantabularTable,
				})
				Convey("Then the expected error is seen", t, func() {
					c.So(err[0], ShouldResemble, errs)
				})
			}()

			Convey("Then the validateMessagesCount consumes no messages", func() {
				expected := make([]event.CategoryDimensionImport, 0)

				for _, cl := range codelists {
					event := event.CategoryDimensionImport{
						JobID:          testJobID,
						InstanceID:     testInstanceID,
						DimensionID:    "",
						CantabularBlob: testBlob,
						IsGeography:    *cl.IsCantabularGeography,
					}
					expected = append(expected, event)
				}

				var result int
				for _, e := range expected {
					err := producer.WaitForMessageSent(schema.CategoryDimensionImport, &e, 5*time.Second)
					if err == nil {
						result++
					}
					// if validateMessagesCount(t, producer.Mock, c, e) == true {
					// 	result++
					// }
				}

				wg.Wait()
				So(result, ShouldEqual, 2)

				err := producer.Mock.CloseFunc(ctx)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestCreateUpdateInstanceRequest_Multivariate_NoGeography(t *testing.T) {
	testCfg := testCfg()
	trueValue := true
	falseValue := false

	Convey("Given CreateUpdateInstanceRequest() is called with two Edges and the format is of type: cantabular_multivariate_table, and neither edges are Geography", t, func() {
		ctblrClient := cantabularClientHappy()
		recipeAPIClient := recipeAPIClientOnlyGeographyWithEdition()
		importAPIClient := importAPIClientHappy()
		datasetAPIClient := datasetAPIClientHappy()
		producer, _ := kafkatest.NewProducer(ctx, &kafka.ProducerConfig{
			BrokerAddrs: testCfg.KafkaConfig.Addr,
			Topic:       testCfg.KafkaConfig.InstanceStartedTopic,
		},
			nil)

		h := handler.NewInstanceStarted(
			testCfg,
			ctblrClient,
			recipeAPIClient,
			importAPIClient,
			datasetAPIClient,
			producer.Mock,
		)

		var mfVariables gql.Variables

		mfVariables.Edges = []gql.Edge{
			{
				Node: gql.Node{
					Name:  "NameVar1",
					Label: "LabelVar1",
					Categories: gql.Categories{
						TotalCount: 100,
					},
				},
			},
			{
				Node: gql.Node{
					Name:  "NameVar2",
					Label: "LabelVar2",
					Categories: gql.Categories{
						TotalCount: 123,
					},
				},
			},
		}

		e := &event.InstanceStarted{
			RecipeID:       testRecipeID,
			InstanceID:     testInstanceID,
			JobID:          testJobID,
			CantabularType: cantabularTable,
		}

		codelists := []recipe.CodeList{
			{
				IsCantabularGeography:        &falseValue,
				IsCantabularDefaultGeography: &trueValue,
			},
			{
				IsCantabularGeography:        &falseValue,
				IsCantabularDefaultGeography: &trueValue,
			},
		}

		r := &recipe.Recipe{
			Format:         cantabularMultivariateTable,
			CantabularBlob: testBlob,
		}

		req := h.CreateUpdateInstanceRequest(ctx, mfVariables, e, r, codelists, "2021")

		Convey("Then we get the expected result with two Dimensions", func() {

			expected := dataset.UpdateInstance{
				CollectionID: "",
				Downloads:    dataset.DownloadList{},
				Edition:      "2021",
				Dimensions: []dataset.VersionDimension{
					{
						ID:              "NameVar1",
						Name:            "NameVar1",
						Label:           "LabelVar1",
						URL:             "/code-lists/NameVar1",
						Variable:        "NameVar1",
						NumberOfOptions: 100,
						IsAreaType:      &falseValue,
					},
					{
						ID:              "NameVar2",
						Name:            "NameVar2",
						Label:           "LabelVar2",
						URL:             "/code-lists/NameVar2",
						Variable:        "NameVar2",
						NumberOfOptions: 123,
						IsAreaType:      &falseValue,
					},
				},
				ID:         "",
				InstanceID: "test-instance-id",
				CSVHeader: []string{
					cantabularTable,
					"NameVar1",
					"NameVar2"},
				Type: cantabularMultivariateTable,
				IsBasedOn: &dataset.IsBasedOn{
					Type: cantabularMultivariateTable,
					ID:   testBlob,
				},
			}
			// in the following we use ShouldResemble because 'IsBasedOn' exists in a different memory location
			So(req, ShouldResemble, expected)
		})
	})
}

func TestCreateUpdateInstanceRequest_Multivariate_OneGeography(t *testing.T) {
	testCfg := testCfg()
	falseValue := false
	trueValue := true

	Convey("Given CreateUpdateInstanceRequest() is called with two Edges and the format is of type: cantabular_multivariate_table, and one edge is Geography", t, func() {
		ctblrClient := cantabularClientHappy()
		recipeAPIClient := recipeAPIClientOnlyGeographyWithEdition()
		importAPIClient := importAPIClientHappy()
		datasetAPIClient := datasetAPIClientHappy()
		producer, _ := kafkatest.NewProducer(ctx, &kafka.ProducerConfig{
			BrokerAddrs: testCfg.KafkaConfig.Addr,
			Topic:       testCfg.KafkaConfig.InstanceStartedTopic,
		},
			nil)

		h := handler.NewInstanceStarted(
			testCfg,
			ctblrClient,
			recipeAPIClient,
			importAPIClient,
			datasetAPIClient,
			producer.Mock,
		)

		var mfVariables gql.Variables

		mfVariables.Edges = []gql.Edge{
			{
				Node: gql.Node{
					Name:  "NameVar1",
					Label: "LabelVar1",
					Categories: gql.Categories{
						TotalCount: 100,
					},
				},
			},
			{
				Node: gql.Node{
					Name:  "NameVar2",
					Label: "LabelVar2",
					Categories: gql.Categories{
						TotalCount: 123,
					},
				},
			},
		}

		e := &event.InstanceStarted{
			RecipeID:       testRecipeID,
			InstanceID:     testInstanceID,
			JobID:          testJobID,
			CantabularType: cantabularTable,
		}

		codelists := []recipe.CodeList{
			{
				IsCantabularGeography:        &trueValue,
				IsCantabularDefaultGeography: &falseValue,
			},
			{
				IsCantabularGeography:        &falseValue,
				IsCantabularDefaultGeography: &trueValue,
			},
		}

		r := &recipe.Recipe{
			Format:         cantabularMultivariateTable,
			CantabularBlob: testBlob,
		}

		req := h.CreateUpdateInstanceRequest(ctx, mfVariables, e, r, codelists, "2021")

		Convey("Then we get the expected single non-Geography Dimension", func() {

			expected := dataset.UpdateInstance{
				CollectionID: "",
				Downloads:    dataset.DownloadList{},
				Edition:      "2021",
				Dimensions: []dataset.VersionDimension{
					{
						ID:              "NameVar2",
						Name:            "NameVar2",
						Label:           "LabelVar2",
						URL:             "/code-lists/NameVar2",
						Variable:        "NameVar2",
						NumberOfOptions: 123,
						IsAreaType:      &falseValue,
					},
				},
				ID:         "",
				InstanceID: "test-instance-id",
				CSVHeader: []string{
					cantabularTable,
					"NameVar2"},
				Type: cantabularMultivariateTable,
				IsBasedOn: &dataset.IsBasedOn{
					Type: cantabularMultivariateTable,
					ID:   testBlob,
				},
			}
			// in the following we use ShouldResemble because 'IsBasedOn' exists in a different memory location
			So(req, ShouldResemble, expected)
		})
	})

}

func TestCreateUpdateInstanceRequest_Multivariate_BothGeography(t *testing.T) {
	testCfg := testCfg()
	trueValue := true
	falseValue := false

	Convey("Given CreateUpdateInstanceRequest() is called with two Edges and the format is of type: cantabular_multivariate_table, and both edges are Geography", t, func() {
		ctblrClient := cantabularClientHappy()
		recipeAPIClient := recipeAPIClientOnlyGeographyWithEdition()
		importAPIClient := importAPIClientHappy()
		datasetAPIClient := datasetAPIClientHappy()
		producer, _ := kafkatest.NewProducer(ctx, &kafka.ProducerConfig{
			BrokerAddrs: testCfg.KafkaConfig.Addr,
			Topic:       testCfg.KafkaConfig.InstanceStartedTopic,
		},
			nil)

		h := handler.NewInstanceStarted(
			testCfg,
			ctblrClient,
			recipeAPIClient,
			importAPIClient,
			datasetAPIClient,
			producer.Mock,
		)

		var mfVariables gql.Variables

		mfVariables.Edges = []gql.Edge{
			{
				Node: gql.Node{
					Name:  "NameVar1",
					Label: "LabelVar1",
					Categories: gql.Categories{
						TotalCount: 100,
					},
				},
			},
			{
				Node: gql.Node{
					Name:  "NameVar2",
					Label: "LabelVar2",
					Categories: gql.Categories{
						TotalCount: 123,
					},
				},
			},
		}

		e := &event.InstanceStarted{
			RecipeID:       testRecipeID,
			InstanceID:     testInstanceID,
			JobID:          testJobID,
			CantabularType: cantabularTable,
		}

		codelists := []recipe.CodeList{
			{
				IsCantabularGeography:        &trueValue,
				IsCantabularDefaultGeography: &falseValue,
			},
			{
				IsCantabularGeography:        &trueValue,
				IsCantabularDefaultGeography: &falseValue,
			},
		}

		r := &recipe.Recipe{
			Format:         cantabularMultivariateTable,
			CantabularBlob: testBlob,
		}

		req := h.CreateUpdateInstanceRequest(ctx, mfVariables, e, r, codelists, "2021")

		Convey("Then we get the expected result with no Dimensions in it", func() {

			expected := dataset.UpdateInstance{
				CollectionID: "",
				Downloads:    dataset.DownloadList{},
				Edition:      "2021",
				ID:           "",
				InstanceID:   "test-instance-id",
				CSVHeader: []string{
					cantabularTable,
				},
				Type: cantabularMultivariateTable,
				IsBasedOn: &dataset.IsBasedOn{
					Type: cantabularMultivariateTable,
					ID:   testBlob,
				},
			}
			// in the following we use ShouldResemble because 'IsBasedOn' exists in a different memory location
			So(req, ShouldResemble, expected)
		})
	})
}
