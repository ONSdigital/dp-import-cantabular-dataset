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

### Configuration

| Environment variable                         | Default                                      | Description
| ----------------------------                 | ---------------------------------            | -----------
| BIND_ADDR                                    | localhost:26100                              | The host and port to bind to
| GRACEFUL_SHUTDOWN_TIMEOUT                    | 5s                                           | The graceful shutdown timeout in seconds (`time.Duration` format)
| HEALTHCHECK_INTERVAL                         | 30s                                          | Time between self-healthchecks (`time.Duration` format)
| HEALTHCHECK_CRITICAL_TIMEOUT                 | 90s                                          | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format)
| KAFKA_ADDR                                   | "localhost:9092"                             | The address of Kafka (accepts list)
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

### Healthcheck

 The `/health` endpoint returns the current status of the service. Dependent services are health checked on an interval defined by the `HEALTHCHECK_INTERVAL` environment variable.

 On a development machine a request to the health check endpoint can be made by:

 `curl localhost:8125/health`

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2021, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.

