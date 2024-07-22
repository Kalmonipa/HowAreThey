package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"howarethey/pkg/models"
	"io"
	"net/http"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	git "github.com/go-git/go-git/v5"
)

var (
	imageName  string
	portNumber = "8080"

	// Mock friends list used in the tests
	mockFriendsList = models.FriendsList{
		models.Friend{ID: "1", Name: "John Wick", LastContacted: "2023-06-06", Birthday: "1996-02-23", Notes: "Nice guy"},
		models.Friend{ID: "2", Name: "Peter Parker", LastContacted: "2023-12-12", Birthday: "1996-02-23", Notes: "I think he's Spiderman"},
	}
)

// Helper Functions

func SetupTests() (*client.Client, context.Context, error) {
	// Create a Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, nil, err
	}
	defer cli.Close()

	ctx := context.Background()

	return cli, ctx, nil
}

func addFriend(mockFriend models.Friend) (statusCode int, body []byte, err error) {
	data := map[string]string{
		"Name":          mockFriend.Name,
		"LastContacted": mockFriend.LastContacted,
		"Birthday":      mockFriend.Birthday,
		"Notes":         mockFriend.Notes,
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return 0, nil, err
	}

	respStatus, respBody, err := performContainerRequest("POST", "/friends", payload)
	if err != nil {
		return 0, nil, err
	}

	return respStatus, respBody, nil
}

func getGitBranchName() (string, error) {
	repo, err := git.PlainOpen("../../..")
	if err != nil {
		return "", err
	}

	ref, err := repo.Head()
	if err != nil {
		return "", err
	}

	branchName := ref.Name().Short()

	return branchName, nil
}

func performContainerRequest(method, path string, body []byte) (respStatusCode int, respBody []byte, err error) {
	client := &http.Client{}

	fullPath := "http://localhost:" + portNumber + path

	req, err := http.NewRequest(method, fullPath, bytes.NewBuffer(body))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, respBody, nil
}

func startContainer(cli *client.Client, ctx context.Context) (response container.CreateResponse, branch string, err error) {
	branchName, err := getGitBranchName()
	if err != nil {
		return container.CreateResponse{}, "", err
	}

	imageName = "kalmonipa/howarethey:" + branchName

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        imageName,
		ExposedPorts: nat.PortSet{"8080": struct{}{}},
	}, &container.HostConfig{
		PortBindings: map[nat.Port][]nat.PortBinding{nat.Port(portNumber): {{HostIP: "127.0.0.1", HostPort: portNumber}}},
	}, nil, nil, branchName)
	if err != nil {
		return container.CreateResponse{}, "", err
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return container.CreateResponse{}, "", err
	}

	// Sleeping to give the webserver time to start up
	time.Sleep(1 * time.Second)

	return resp, branchName, nil
}

func stopContainer(cli *client.Client, ctx context.Context, containerID string) {
	err := cli.ContainerStop(ctx, containerID, container.StopOptions{})
	if err != nil {
		return
	}

	err = cli.ContainerRemove(ctx, containerID, container.RemoveOptions{})
	if err != nil {
		return
	}
}
