package integration

import (
	"context"
	"encoding/json"
	"howarethey/pkg/models"
	"net/http"
	"testing"
	"time"

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

func TestContainerBirthdays(t *testing.T) {
	cli, ctx, err := SetupTests()
	assert.NoError(t, err)

	resp, _, err := startContainer(cli, ctx)
	assert.NoError(t, err)

	defer stopContainer(cli, ctx, resp.ID)

	mockFriend := models.Friend{Name: "Jimmy Neutron", LastContacted: "06/06/2023", Birthday: time.Now().Format("02/01/2006")}

	_, _, err = addFriend(mockFriend)
	assert.NoError(t, err)

	respStatus, respBody, err := performContainerRequest("GET", "/birthdays", nil)
	assert.NoError(t, err)

	var respJson []map[string]string
	err = json.Unmarshal(respBody, &respJson)
	assert.NoError(t, err)

	friend := respJson[0]
	assert.Equal(t, http.StatusOK, respStatus)
	assert.Equal(t, mockFriend.Name, friend["Name"])
}

// TODO: GET /friends/random
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

// TODO: GET /friends/id/:id
func TestContainerGetFriendByID(t *testing.T) {
	cli, ctx, err := SetupTests()
	assert.NoError(t, err)

	resp, _, err := startContainer(cli, ctx)
	assert.NoError(t, err)

	defer stopContainer(cli, ctx, resp.ID)

	_, _, err = addFriend(mockFriendsList[0])
	assert.NoError(t, err)

	statusCode, respBody, err := performContainerRequest("GET", "/friends/id/1", nil)
	assert.NoError(t, err)

	var respJson models.Friend
	err = json.Unmarshal(respBody, &respJson)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "1", respJson.ID)
	assert.Equal(t, "John Wick", respJson.Name)
}

// TODO: GET /friends/name/:name
func TestContainerGetFriendByName(t *testing.T) {
	cli, ctx, err := SetupTests()
	assert.NoError(t, err)

	resp, _, err := startContainer(cli, ctx)
	assert.NoError(t, err)

	defer stopContainer(cli, ctx, resp.ID)

	_, _, err = addFriend(mockFriendsList[1])
	assert.NoError(t, err)

	statusCode, respBody, err := performContainerRequest("GET", "/friends/name/peter-parker", nil)
	assert.NoError(t, err)

	var respJson models.Friend
	err = json.Unmarshal(respBody, &respJson)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "1", respJson.ID)
	assert.Equal(t, "Peter Parker", respJson.Name)
}

// TODO: PUT /friends/:id
func TestContainerPutFriendByID(t *testing.T) {
	cli, ctx, err := SetupTests()
	assert.NoError(t, err)

	resp, _, err := startContainer(cli, ctx)
	assert.NoError(t, err)

	defer stopContainer(cli, ctx, resp.ID)

	_, _, err = addFriend(mockFriendsList[0])
	assert.NoError(t, err)

	newNote := "This is the new note"

	updatedFriend := models.Friend{
		Name:          "Master Chief",
		LastContacted: "15/01/2024",
		Birthday:      "23/02/1996",
		Notes:         "Doesn't talk much",
	}
	jsonValue, _ := json.Marshal(updatedFriend)

	statusCode, respBody, err := performContainerRequest("PUT", "/friends/1", jsonValue)
	assert.NoError(t, err)

	var respJson models.Friend
	err = json.Unmarshal(respBody, &respJson)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "1", respJson.ID)
	assert.Equal(t, "Peter Parker", respJson.Name)
	assert.Equal(t, newNote, respJson.Notes)
}