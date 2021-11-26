package steps

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-import-cantabular-dataset/event"
	"github.com/ONSdigital/dp-import-cantabular-dataset/schema"
	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	"github.com/ONSdigital/log.go/v2/log"

	"github.com/cucumber/godog"
	"github.com/google/go-cmp/cmp"
	"github.com/rdumont/assistdog"
)

func (c *Component) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the following recipe with id "([^"]*)" is available from dp-recipe-api:$`, c.theFollowingRecipeIsAvailable)
	ctx.Step(`^no recipe with id "([^"]*)" is available from dp-recipe-api`, c.theFollowingRecipeIsNotFound)
	ctx.Step(`^the following response is available from Cantabular from the codebook "([^"]*)" and query "([^"]*)":$`, c.theFollowingCodebookIsAvailable)

	ctx.Step(`^the call to update job "([^"]*)" is succesful`, c.theCallToUpdateJobIsSuccessful)
	ctx.Step(`^the call to update instance "([^"]*)" is succesful`, c.theCallToUpdateInstanceIsSuccessful)
	ctx.Step(`^the call to update job "([^"]*)" is unsuccesful`, c.theCallToUpdateJobIsUnsuccessful)
	ctx.Step(`^the call to update instance "([^"]*)" is unsuccesful`, c.theCallToUpdateInstanceIsUnsuccessful)

	ctx.Step(`^this instance-started event is consumed:$`, c.thisInstanceStartedEventIsConsumed)
	ctx.Step(`^these category dimension import events should be produced:$`, c.theseCategoryDimensionImportEventsShouldBeProduced)
	ctx.Step(`^no category dimension import events should be produced`, c.noCategoryDimensionImportEventsShouldBeProduced)
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

func (c *Component) theFollowingCodebookIsAvailable(name, q string, cb *godog.DocString) error {
	c.CantabularSrv.NewHandler().
		Get("/v9/codebook/" + name + q).
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
		case <-time.After(time.Second * 1):
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
		case <-time.After(time.Second * 1):
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

func (c *Component) thisInstanceStartedEventIsConsumed(input *godog.DocString) error {
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
