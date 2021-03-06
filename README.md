[![Build Status](https://travis-ci.org/adbourne/go-archetype-kafka-processor.svg?branch=master)](https://travis-ci.org/adbourne/go-archetype-kafka-processor)
[![codecov](https://codecov.io/gh/adbourne/go-archetype-kafka-processor/branch/master/graph/badge.svg)](https://codecov.io/gh/adbourne/go-archetype-kafka-processor)
[![Go Report Card](https://goreportcard.com/badge/github.com/adbourne/go-archetype-kafka-processor)](https://goreportcard.com/report/github.com/adbourne/go-archetype-kafka-processor)
[![GoDoc](https://godoc.org/github.com/adbourne/go-archetype-kafka-processor?status.svg)](https://godoc.org/github.com/adbourne/go-archetype-kafka-processor)

# go-archetype-kafka-processor

A Golang project archetype for a Kafka Processor.

What's a Kafka Processor? It's a headless service that consumes messages from an [Apache Kafka](https://kafka.apache.org/)
topic, processes the information in some way and then places the resulting message onto another Kafka topic. In this
instance the input message contains a seed number, the processing is generation of a random number using the seed and the resulting
message contains the random number. Combining services such as this together is a trait of a [Reactive System](http://www.reactivemanifesto.org/).

Why Golang? It's richly-featured, relatively simple to learn, produces tiny binaries, and on average consumes
far less memory than a JVM application. This final point, in particular, makes running a system of such services less
resource intensive and if you're running in the cloud, easier on your bank account.

## Dependency Injection
Golang makes [dependency injection](https://martinfowler.com/articles/injection.html) relatively straightforward
out of the box. This is mainly due to the fact that structs do not have to declare they implement an interface,
they simply have to have the correct function signatures that interface requires.

The project has a structure similar to that of a [Spring Boot](https://projects.spring.io/spring-boot/) application,
but with the concrete implementations declared in `application_context.go` in the root of the project.

## Health Check Endpoint
The application responds on a GET requests to `/health` with the following:
```
  {
    "status" : "ok"
  }
```
This is a very simple health check endpoint which can be used by Docker or container orchestrators, such as [Kubernetes](https://kubernetes.io/)
to signal that the service is available.

## Configuration
The application sources configuration from a set of environment variables. This compliments running the application in a Docker
container rather nicely. For the sake of the archetype, each value has a default value.

| Environment Variable                      | Purpose                                         | Default Value  |
| -----                                     | -----                                           | -----          |
| `KAFKA_PROCESSOR_ARCHETYPE_PORT`          | The port to run the HTTP health-check on        | 8080           |
| `KAFKA_PROCESSOR_ARCHETYPE_RANDOM_SEED`   | The seed to use when generating a random number | 1              |
| `KAFKA_PROCESSOR_ARCHETYPE_KAFKA_BROKERS` | The comma separated list of Kafka brokers       | localhost:9092 |
| `KAFKA_PROCESSOR_ARCHETYPE_SOURCE_TOPIC`  | The topic to get messages from                  | source-topic   |
| `KAFKA_PROCESSOR_ARCHETYPE_SINK_TOPIC`    | The topic to place messages onto                | sink-topic     |

## Building
This project uses [make](https://www.gnu.org/software/make/) for building the project. The following make tasks exist:
```
# Clean built artefacts
make clean

# Build the project
make build

# Run the tests
make test

# Package as a Docker container
make package
```

In addition to performing a `go install`, building the project also runs [gofmt](https://golang.org/cmd/gofmt/), [govet](https://golang.org/cmd/vet/), [gocyclo](https://github.com/fzipp/gocyclo), [golint](https://github.com/golang/lint), [ineffassign](https://github.com/gordonklaus/ineffassign) and [misspell](https://github.com/client9/misspell). The build does not fail if these tools report an error; it is up to the Developer if they take the advice. The vendor folder is excluded from the analysis.


## Testing
The testing library of choice for this project is [Testify](https://github.com/stretchr/testify). Testify is a sturdy, assert style framework. 

In addition to usint tests, this project has integration tests that run against a real instance of Kafka by using the [spotify/kafka](https://hub.docker.com/r/spotify/kafka/) Docker image. The kafka instance is spun up fresh for each integration test and then stopped after.

## Docker
The application can be packaged as a Docker container by using the following command:

`make package `

The container is built on [Alpine Linux](https://alpinelinux.org/) and works out to be about 14MB. The container can be
run using the following command:

`docker run -it -p8080:8080 adbourne/go-archetype-kafa-processor`

## Dependency Management
This project uses [Govendor](https://github.com/kardianos/govendor).
