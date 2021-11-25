package event_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/importapi"
	"github.com/ONSdigital/dp-import-cantabular-dataset/config"
	"github.com/ONSdigital/dp-import-cantabular-dataset/event"
	"github.com/ONSdigital/dp-import-cantabular-dataset/event/mock"
	"github.com/ONSdigital/dp-import-cantabular-dataset/schema"

	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	testCtx    = context.Background()
	errHandler = errors.New("handler Error")
)

var testEvent = event.InstanceStarted{
	RecipeID:       "test-recipe-id",
	InstanceID:     "test-instance-id",
	JobID:          "test-job-id",
	CantabularType: "cantabular_table",
}

// kafkaStubConsumer mock which exposes Channels function returning empty channels
// to be used on tests that are not supposed to receive any kafka message
var kafkaStubConsumer = &kafkatest.IConsumerGroupMock{
	ChannelsFunc: func() *kafka.ConsumerGroupChannels {
		return &kafka.ConsumerGroupChannels{}
	},
}

func TestConsume(t *testing.T) {
	cfg, err := config.Get()
	if err != nil {
		t.Fatalf("failed to get config: %s", err)
	}

	var proc = event.NewProcessor(*cfg, importAPIClientHappy(), datasetAPIClientHappy())

	Convey("Given kafka consumer and event handler mocks", t, func() {
		cgChannels := &kafka.ConsumerGroupChannels{Upstream: make(chan kafka.Message, 2)}
		mockConsumer := &kafkatest.IConsumerGroupMock{
			ChannelsFunc: func() *kafka.ConsumerGroupChannels { return cgChannels },
		}

		handlerWg := &sync.WaitGroup{}
		mockEventHandler := &mock.HandlerMock{
			HandleFunc: func(ctx context.Context, event *event.InstanceStarted) error {
				defer handlerWg.Done()
				return nil
			},
		}

		Convey("And a kafka message with the valid schema being sent to the Upstream channel", func() {
			message := kafkatest.NewMessage(marshal(testEvent), 0)
			mockConsumer.Channels().Upstream <- message

			Convey("When consume message is called", func() {
				handlerWg.Add(1)
				proc.Consume(testCtx, mockConsumer, mockEventHandler)
				handlerWg.Wait()

				Convey("An event is sent to the mockEventHandler ", func() {
					So(len(mockEventHandler.HandleCalls()), ShouldEqual, 1)
					So(*mockEventHandler.HandleCalls()[0].InstanceStarted, ShouldResemble, testEvent)
				})

				Convey("The message is committed and the consumer is released", func() {
					<-message.UpstreamDone()
					So(len(message.CommitCalls()), ShouldEqual, 1)
					So(len(message.ReleaseCalls()), ShouldEqual, 1)
				})
			})
		})

		Convey("And two kafka messages, one with a valid schema and one with an invalid schema", func() {
			validMessage := kafkatest.NewMessage(marshal(testEvent), 1)
			invalidMessage := kafkatest.NewMessage([]byte("invalid schema"), 0)
			mockConsumer.Channels().Upstream <- invalidMessage
			mockConsumer.Channels().Upstream <- validMessage

			Convey("When consume messages is called", func() {

				handlerWg.Add(1)
				proc.Consume(testCtx, mockConsumer, mockEventHandler)
				handlerWg.Wait()

				Convey("Only the valid event is sent to the mockEventHandler ", func() {
					So(len(mockEventHandler.HandleCalls()), ShouldEqual, 1)
					So(*mockEventHandler.HandleCalls()[0].InstanceStarted, ShouldResemble, testEvent)
				})

				Convey("Only the valid message is committed, but the consumer is released for both messages", func() {
					<-validMessage.UpstreamDone()
					<-invalidMessage.UpstreamDone()
					So(len(validMessage.CommitCalls()), ShouldEqual, 1)
					So(len(invalidMessage.CommitCalls()), ShouldEqual, 1)
					So(len(validMessage.ReleaseCalls()), ShouldEqual, 1)
					So(len(invalidMessage.ReleaseCalls()), ShouldEqual, 1)
				})
			})
		})

		Convey("With a failing handler and a kafka message with the valid schema being sent to the Upstream channel", func() {
			mockEventHandler.HandleFunc = func(ctx context.Context, event *event.InstanceStarted) error {
				defer handlerWg.Done()
				return errHandler
			}

			message := kafkatest.NewMessage(marshal(testEvent), 0)
			mockConsumer.Channels().Upstream <- message

			Convey("When consume message is called", func() {
				handlerWg.Add(1)
				proc.Consume(testCtx, mockConsumer, mockEventHandler)
				handlerWg.Wait()

				Convey("An event is sent to the mockEventHandler ", func() {
					So(len(mockEventHandler.HandleCalls()), ShouldEqual, 1)
					So(*mockEventHandler.HandleCalls()[0].InstanceStarted, ShouldResemble, testEvent)
				})

				Convey("The message is committed and the consumer is released", func() {
					<-message.UpstreamDone()
					So(len(message.CommitCalls()), ShouldEqual, 1)
					So(len(message.ReleaseCalls()), ShouldEqual, 1)
				})

			})

			Convey("With also broken dataset-api and import-api clients", func() {
				proc = event.NewProcessor(*cfg, importAPIClientUnhappy(), datasetAPIClientUnhappy())

				Convey("When consume message is called", func() {
					handlerWg.Add(1)
					proc.Consume(testCtx, mockConsumer, mockEventHandler)
					handlerWg.Wait()

					Convey("An event is sent to the mockEventHandler ", func() {
						So(len(mockEventHandler.HandleCalls()), ShouldEqual, 1)
						So(*mockEventHandler.HandleCalls()[0].InstanceStarted, ShouldResemble, testEvent)
					})

					Convey("The message is committed and the consumer is released", func() {
						<-message.UpstreamDone()
						So(len(message.CommitCalls()), ShouldEqual, 1)
						So(len(message.ReleaseCalls()), ShouldEqual, 1)
					})
				})
			})
		})
	})
}

// marshal helper method to marshal a event into a []byte
func marshal(event event.InstanceStarted) []byte {
	s := schema.InstanceStarted

	bytes, err := s.Marshal(event)
	So(err, ShouldBeNil)
	return bytes
}

func datasetAPIClientHappy() *mock.DatasetAPIClientMock {
	return &mock.DatasetAPIClientMock{
		PutInstanceStateFunc: func(ctx context.Context, uaToken, instanceID string, s dataset.State, mi string) (string, error) {
			return "", nil
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

func datasetAPIClientUnhappy() *mock.DatasetAPIClientMock {
	return &mock.DatasetAPIClientMock{
		PutInstanceStateFunc: func(ctx context.Context, uaToken, instanceID string, s dataset.State, mi string) (string, error) {
			return "", errors.New("invalid instance id")
		},
	}
}

func importAPIClientUnhappy() *mock.ImportAPIClientMock {
	return &mock.ImportAPIClientMock{
		UpdateImportJobStateFunc: func(ctx context.Context, uaToken string, jobID string, state importapi.State) error {
			return errors.New("invalid state")
		},
	}
}
