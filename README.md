dp-import-cantabular-dataset
================
Begins the process of retrieving categories frpm Catabular API to being the population of dimensions into our datastore


### Getting started

* Run `make debug`

The service runs in the background consuming messages from Kafka.
An example event can be created using the helper script, `make produce`.

### Dependencies

* Requires running…
  * [kafka](https://github.com/ONSdigital/dp/blob/main/guides/INSTALLING.md#prerequisites)
* No further dependencies other than those defined in `go.mod`

### Docker-Compose ###

Alternatively this service can be brought up with all it's dependencies as part of the
Cantabular dataset import journey with the docker-compose environment found at:
https://github.com/ONSdigital/dp-compose/cantabular-import

### Running Tests ###

## Unit Tests ##

* `make test`

## Component Tests ##

* `sudo make test-component`

### Configuration

| Environment variable                         | Default                                      | Description
| ----------------------------                 | ---------------------------------            | -----------
| BIND_ADDR                                    | localhost:26100                              | The host and port to bind to
| GRACEFUL_SHUTDOWN_TIMEOUT                    | 5s                                           | The graceful shutdown timeout in seconds (`time.Duration` format)
| HEALTHCHECK_INTERVAL                         | 30s                                          | Time between self-healthchecks (`time.Duration` format)
| HEALTHCHECK_CRITICAL_TIMEOUT                 | 90s                                          | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format)
| KAFKA_ADDR                                   | localhost:9092                               | The kafka broker addresses (can be comma separated)
| KAFKA_VERSION                                | "1.0.2"                                      | The kafka version that this service expects to connect to
| KAFKA_OFFSET_OLDEST                          | true                                         | Start processing Kafka messages in order from the oldest in the queue
| KAFKA_NUM_WORKERS                            | 1                                            | The maximum number of parallel kafka consumers
| KAFKA_DATASET_INSTANCE_STARTED_GROUP         | dp-import-cantabular-dataset                 | The consumer group this application to consume ImageUploaded messages
| KAFKA_DATASET_INSTANCE_STARTED_TOPIC         | cantabular-dataset-instance-started          | The name of the topic to consume messages from
| CANTABULAR_DATASET_CATEGORY_DIMENSION_IMPORT | cantabular-dataset-category-dimension-import | The name of the topic to produce messages to
| DATASET_API_URL                              | http://localhost:22000                       | HOST URL for dp-dataset-api
| RECIPE_API_URL                               | http://localhost:22300                       | HOST URL for dp-recipe-api
| IMPORT_API_URL                               | http://localhost:21800                       | HOST URL for dp-import-api
| CANTABULAR_URL                               | http://localhost:8491                        | HOST URL for dp-cantabular-server
| SERVICE_AUTH_TOKEN                           | ""                                           | Service auth token for authorizing requests
| COMPONENT_TEST_USE_LOG_FILE                  | false                                        | Output component test logs to temporary file instead of stdout. Used for displaying output for component tests in Concourse.
| KAFKA_SEC_PROTO                              | _unset_                                      | if set to `TLS`, kafka connections will use TLS [[1]](#notes_1)
| KAFKA_SEC_CA_CERTS                           | _unset_                                      | CA cert chain for the server cert [[1]](#notes_1)
| KAFKA_SEC_CLIENT_KEY                         | _unset_                                      | PEM for the client key [[1]](#notes_1)
| KAFKA_SEC_CLIENT_CERT                        | _unset_                                      | PEM for the client certificate [[1]](#notes_1)
| KAFKA_SEC_SKIP_VERIFY                        | false                                        | ignores server certificate issues if `true` [[1]](#notes_1)

**Notes:**

    1. <a name="notes_1">For more info, see the [kafka TLS examples documentation](https://github.com/ONSdigital/dp-kafka/tree/main/examples#tls)</a>

### Healthcheck

 The `/health` endpoint returns the current status of the service. Dependent services are health checked on an interval defined by the `HEALTHCHECK_INTERVAL` environment variable.

 On a development machine a request to the health check endpoint can be made by:

 `curl localhost:8125/health`

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2021, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
