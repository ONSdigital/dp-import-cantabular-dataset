module github.com/ONSdigital/dp-import-cantabular-dataset

go 1.16

replace github.com/coreos/etcd => github.com/coreos/etcd v3.3.24+incompatible

require (
	github.com/ONSdigital/dp-api-clients-go/v2 v2.1.7-beta
	github.com/ONSdigital/dp-component-test v0.3.1
	github.com/ONSdigital/dp-healthcheck v1.0.5
	github.com/ONSdigital/dp-import v1.1.0
	github.com/ONSdigital/dp-import-api v1.16.0
	github.com/ONSdigital/dp-kafka/v2 v2.2.0
	github.com/ONSdigital/dp-net v1.0.12
	github.com/ONSdigital/log.go v1.0.1
	github.com/ONSdigital/log.go/v2 v2.0.5
	github.com/Shopify/sarama v1.29.1 // indirect
	github.com/cucumber/godog v0.11.0
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/google/go-cmp v0.5.6
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-memdb v1.3.1 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/maxcnunes/httpfake v1.2.1
	github.com/rdumont/assistdog v0.0.0-20201106100018-168b06230d14
	github.com/smartystreets/goconvey v1.6.4
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
)
