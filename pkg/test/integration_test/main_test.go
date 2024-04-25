package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"howarethey/pkg/models"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	git "github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
)

var (
	imageName  string
	portNumber = "8080"

	// Mock friends list used in the tests
	mockFriendsList = models.FriendsList{
		models.Friend{ID: "1", Name: "John Wick", LastContacted: "06/06/2023", Birthday: "23/02/1996", Notes: "Nice guy"},
		models.Friend{ID: "2", Name: "Peter Parker", LastContacted: "12/12/2023", Birthday: "23/02/1996", Notes: "I think he's Spiderman"},
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

func addFriend() (statusCode int, body []byte, err error) {
	mockFriend := mockFriendsList[0]

	data := map[string]string{"Name": mockFriend.Name, "LastContacted": mockFriend.LastContacted}
	payload, err := json.Marshal(data)
	if err != nil {
		return 0, nil, err
	}

	respStatus, respBody, err := performRequest("POST", "/friends", payload)
	if err != nil {
		return 0, nil, err
	}

	return respStatus, respBody, nil
}

func getGitBranchName() (string, error) {
	// Open the repository in the current directory
	repo, err := git.PlainOpen("../../..")
	if err != nil {
		return "", err
	}

	// Retrieve the HEAD reference
	ref, err := repo.Head()
	if err != nil {
		return "", err
	}

	// Get the branch name from the reference
	branchName := ref.Name().Short()

	return branchName, nil
}

func performRequest(method, path string, body []byte) (respStatusCode int, respBody []byte, err error) {
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

func startContainer(cli *client.Client, ctx context.Context) (container.CreateResponse, string, error) {
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
	time.Sleep(2 * time.Second)

	return resp, branchName, nil
}

// Actual Test functions

// This test is redundant because we wouldn't have got this far if the daemon wasn't running
func TestDockerDaemonRunning(t *testing.T) {
	cli, _, err := SetupTests()
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	// Ping the Docker daemon
	_, err = cli.Ping(context.Background())
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}

func TestDockerContainerRunning(t *testing.T) {
	cli, ctx, err := SetupTests()
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	resp, branchName, err := startContainer(cli, ctx)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	containerState, err := cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	assert.Equal(t, "/"+branchName, containerState.ContainerJSONBase.Name) // Apparently the container name has a leading /
	assert.Equal(t, "running", containerState.ContainerJSONBase.State.Status)

}

func TestAddFriend(t *testing.T) {
	respStatus, respBody, err := addFriend()
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	expectedResponse := "{\"message\":\"John Wick added successfully\"}"

	assert.Equal(t, expectedResponse, string(respBody))

	assert.Equal(t, http.StatusCreated, respStatus)
}

func TestDeleteFriend(t *testing.T) {
	respStatus, respBody, err := performRequest("DELETE", "/friends/1", nil)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	var respJson map[string]string
	err = json.Unmarshal(respBody, &respJson)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, respStatus)
	assert.Equal(t, "John Wick removed successfully", respJson["message"])
}

// TODO: GET /birthdays
func TestBirthdays(t *testing.T) {
	assert.Equal(t, true, true)
}

// TODO: GET /friends/random
func TestGetRandomFriend(t *testing.T) {
	assert.Equal(t, true, true)
}

func TestGetFriendCount(t *testing.T) {
	_, _, err := addFriend()
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	statusCode, body, err := performRequest("GET", "/friends/count", nil)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "1", string(body))
}

// TODO: GET /friends/id/:id
func TestGetFriendByID(t *testing.T) {
	assert.Equal(t, true, true)
}

// TODO: GET /friends/name/:name
func TestGetFriendByName(t *testing.T) {
	assert.Equal(t, true, true)
}

// TODO: PUT /friends/:id
func TestPutFriendByID(t *testing.T) {
	assert.Equal(t, true, true)
}
