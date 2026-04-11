package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

var shortcutURL string

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	ctx := context.Background()

	net, err := network.New(ctx)
	if err != nil {
		fmt.Printf("failed to create network: %v\n", err)
		return 1
	}
	defer func() { _ = net.Remove(ctx) }()

	mongo, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:    "mongo:8",
			Networks: []string{net.Name},
			NetworkAliases: map[string][]string{
				net.Name: {"mongo"},
			},
			WaitingFor: wait.ForLog("Waiting for connections"),
		},
		Started: true,
	})
	if err != nil {
		fmt.Printf("failed to start mongo: %v\n", err)
		return 1
	}
	defer func() { _ = mongo.Terminate(ctx) }()

	mockService, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context:       "../../",
				Dockerfile:    "tests/mock-service/Dockerfile",
				PrintBuildLog: true,
			},
			Networks: []string{net.Name},
			NetworkAliases: map[string][]string{
				net.Name: {"mock-service"},
			},
			ExposedPorts: []string{"9001/tcp"},
			WaitingFor: wait.ForHTTP("/health").
				WithPort("9001/tcp").
				WithStartupTimeout(3 * time.Minute),
		},
		Started: true,
	})
	if err != nil {
		fmt.Printf("failed to start mock-service: %v\n", err)
		return 1
	}
	defer func() { _ = mockService.Terminate(ctx) }()

	shortcut, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context:       "../../",
				Dockerfile:    "Dockerfile",
				PrintBuildLog: true,
			},
			Networks:     []string{net.Name},
			ExposedPorts: []string{"8080/tcp"},
			Env:          map[string]string{"ENV": "testing"},
			WaitingFor: wait.ForHTTP("/health").
				WithPort("8080/tcp").
				WithStartupTimeout(3 * time.Minute),
		},
		Started: true,
	})
	if err != nil {
		fmt.Printf("failed to start shortcut: %v\n", err)
		return 1
	}
	defer func() { _ = shortcut.Terminate(ctx) }()

	host, err := shortcut.Host(ctx)
	if err != nil {
		fmt.Printf("failed to get shortcut host: %v\n", err)
		return 1
	}
	port, err := shortcut.MappedPort(ctx, "8080/tcp")
	if err != nil {
		fmt.Printf("failed to get shortcut port: %v\n", err)
		return 1
	}

	shortcutURL = fmt.Sprintf("http://%s:%s", host, port.Port())

	return m.Run()
}
