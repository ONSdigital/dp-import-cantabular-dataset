package steps

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-import-cantabular-dataset/event"
	"github.com/ONSdigital/dp-import-cantabular-dataset/schema"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"
	"github.com/ONSdigital/log.go/v2/log"

	"github.com/cucumber/godog"
	"github.com/google/go-cmp/cmp"
	"github.com/rdumont/assistdog"
)

func (c *Component) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the following recipe with id "([^"]*)" is available from dp-recipe-api:$`, c.theFollowingRecipeIsAvailable)
	ctx.Step(`^no recipe with id "([^"]*)" is available from dp-recipe-api`, c.theFollowingRecipeIsNotFound)
	ctx.Step(`^the following query response is available from Cantabular api extension for the dataset "([^"]*)" and variables "([^"]*)":$`, c.theFollowingCantabularVariablesAreAvailable)
	ctx.Step(`^the service starts`, c.theServiceStarts)
	ctx.Step(`^dp-dataset-api is healthy`, c.datasetAPIIsHealthy)
	ctx.Step(`^dp-dataset-api is unhealthy`, c.datasetAPIIsUnhealthy)
	ctx.Step(`^dp-recipe-api is healthy`, c.recipeAPIIsHealthy)
	ctx.Step(`^cantabular server is healthy`, c.cantabularServerIsHealthy)
	ctx.Step(`^cantabular api extension is healthy`, c.cantabularAPIExtIsHealthy)
	ctx.Step(`^the call to update job "([^"]*)" is succesful`, c.theCallToUpdateJobIsSuccessful)
	ctx.Step(`^the call to update instance "([^"]*)" is succesful`, c.theCallToUpdateInstanceIsSuccessful)
	ctx.Step(`^the call to update job "([^"]*)" is unsuccesful`, c.theCallToUpdateJobIsUnsuccessful)
	ctx.Step(`^the call to update instance "([^"]*)" is unsuccesful`, c.theCallToUpdateInstanceIsUnsuccessful)

	ctx.Step(`^this instance-started event is queued, to be consumed:$`, c.thisInstanceStartedEventIsQueued)
	ctx.Step(`^these category dimension import events should be produced:$`, c.theseCategoryDimensionImportEventsShouldBeProduced)
	ctx.Step(`^no category dimension import events should be produced`, c.noCategoryDimensionImportEventsShouldBeProduced)
}

// theServiceStarts starts the service under test in a new go-routine
// note that this step should be called only after all dependencies have been setup,
// to prevent any race condition, specially during the first healthcheck iteration.
func (c *Component) theServiceStarts() error {
	return c.startService(c.ctx)
}

// datasetAPIIsHealthy generates a mocked healthy response for dataset API healthecheck
func (c *Component) datasetAPIIsHealthy() error {
	const res = `{"status": "OK"}`
	c.DatasetAPI.NewHandler().
		Get("/health").
		Reply(http.StatusOK).
		BodyString(res)
	return nil
}

// datasetAPIIsUnhealthy generates a mocked unhealthy response for dataset API healthecheck
func (c *Component) datasetAPIIsUnhealthy() error {
	const res = `{"status": "CRITICAL"}`
	c.DatasetAPI.NewHandler().
		Get("/health").
		Reply(http.StatusInternalServerError).
		BodyString(res)
	return nil
}

// recipeAPIIsHealthy generates a mocked healthy response for recipe API healthecheck
func (c *Component) recipeAPIIsHealthy() error {
	const res = `{"status": "OK"}`
	c.RecipeAPI.NewHandler().
		Get("/health").
		Reply(http.StatusOK).
		BodyString(res)
	return nil
}

// cantabularServerIsHealthy generates a mocked healthy response for cantabular server
func (c *Component) cantabularServerIsHealthy() error {
	const res = `{"status": "OK"}`
	c.CantabularSrv.NewHandler().
		Get("/v10/datasets").
		Reply(http.StatusOK).
		BodyString(res)
	return nil
}

// cantabularAPIExtIsHealthy generates a mocked healthy response for cantabular server
func (c *Component) cantabularAPIExtIsHealthy() error {
	const res = `{"status": "OK"}`
	c.CantabularApiExt.NewHandler().
		Get("/graphql?query={}").
		Reply(http.StatusOK).
		BodyString(res)
	return nil
}

func (c *Component) theFollowingRecipeIsAvailable(id string, recipe *godog.DocString) error {
	c.RecipeAPI.NewHandler().
		Get("/recipes/" + id).
		Reply(http.StatusOK).
		BodyString(recipe.Content)

	return nil
}

func (c *Component) theFollowingRecipeIsNotFound(id string) error {
	c.RecipeAPI.NewHandler().
		Get("/recipes/" + id).
		Reply(http.StatusNotFound)

	return nil
}

func (c *Component) theFollowingCantabularVariablesAreAvailable(dataset string, variables string, cb *godog.DocString) error {
	// Parse the variables from csv format to string slice
	vars, err := csv.NewReader(strings.NewReader(variables)).Read()
	if err != nil {
		return fmt.Errorf("could not parse variables as comma separated values, %w", err)
	}

	// Encode the graphQL query with the provided dataset and variables
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	if err := enc.Encode(map[string]interface{}{
		"query": cantabular.QueryDimensionsByName,
		"variables": map[string]interface{}{
			"dataset":   dataset,
			"text":      "",
			"variables": vars,
			"category":  "",
			"limit":     20,
			"offset":    0,
		},
	}); err != nil {
		return fmt.Errorf("failed to encode GraphQL query: %w", err)
	}

	// create graphql handler with expected query body
	c.CantabularApiExt.NewHandler().
		Post("/graphql").
		AssertBody(b.Bytes()).
		Reply(http.StatusOK).
		BodyString(cb.Content)

	return nil
}

func (c *Component) theCallToUpdateInstanceIsSuccessful(instance string) error {
	c.DatasetAPI.NewHandler().
		Put("/instances/" + instance).
		Reply(http.StatusOK)

	return nil
}

func (c *Component) theCallToUpdateInstanceIsUnsuccessful(instance string) error {
	c.DatasetAPI.NewHandler().
		Put("/instances/" + instance).
		Reply(http.StatusInternalServerError)

	return nil
}

func (c *Component) theCallToUpdateJobIsSuccessful(job string) error {
	c.ImportAPI.NewHandler().
		Put("/jobs/" + job).
		Reply(http.StatusOK)

	return nil
}

func (c *Component) theCallToUpdateJobIsUnsuccessful(job string) error {
	c.ImportAPI.NewHandler().
		Put("/jobs/" + job).
		Reply(http.StatusInternalServerError)

	return nil
}

func (c *Component) theseCategoryDimensionImportEventsShouldBeProduced(events *godog.Table) error {
	expected, err := assistdog.NewDefault().CreateSlice(new(event.CategoryDimensionImport), events)
	if err != nil {
		return fmt.Errorf("failed to create slice from godog table: %w", err)
	}

	var got []*event.CategoryDimensionImport
	listen := true

	for listen {
		select {
		case <-time.After(WaitEventTimeout):
			listen = false
		case <-c.consumer.Channels().Closer:
			return errors.New("closer channel closed")
		case msg, ok := <-c.consumer.Channels().Upstream:
			if !ok {
				return errors.New("upstream channel closed")
			}

			var e event.CategoryDimensionImport
			var s = schema.CategoryDimensionImport

			if err := s.Unmarshal(msg.GetData(), &e); err != nil {
				msg.Commit()
				msg.Release()
				return fmt.Errorf("error unmarshalling message: %w", err)
			}

			msg.Commit()
			msg.Release()

			got = append(got, &e)
		}
	}

	if diff := cmp.Diff(got, expected); diff != "" {
		return fmt.Errorf("-got +expected)\n%s\n", diff)
	}

	return nil
}

func (c *Component) noCategoryDimensionImportEventsShouldBeProduced() error {
	listen := true

	for listen {
		select {
		case <-time.After(WaitEventTimeout):
			listen = false
		case <-c.consumer.Channels().Closer:
			return errors.New("closer channel closed")
		case msg, ok := <-c.consumer.Channels().Upstream:
			if !ok {
				return errors.New("upstream channel closed")
			}

			err := fmt.Errorf("unexpected message receieved: %s", msg.GetData())

			msg.Commit()
			msg.Release()

			return err
		}
	}

	return nil
}

func (c *Component) thisInstanceStartedEventIsQueued(input *godog.DocString) error {
	ctx := context.Background()

	// testing kafka message that will be produced
	var testEvent event.InstanceStarted
	if err := json.Unmarshal([]byte(input.Content), &testEvent); err != nil {
		return fmt.Errorf("error unmarshaling input to event: %w body: %s", err, input.Content)
	}

	log.Info(ctx, "event to marshal: ", log.Data{
		"event": testEvent,
	})

	// marshal and send message
	b, err := schema.InstanceStarted.Marshal(testEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal event from schema: %w", err)
	}

	log.Info(ctx, "marshalled event: ", log.Data{
		"event": b,
	})

	msg := kafkatest.NewMessage(b, 1)
	c.producer.Channels().Output <- msg.GetData()

	log.Info(ctx, "thisInstanceStartedEventIsConsumed done")
	return nil
}
