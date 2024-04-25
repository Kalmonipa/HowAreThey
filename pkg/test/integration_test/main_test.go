package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"howarethey/pkg/logger"
	"howarethey/pkg/models"
	"io"
	"net/http"
	"os"
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

func getGitBranchName() (string, error) {
	os.Setenv("TEST_ENV", "true")
	logger.SetupLogger()

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

// func performRequest(method, path string, body []byte) (*http.Response, error) {
// 	client := &http.Client{}

// 	fullPath := "http://localhost:" + portNumber + path

// 	req, err := http.NewRequest(method, fullPath, bytes.NewBuffer(body))
// 	if err != nil {
// 		return nil, err
// 	}
// 	req.Header.Set("Content-Type", "application/json")

// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	return resp, nil
// }

// This test is redundant because we wouldn't have got this far if the daemon wasn't running
func TestDockerDaemonRunning(t *testing.T) {
	os.Setenv("TEST_ENV", "true")
	logger.SetupLogger()

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
	os.Setenv("TEST_ENV", "true")
	logger.SetupLogger()

	cli, ctx, err := SetupTests()
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	branchName, err := getGitBranchName()
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	imageName = "kalmonipa/howarethey:" + branchName

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        imageName,
		ExposedPorts: nat.PortSet{"8080": struct{}{}},
	}, &container.HostConfig{
		PortBindings: map[nat.Port][]nat.PortBinding{nat.Port(portNumber): {{HostIP: "127.0.0.1", HostPort: portNumber}}},
	}, nil, nil, branchName)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		t.Errorf("Error: %v", err)
	}

	// Sleeping to give the webserver time to start up
	time.Sleep(2 * time.Second)

	containerState, err := cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	assert.Equal(t, "/"+branchName, containerState.ContainerJSONBase.Name) // Apparently the container name has a leading /
	assert.Equal(t, "running", containerState.ContainerJSONBase.State.Status)

}

func TestAddFriend(t *testing.T) {
	os.Setenv("TEST_ENV", "true")
	logger.SetupLogger()

	mockFriend := mockFriendsList[0]

	data := map[string]string{"Name": mockFriend.Name, "LastContacted": mockFriend.LastContacted}
	payload, err := json.Marshal(data)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	resp, err := http.Post("http://localhost:"+portNumber+"/friends",
		"application/json",
		bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	expectedResponse := "{\"message\":\"John Wick added successfully\"}"

	assert.Equal(t, expectedResponse, string(bodyBytes))
}

func TestDeleteFriend(t *testing.T) {
	// response, err := performRequest("DELETE", "/friends/1", nil)
	// if err != nil {
	// 	t.Errorf("Error: %v", err)
	// }

	client := &http.Client{}

	fullPath := "http://localhost:" + portNumber + "/friends/1"

	req, err := http.NewRequest(http.MethodDelete, fullPath, nil)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	var respJson map[string]string
	err = json.Unmarshal(respBody, &respJson)
	assert.NoError(t, err)
	assert.Equal(t, "John Wick removed successfully", respJson["message"])
}
