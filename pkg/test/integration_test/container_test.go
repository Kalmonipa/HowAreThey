package integration

import (
	"context"
	"encoding/json"
	"howarethey/pkg/models"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// This test is redundant because we wouldn't have got this far if the daemon wasn't running
func TestDockerDaemonRunning(t *testing.T) {
	cli, _, err := SetupTests()
	assert.NoError(t, err)

	// Ping the Docker daemon
	_, err = cli.Ping(context.Background())
	assert.NoError(t, err)
}

func TestDockerContainerRunning(t *testing.T) {
	cli, ctx, err := SetupTests()
	assert.NoError(t, err)

	resp, branchName, err := startContainer(cli, ctx)
	assert.NoError(t, err)

	defer stopContainer(cli, ctx, resp.ID)

	containerState, err := cli.ContainerInspect(ctx, resp.ID)
	assert.NoError(t, err)

	assert.Equal(t, "/"+branchName, containerState.ContainerJSONBase.Name) // Apparently the container name has a leading /
	assert.Equal(t, "running", containerState.ContainerJSONBase.State.Status)

}

func TestContainerAddFriend(t *testing.T) {
	cli, ctx, err := SetupTests()
	assert.NoError(t, err)

	resp, _, err := startContainer(cli, ctx)
	assert.NoError(t, err)

	defer stopContainer(cli, ctx, resp.ID)

	respStatus, respBody, err := addFriend(mockFriendsList[0])
	assert.NoError(t, err)

	expectedResponse := "{\"message\":\"John Wick added successfully\"}"

	assert.Equal(t, expectedResponse, string(respBody))

	assert.Equal(t, http.StatusCreated, respStatus)
}

func TestContainerDeleteFriend(t *testing.T) {
	cli, ctx, err := SetupTests()
	assert.NoError(t, err)

	resp, _, err := startContainer(cli, ctx)
	assert.NoError(t, err)

	defer stopContainer(cli, ctx, resp.ID)

	_, _, err = addFriend(mockFriendsList[0])
	assert.NoError(t, err)

	respStatus, respBody, err := performContainerRequest("DELETE", "/friends/1", nil)
	assert.NoError(t, err)

	var respJson map[string]string
	err = json.Unmarshal(respBody, &respJson)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, respStatus)
	assert.Equal(t, "John Wick removed successfully", respJson["message"])
}

// GET /friends/random
func TestContainerGetRandomFriend(t *testing.T) {
	cli, ctx, err := SetupTests()
	assert.NoError(t, err)

	resp, _, err := startContainer(cli, ctx)
	assert.NoError(t, err)

	defer stopContainer(cli, ctx, resp.ID)

	for _, friend := range mockFriendsList {
		_, _, err = addFriend(friend)
		assert.NoError(t, err)
	}

	statusCode, respBody, err := performContainerRequest("GET", "/friends/random", nil)
	assert.NoError(t, err)

	var respJson models.Friend
	err = json.Unmarshal(respBody, &respJson)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, statusCode)

	found := false
	for _, mockFriend := range mockFriendsList {
		if mockFriend.ID == respJson.ID && mockFriend.Name == respJson.Name && mockFriend.LastContacted == respJson.LastContacted {
			found = true
			break
		}
	}

	assert.True(t, found, "The returned friend should be in the mock friends list")

}

func TestContainerGetFriendCount(t *testing.T) {
	cli, ctx, err := SetupTests()
	assert.NoError(t, err)

	resp, _, err := startContainer(cli, ctx)
	assert.NoError(t, err)

	defer stopContainer(cli, ctx, resp.ID)

	_, _, err = addFriend(mockFriendsList[0])
	assert.NoError(t, err)

	statusCode, body, err := performContainerRequest("GET", "/friends/count", nil)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "1", string(body))
}
