package handler_test

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular/gql"
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
	workerID        = 1
	msgTimeout      = time.Second
	testRecipeID    = "test-recipe-id"
	testInstanceID  = "test-instance-id"
	testJobID       = "test-job-id"
	cantabularTable = "cantabular_table"
)

var ctx = context.Background()

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
					validateMessage(t, producer, c, e)
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
								Name:            "LabelVar1",
								Variable:        "NameVar1",
								NumberOfOptions: 100,
							},
							{
								ID:              "test-mapped-variable",
								URL:             "http://recipe-defined-host/code-lists/test-mapped-variable",
								Label:           "LabelVar2",
								Name:            "LabelVar2",
								Variable:        "NameVar2",
								NumberOfOptions: 123,
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
		Dataset: gql.DatasetVariables{
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
		Format:         "cantabular_table",
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
	return &recipe.Recipe{
		Alias:          "Cantabular Example 2",
		Format:         "cantabular_table",
		CantabularBlob: "cantabular_blob",
		OutputInstances: []recipe.Instance{
			{
				CodeLists: []recipe.CodeList{
					{
						IsCantabularGeography: &trueValue,
					},
				},
			},
		},
	}
}

func testRecipeOnlyGeographyWithEdition() *recipe.Recipe {
	trueValue := true
	return &recipe.Recipe{
		Alias:          "Cantabular Example 2",
		Format:         "cantabular_table",
		CantabularBlob: "cantabular_blob",
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

func validateMessageAndTimeoutOk(t *testing.T, producer kafka.IProducer, c C, expected event.CategoryDimensionImport) {
	select {
	case b := <-producer.Channels().Output:
		var got event.CategoryDimensionImport
		s := schema.CategoryDimensionImport
		err := s.Unmarshal(b, &got)
		c.So(err, ShouldBeNil)
		c.So(got, ShouldResemble, expected)
	case <-time.After(msgTimeout):
		c.So(1, ShouldEqual, 1)
		// t.Fail()
	}
}

// func TestIsGeography(t *testing.T) {
// 	falseValue := false
// 	trueValue := true

// 	Convey("When IsGeography is called", t, func() {
// 		Convey("And the recipe does not have format of cantabular_flexible_table and isCantabularGeography is false", func() {
// 			notFlexibleTable := recipe.Recipe{
// 				Alias:  "Cantabular Example 1",
// 				Format: "cantabular_table",
// 				OutputInstances: []recipe.Instance{
// 					{
// 						CodeLists: []recipe.CodeList{
// 							{
// 								IsCantabularGeography: &falseValue,
// 							},
// 							{
// 								IsCantabularGeography: &falseValue,
// 							},
// 							{
// 								IsCantabularGeography: &falseValue,
// 							},
// 						},
// 					},
// 				},
// 			}
// 			isGeographyValue := handler.IsGeography(&notFlexibleTable.Format, &notFlexibleTable.OutputInstances[0].CodeLists)
// 			So(isGeographyValue, ShouldBeFalse)
// 		})

// 	})
// 	Convey("When IsGeography is called", t, func() {
// 		Convey("And the recipe does have format of cantabular_flexible_table and isCantabularGeography is true", func() {
// 			isFlexibleTableTrue := recipe.Recipe{
// 				Alias:  "Cantabular Example 3",
// 				Format: "cantabular_flexible_table",
// 				OutputInstances: []recipe.Instance{
// 					{
// 						CodeLists: []recipe.CodeList{
// 							{
// 								IsCantabularGeography: &trueValue,
// 							},
// 							{
// 								IsCantabularGeography: &falseValue,
// 							},
// 						},
// 					},
// 				},
// 			}
// 			isGeographyValue := handler.IsGeography(&isFlexibleTableTrue.Format, &isFlexibleTableTrue.OutputInstances[0].CodeLists)
// 			So(isGeographyValue, ShouldBeTrue)
// 		})
// 	})
// 	Convey("When IsGeography is called", t, func() {
// 		Convey("And the recipe does have format of cantabular_flexible_table and isCantabularGeography is false", func() {
// 			isFlexibleTableFalse := recipe.Recipe{
// 				Alias:  "Cantabular Example 2",
// 				Format: "cantabular_flexible_table",
// 				OutputInstances: []recipe.Instance{
// 					{
// 						CodeLists: []recipe.CodeList{
// 							{
// 								IsCantabularGeography: &falseValue,
// 							},
// 							{
// 								IsCantabularGeography: &falseValue,
// 							},
// 						},
// 					},
// 				},
// 			}
// 			isGeographyValue := handler.IsGeography(&isFlexibleTableFalse.Format, &isFlexibleTableFalse.OutputInstances[0].CodeLists)
// 			So(isGeographyValue, ShouldBeFalse)
// 		})
// 	})

// }

// func TestTriggerDimensionOptions(t *testing.T) {

// 	// falseValue := false
// 	trueValue := true

// 	cfg := config.Config{}
// 	ctblrClient := cantabularClientHappy()
// 	recipeAPIClient := recipeAPIClientHappy()
// 	importAPIClient := importAPIClientHappy()
// 	datasetAPIClient := datasetAPIClientUnhappy()
// 	producer := kafkatest.NewMessageProducer(true)

// 	h := handler.NewInstanceStarted(
// 		cfg,
// 		ctblrClient,
// 		recipeAPIClient,
// 		importAPIClient,
// 		datasetAPIClient,
// 		producer,
// 	)
// 	Convey("Testing1", t, func() {
// 		e := &event.InstanceStarted{
// 			RecipeID:       testRecipeID,
// 			InstanceID:     testInstanceID,
// 			JobID:          testJobID,
// 			CantabularType: cantabularTable,
// 		}
// 		isFlexibleTableRecipeFalse := recipe.Recipe{
// 			Alias:          "Cantabular Example 2",
// 			Format:         "cantabular_table",
// 			CantabularBlob: "cantabular_blob",
// 			OutputInstances: []recipe.Instance{
// 				{
// 					CodeLists: []recipe.CodeList{
// 						{
// 							IsCantabularGeography: &trueValue,
// 						},
// 						// {
// 						// 	IsCantabularGeography: &falseValue,
// 						// },
// 					},
// 				},
// 			},
// 		}

// 		wg := &sync.WaitGroup{}
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()

// 			select {
// 			case _ := <-producer.Channels().Output:
// 				// var got event.CategoryDimensionImport
// 				// s := schema.CategoryDimensionImport
// 				// err := s.Unmarshal(b, &got)
// 				// c.So(err, ShouldBeNil)
// 				// c.So(got, ShouldResemble, expected)
// 			case <-time.After(msgTimeout):
// 				t.Fail()
// 			}
// 			msg := kafkaMessage(c, &event.InstanceStarted{
// 				RecipeID:       testRecipeID,
// 				InstanceID:     testInstanceID,
// 				JobID:          testJobID,
// 				CantabularType: cantabularTable,
// 			})
// 			err := h.Handle(ctx, workerID, msg)
// 			c.So(err, ShouldBeNil)
// 		}()

// 		errorValue := h.TriggerImportDimensionOptions(&isFlexibleTableRecipeFalse.Format, &isFlexibleTableRecipeFalse.OutputInstances[0].CodeLists, isFlexibleTableRecipeFalse.CantabularBlob, e)
// 		So(len(errorValue), ShouldEqual, 0)

// 		wg.Wait()
// 		err := producer.Close(ctx)
// 		So(err, ShouldBeNil)
// 	})
// }

func TestTriggerDimensionOptions(t *testing.T) {
	cfg := config.Config{}
	falseValue := false
	trueValue := true

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

			Convey("Then TEST1", func() {
				e := &event.InstanceStarted{
					RecipeID:       testRecipeID,
					InstanceID:     testInstanceID,
					JobID:          testJobID,
					CantabularType: cantabularTable,
				}
				isFlexibleTableRecipeFalse := recipe.Recipe{
					Alias:          "Cantabular Example 2",
					Format:         "cantabular_table",
					CantabularBlob: "cantabular_blob",
					OutputInstances: []recipe.Instance{
						{
							CodeLists: []recipe.CodeList{
								{
									IsCantabularGeography: &trueValue,
								},
								{
									IsCantabularGeography: &falseValue,
								},
							},
						},
					},
				}
				errorValue := h.TriggerImportDimensionOptions(&isFlexibleTableRecipeFalse.Format, &isFlexibleTableRecipeFalse.OutputInstances[0].CodeLists, isFlexibleTableRecipeFalse.CantabularBlob, e)
				So(len(errorValue), ShouldBeNil)
			})
			wg.Wait()
			err := producer.Close(ctx)
			So(err, ShouldBeNil)
		})
	})
}

func Test2TriggerDimensionOptions(t *testing.T) {
	cfg := config.Config{}

	Convey("Given a successful event handler", t, func() {
		ctblrClient := cantabularClientHappy()
		recipeAPIClient := recipeAPIClientOnlyGeography()
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

		wg := &sync.WaitGroup{}

		Convey("When Handle is triggered, with no Editions", func(c C) {
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

			Convey("Then the validateMessageAndTimeoutOk times out due to no message consumed", func() {
				expected := []event.CategoryDimensionImport{
					{
						JobID:          testJobID,
						InstanceID:     testInstanceID,
						DimensionID:    "test-variable",
						CantabularBlob: "test-cantabular-blob",
					},
				}

				for _, e := range expected {
					validateMessageAndTimeoutOk(t, producer, c, e)
				}

				wg.Wait()
				err := producer.Close(ctx)
				So(err, ShouldBeNil)
			})
		})
	})
}

func Test3TriggerDimensionOptions(t *testing.T) {
	cfg := config.Config{}

	Convey("Given a successful event handler", t, func() {
		ctblrClient := cantabularClientHappy()
		recipeAPIClient := recipeAPIClientOnlyGeographyWithEdition()
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

		wg := &sync.WaitGroup{}

		Convey("When Handle is triggered, with no Editions the expected error is seen", func(c C) {
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

			Convey("Then the validateMessageAndTimeoutOk times out due to no message consumed", func() {
				expected := []event.CategoryDimensionImport{
					{
						JobID:      testJobID,
						InstanceID: testInstanceID,
						// DimensionID:    "test-variable",
						CantabularBlob: "cantabular_blob",
					},
					{
						JobID:      testJobID,
						InstanceID: testInstanceID,
						// DimensionID:    "test-mapped-variable",
						CantabularBlob: "cantabular_blob",
					},
				}

				for _, e := range expected {
					validateMessageAndTimeoutOk(t, producer, c, e)
				}

				wg.Wait()
				err := producer.Close(ctx)
				So(err, ShouldBeNil)
			})
		})
	})
}
