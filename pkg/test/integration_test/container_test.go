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

// GET /friends/id/:id
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

// GET /friends/name/:name
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

// PUT /friends/:id
// Tests updating all the fields of a friend in the DB
func TestContainerPutFriendByID(t *testing.T) {
	cli, ctx, err := SetupTests()
	assert.NoError(t, err)

	resp, _, err := startContainer(cli, ctx)
	assert.NoError(t, err)

	defer stopContainer(cli, ctx, resp.ID)

	_, _, err = addFriend(mockFriendsList[0])
	assert.NoError(t, err)

	newName := "Master Chief"
	newLastContacted := "15/01/2024"
	newBirthday := "23/02/1996"
	newNote := "This is the new note"

	updatedFriend := models.Friend{
		Name:          newName,
		LastContacted: newLastContacted,
		Birthday:      newBirthday,
		Notes:         newNote,
	}
	jsonValue, _ := json.Marshal(updatedFriend)

	statusCode, _, err := performContainerRequest("PUT", "/friends/1", jsonValue)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, statusCode) // Check that we get the expected status code from the update

	_, respBody, err := performContainerRequest("GET", "/friends/id/1", nil) // Pull the latest info for the ID
	assert.NoError(t, err)

	var respJson models.Friend
	err = json.Unmarshal(respBody, &respJson)
	assert.NoError(t, err)

	assert.Equal(t, "1", respJson.ID)
	assert.Equal(t, newName, respJson.Name)
	assert.Equal(t, newNote, respJson.Notes)
	assert.Equal(t, newLastContacted, respJson.LastContacted)
	assert.Equal(t, newBirthday, respJson.Birthday)
}

// PUT /friends/:id
// Tests updating the name field of a friend
func TestContainerPutFriendByIDNameOnly(t *testing.T) {
	cli, ctx, err := SetupTests()
	assert.NoError(t, err)

	resp, _, err := startContainer(cli, ctx)
	assert.NoError(t, err)

	defer stopContainer(cli, ctx, resp.ID)

	_, _, err = addFriend(mockFriendsList[0])
	assert.NoError(t, err)

	newName := "Master Chief"

	updatedFriend := models.Friend{
		Name: newName,
	}
	jsonValue, _ := json.Marshal(updatedFriend)

	statusCode, _, err := performContainerRequest("PUT", "/friends/1", jsonValue)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, statusCode) // Check that we get the expected status code from the update

	_, respBody, err := performContainerRequest("GET", "/friends/id/1", nil) // Pull the latest info for the ID
	assert.NoError(t, err)

	var respJson models.Friend
	err = json.Unmarshal(respBody, &respJson)
	assert.NoError(t, err)

	assert.Equal(t, "1", respJson.ID)
	assert.Equal(t, newName, respJson.Name)
	assert.Equal(t, "Nice guy", respJson.Notes)
	assert.Equal(t, "06/06/2023", respJson.LastContacted)
	assert.Equal(t, "23/02/1996", respJson.Birthday)
}

// PUT /friends/:id
// Tests updating the LastContacted field of a friend
func TestContainerPutFriendByIDLastContactedOnly(t *testing.T) {
	cli, ctx, err := SetupTests()
	assert.NoError(t, err)

	resp, _, err := startContainer(cli, ctx)
	assert.NoError(t, err)

	defer stopContainer(cli, ctx, resp.ID)

	_, _, err = addFriend(mockFriendsList[0])
	assert.NoError(t, err)

	newLastContacted := "05/06/2024"

	updatedFriend := models.Friend{
		LastContacted: newLastContacted,
	}
	jsonValue, _ := json.Marshal(updatedFriend)

	statusCode, _, err := performContainerRequest("PUT", "/friends/1", jsonValue)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, statusCode) // Check that we get the expected status code from the update

	_, respBody, err := performContainerRequest("GET", "/friends/id/1", nil) // Pull the latest info for the ID
	assert.NoError(t, err)

	var respJson models.Friend
	err = json.Unmarshal(respBody, &respJson)
	assert.NoError(t, err)

	assert.Equal(t, "1", respJson.ID)
	assert.Equal(t, "John Wick", respJson.Name)
	assert.Equal(t, "Nice guy", respJson.Notes)
	assert.Equal(t, newLastContacted, respJson.LastContacted)
	assert.Equal(t, "23/02/1996", respJson.Birthday)
}

// PUT /friends/:id
// Tests updating the Birthday field of a friend
func TestContainerPutFriendByIDBirthdayOnly(t *testing.T) {
	cli, ctx, err := SetupTests()
	assert.NoError(t, err)

	resp, _, err := startContainer(cli, ctx)
	assert.NoError(t, err)

	defer stopContainer(cli, ctx, resp.ID)

	_, _, err = addFriend(mockFriendsList[0])
	assert.NoError(t, err)

	newBirthday := "05/06/2024"

	updatedFriend := models.Friend{
		Birthday: newBirthday,
	}
	jsonValue, _ := json.Marshal(updatedFriend)

	statusCode, _, err := performContainerRequest("PUT", "/friends/1", jsonValue)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, statusCode) // Check that we get the expected status code from the update

	_, respBody, err := performContainerRequest("GET", "/friends/id/1", nil) // Pull the latest info for the ID
	assert.NoError(t, err)

	var respJson models.Friend
	err = json.Unmarshal(respBody, &respJson)
	assert.NoError(t, err)

	assert.Equal(t, "1", respJson.ID)
	assert.Equal(t, "John Wick", respJson.Name)
	assert.Equal(t, "Nice guy", respJson.Notes)
	assert.Equal(t, "06/06/2023", respJson.LastContacted)
	assert.Equal(t, newBirthday, respJson.Birthday)
}

// PUT /friends/:id
// Tests updating the Notes field of a friend
func TestContainerPutFriendByIDNotesOnly(t *testing.T) {
	cli, ctx, err := SetupTests()
	assert.NoError(t, err)

	resp, _, err := startContainer(cli, ctx)
	assert.NoError(t, err)

	defer stopContainer(cli, ctx, resp.ID)

	_, _, err = addFriend(mockFriendsList[0])
	assert.NoError(t, err)

	newNotes := "RIP his wife and dog"

	updatedFriend := models.Friend{
		Notes: newNotes,
	}
	jsonValue, _ := json.Marshal(updatedFriend)

	statusCode, _, err := performContainerRequest("PUT", "/friends/1", jsonValue)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, statusCode) // Check that we get the expected status code from the update

	_, respBody, err := performContainerRequest("GET", "/friends/id/1", nil) // Pull the latest info for the ID
	assert.NoError(t, err)

	var respJson models.Friend
	err = json.Unmarshal(respBody, &respJson)
	assert.NoError(t, err)

	assert.Equal(t, "1", respJson.ID)
	assert.Equal(t, "John Wick", respJson.Name)
	assert.Equal(t, newNotes, respJson.Notes)
	assert.Equal(t, "06/06/2023", respJson.LastContacted)
	assert.Equal(t, "23/02/1996", respJson.Birthday)
}
