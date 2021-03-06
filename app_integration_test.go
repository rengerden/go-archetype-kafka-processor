// These tests make use of Docker, spinning up an Apache Kafka container for integration level tests
// these tests do not test that the application is packaged correctly after a Docker build
package main

import (
	"encoding/json"
	"fmt"
	"github.com/adbourne/go-archetype-kafka-processor/config"
	"github.com/adbourne/go-archetype-kafka-processor/messages"
	"github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/assert"
	"gopkg.in/Shopify/sarama.v1"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"
)

const (
	DockerImageNameKafka   = "spotify/kafka:latest"
	KafkaPort              = 9092
	ServerPort             = 8080
	IntegrationSourceTopic = "source-topic"
	IntegrationSinkTopic   = "sink-topic"
)

func TestHealthCheckEndpointReturnsCorrectResponse(t *testing.T) {
	withKafka(t, func() {
		testHTTPGetRequest(t, fmt.Sprintf("http://localhost:%d/health", ServerPort), `{"status":"ok"}`)
	})
}

func TestMessageIsProcessedCorrectly(t *testing.T) {
	withKafka(t, func() {
		brokers := []string{"localhost:9092"}
		// Send source message
		saramaConfig := sarama.NewConfig()
		asyncProducer, _ := sarama.NewAsyncProducer(brokers, saramaConfig)
		asyncProducer.Input() <- &sarama.ProducerMessage{
			Topic: SourceTopic,
			Key:   nil,
			Value: &messages.SourceMessage{
				Seed: 1,
			},
		}

		// Wait for it to be processed
		time.Sleep(3 * time.Second)

		// Check the result
		consumer, _ := sarama.NewConsumer(brokers, nil)
		partitionConsumer, _ := consumer.ConsumePartition(SinkTopic, 0, 0)
		sinkMessage := <-partitionConsumer.Messages()

		sinkMessageValue := &messages.SinkMessage{}
		json.Unmarshal(sinkMessage.Value, sinkMessageValue)

		assert.Equal(t, sinkMessageValue.RandomNumber, 5577006791947779410)
	})
}

// Sends a HTTP GET request to the server and asserts the response is as expected
func testHTTPGetRequest(t *testing.T, url string, expectedBody string) {
	resp, err := http.Get(url)
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, ioerr := ioutil.ReadAll(resp.Body)
	assert.NoError(t, ioerr)
	assert.Equal(t, expectedBody, string(body), "Response body was not correct")
	assert.Equal(t, "200 OK", resp.Status, "HTTP status was not 200")
}

// Starts and manages an Apache Kafka docker container for the duration of the test
func withKafka(t *testing.T, testFunc func()) {

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// Create Docker client
	dockerClient := createDockerClient(t)

	// Start Kafka container
	kafkaContainer := createDockerContainer(t, dockerClient, kafkaCreateOptions())
	log.Println(fmt.Sprintf("Kafka container '%s' started", kafkaContainer.ID))
	defer func() {
		if err := dockerClient.RemoveContainer(docker.RemoveContainerOptions{
			ID:    kafkaContainer.ID,
			Force: true,
		}); err != nil {
			t.Fatalf("cannot remove container: %s", err)
		}
	}()

	// Wait for container to wake up
	waitStarted(t, dockerClient, kafkaContainer.ID, 5*time.Second)

	// Inspect the docker container for additional information
	kafkaContainer, err := dockerClient.InspectContainer(kafkaContainer.ID)
	if err != nil {
		t.Fatalf("Couldn't inspect container: %s", err)
	}

	// Wait for kafka to be available
	time.Sleep(15 * time.Second) // TODO: Make this an implicit wait

	// Setup and run the app
	appConfig := &config.AppConfig{
		Port:        ServerPort,
		RandomSeed:  int64(1),
		Brokers:     fmt.Sprintf("localhost:%d", KafkaPort),
		SourceTopic: IntegrationSourceTopic,
		SinkTopic:   IntegrationSinkTopic,
	}

	appContext := NewAppContext(*appConfig)

	go func() {
		RunApp(appContext)
	}()

	// Wait for the server to start up
	time.Sleep(1 * time.Second)

	// Run the test
	testFunc()
}

// createDockerClient creates a Docker client
func createDockerClient(t *testing.T) *docker.Client {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		t.Fatalf("Cannot connect to Docker daemon: %s", err)
	}
	return client
}

// createContainer creates a Docker container
func createDockerContainer(t *testing.T, client *docker.Client, createOptions docker.CreateContainerOptions) *docker.Container {
	c, err := client.CreateContainer(createOptions)
	if err != nil {
		t.Fatalf("Cannot create Docker container: %s", err)
	}

	err = client.StartContainer(c.ID, &docker.HostConfig{})
	if err != nil {
		t.Fatalf("Cannot start Docker container: %s", err)
	}
	return c
}

// kafkaCreateOptions creates CreateContainerOptions for an Apache Kafka docker container
func kafkaCreateOptions() docker.CreateContainerOptions {
	envVars := []string{
		"ADVERTISED_HOST=localhost",
	}

	exposedPorts := map[docker.Port]struct{}{"9092/tcp": {}}

	portBindings := map[docker.Port][]docker.PortBinding{
		"9092/tcp": {{HostIP: "0.0.0.0", HostPort: "9092"}}}

	hostConfig := docker.HostConfig{
		PortBindings:    portBindings,
		PublishAllPorts: true,
		Privileged:      false,
	}

	opts := docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:        DockerImageNameKafka,
			ExposedPorts: exposedPorts,
			Env:          envVars,
		},
		HostConfig: &hostConfig,
	}

	return opts
}

// waitStarted waits for a container to start for the maxWait time.
func waitStarted(t *testing.T, client *docker.Client, id string, maxWait time.Duration) {
	done := time.Now().Add(maxWait)
	for time.Now().Before(done) {
		c, err := client.InspectContainer(id)
		if err != nil {
			return
		}
		if c.State.Running {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf(fmt.Sprintf("Cannot start container '%s' for '%v', aborting", id, maxWait))
}
