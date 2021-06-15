module github.com/ONSdigital/dp-import-cantabular-dataset

go 1.16

replace github.com/coreos/etcd => github.com/coreos/etcd v3.3.24+incompatible
replace github.com/ONSdigital/dp-api-clients-go => /home/jpm/go-modules/onsdigital/dp-api-clients-go

require (
	github.com/ONSdigital/dp-api-clients-go v1.37.0
	github.com/ONSdigital/dp-component-test v0.3.1
	github.com/ONSdigital/dp-healthcheck v1.0.5
	github.com/ONSdigital/dp-import v1.1.0
	github.com/ONSdigital/dp-kafka/v2 v2.2.0
	github.com/ONSdigital/dp-net v1.0.12
	github.com/ONSdigital/go-ns v0.0.0-20210410105122-6d6a140e952e
	github.com/ONSdigital/log.go v1.0.1
	github.com/ONSdigital/log.go/v2 v2.0.0
	github.com/cucumber/godog v0.11.0
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-memdb v1.3.1 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/rdumont/assistdog v0.0.0-20201106100018-168b06230d14
	github.com/smartystreets/goconvey v1.6.4
	github.com/stretchr/testify v1.7.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
)
