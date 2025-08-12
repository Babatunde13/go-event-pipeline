.PHONY: build-eventbridge-producer
build-eventbridge-producer:
# build-eventbridge-producer: build the event bridge producer application
	@echo "Building event bridge producer..."
	@export GO111MODULE=on
	@env GOARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o bin/eventbridge-producer/bootstrap cmd/eventbridge-producer/main.go

.PHONY: package-eventbridge-producer
package-eventbridge-producer: build-eventbridge-producer
## package: package the application for deployment
	zip -jr "bin/eventbridge-producer/bootstrap.zip" "bin/eventbridge-producer/bootstrap"

.PHONY: build-kafka-consumer
build-kafka-consumer:
# build-kafka-consumer: build the kafka consumer application
	@echo "Building kafka consumer..."
	@export GO111MODULE=on
	@env GOARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o bin/kafka-consumer/bootstrap cmd/kafka-consumer/main.go

.PHONY: package-kafka-consumer
package-kafka-consumer: build-kafka-consumer
## package: package the application for deployment
	zip -jr "bin/kafka-consumer/bootstrap.zip" "bin/kafka-consumer/bootstrap"


.PHONY: build-kafka-producer
build-kafka-producer:
# build-kafka-producer: build the kafka producer application
	@echo "Building kafka producer..."
	@export GO111MODULE=on
	@env GOARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o bin/kafka-producer/bootstrap cmd/kafka-producer/main.go

.PHONY: package-kafka-producer
package-kafka-producer: build-kafka-producer
## package: package the application for deployment
	zip -jr "bin/kafka-producer/bootstrap.zip" "bin/kafka-producer/bootstrap"

.PHONY: build-lambda-consumer
build-lambda-consumer:
# build-lambda-consumer: build the lambda consumer application
	@echo "Building lambda consumer..."
	@export GO111MODULE=on
	@env GOARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o bin/lambda-consumer/bootstrap cmd/lambda-consumer/main.go

.PHONY: package-lambda-consumer
package-lambda-consumer: build-lambda-consumer
## package: package the application for deployment
	zip -jr "bin/lambda-consumer/bootstrap.zip" "bin/lambda-consumer/bootstrap"

.PHONY: build-load-generator
build-load-generator:
# build-load-generator: build the load generator application
	@echo "Building load generator..."
	@export GO111MODULE=on
	@go build -o bin/load-generator/bootstrap cmd/load-generator/main.go

.PHONY: package-load-generator
package-load-generator: build-load-generator
## package: package the application for deployment
	zip -jr "bin/load-generator/bootstrap.zip" "bin/load-generator/bootstrap"

.PHONY: build
build: build-eventbridge-producer build-kafka-consumer build-kafka-producer build-lambda-consumer build-load-generator
## build: build all applications

.PHONY: package
package: package-eventbridge-producer package-kafka-consumer package-kafka-producer package-lambda-consumer package-load-generator
## package: package all applications

.PHONY: clean	
clean:
## clean: clean up
	rm -rf ./bin ./vendor go.sum

.PHONY: fmt-check
fmt-check:
	@echo "Checking formatting..."
	@go fmt ./... | awk '{if ($$0 != "") {print "Code not formatted"; exit 1}}'

.PHONY: fmt
fmt:
	@echo "Checking formatting..."
	@go fmt ./...

help:
	@echo "Available commands:"
	@echo "  make build-eventbridge-producer       Build the EventBridge producer application"
	@echo "  make package-eventbridge-producer     Package the EventBridge producer application"
	@echo "  make build-kafka-consumer             Build the Kafka consumer application"
	@echo "  make package-kafka-consumer           Package the Kafka consumer application"
	@echo "  make build-kafka-producer             Build the Kafka producer application"
	@echo "  make package-kafka-producer           Package the Kafka producer application"
	@echo "  make build-lambda-consumer            Build the Lambda consumer application"
	@echo "  make package-lambda-consumer          Package the Lambda consumer application"
	@echo "  make build                           Build all applications"
	@echo "  make package                         Package all applications"
	@echo "  make clean                           Clean up build artifacts"
	@echo "  make fmt-check                       Check code formatting"
	@echo "  make fmt                             Format the code"
	@echo "  make help                            Show this help message"