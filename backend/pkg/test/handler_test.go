package test

import (
	"bytes"
	"encoding/json"
	"howarethey/pkg/handler"
	"howarethey/pkg/logger"
	"howarethey/pkg/models"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func performRequest(r http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)
	return recorder
}

func setupMockHandler() *handler.FriendsHandler {
	mockFriendsList := models.FriendsList{
		models.Friend{ID: "1", Name: "John Wick", LastContacted: "06/06/2023", Notes: "Nice guy"},
		models.Friend{ID: "2", Name: "Peter Parker", LastContacted: "12/12/2023", Notes: "I think he's Spiderman"},
	}

	mockDb, _ := SetupTestDB()

	mockFriendsHandler := &handler.FriendsHandler{
		FriendsList: mockFriendsList,
		DB:          mockDb,
	}

	return mockFriendsHandler
}

func setupTestEnvironment() (*gin.Engine, *handler.FriendsHandler, error) {
	mockFriendsHandler := setupMockHandler()
	mockRouter := handler.SetupRouter(mockFriendsHandler)

	return mockRouter, mockFriendsHandler, nil
}

// Test GET /friends/count
func TestFriendsCountRoute(t *testing.T) {
	mockRouter, _, err := setupTestEnvironment()
	assert.NoError(t, err)

	response := performRequest(mockRouter, "GET", "/friends/count", nil)

	assert.Equal(t, 200, response.Code)
	assert.Equal(t, "2", response.Body.String())
}

// Test GET /friends
func TestFriendsListRoute(t *testing.T) {
	router, _, err := setupTestEnvironment()
	assert.NoError(t, err)

	response := performRequest(router, "GET", "/friends", nil)

	expectedResult := `[{"ID":"1","Name":"John Wick","LastContacted":"06/06/2023","Notes":"Nice guy"},{"ID":"2","Name":"Peter Parker","LastContacted":"12/12/2023","Notes":"I think he's Spiderman"}]`

	assert.Equal(t, 200, response.Code)
	assert.Equal(t, expectedResult, response.Body.String())
}

// Test GET /friends/id/:id
func TestFriendIDRoute(t *testing.T) {
	router, _, err := setupTestEnvironment()
	assert.NoError(t, err)

	response := performRequest(router, "GET", "/friends/id/1", nil)

	expectedResult := `{"ID":"1","Name":"John Wick","LastContacted":"06/06/2023","Notes":"Nice guy"}`

	assert.Equal(t, 200, response.Code)
	assert.Equal(t, expectedResult, response.Body.String())
}

// Test GET /friends/name/:name
func TestFriendNameRoute(t *testing.T) {
	router, _, err := setupTestEnvironment()
	assert.NoError(t, err)

	response := performRequest(router, "GET", "/friends/name/john-wick", nil)

	expectedResult := `{"ID":"1","Name":"John Wick","LastContacted":"06/06/2023","Notes":"Nice guy"}`

	assert.Equal(t, 200, response.Code)
	assert.Equal(t, expectedResult, response.Body.String())
}

// Test GET /friends/id/:id
func TestMissingFriendIDRoute(t *testing.T) {
	router, _, err := setupTestEnvironment()
	assert.NoError(t, err)

	response := performRequest(router, "GET", "/friends/id/100", nil)

	expectedResult := `{"error":"friend not found"}`

	assert.Equal(t, 404, response.Code)
	assert.Equal(t, expectedResult, response.Body.String())
}

// Test GET /friends/random
func TestGetRandomFriend(t *testing.T) {

	mockRouter, mockFriendsHandler, err := setupTestEnvironment()

	response := performRequest(mockRouter, "GET", "/friends/random", nil)

	assert.Equal(t, http.StatusOK, response.Code)

	var friendResponse models.Friend
	err = json.Unmarshal(response.Body.Bytes(), &friendResponse)
	assert.NoError(t, err)

	found := false
	today := time.Now().Format("02/01/2006")
	for _, mockFriend := range mockFriendsHandler.FriendsList {
		logger.LogMessage(logger.LogLevelDebug, mockFriend.Name)
		if mockFriend.ID == friendResponse.ID && mockFriend.Name == friendResponse.Name && mockFriend.LastContacted == today {
			found = true
			break
		}
	}

	assert.True(t, found, "The returned friend should be in the mock friends list")
}

// Test POST /friends
func TestAddFriendRoute(t *testing.T) {
	newFriend := models.Friend{
		Name:          "Jane Doe",
		LastContacted: "15/01/2024",
		Notes:         "I don't think she's a real person",
	}
	jsonValue, _ := json.Marshal(newFriend)

	mockRouter, mockFriendsHandler, err := setupTestEnvironment()

	response := performRequest(mockRouter, "POST", "/friends", jsonValue)

	assert.Equal(t, http.StatusCreated, response.Code)

	var resp map[string]string
	err = json.Unmarshal(response.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Jane Doe added successfully", resp["message"])

	var count int
	err = mockFriendsHandler.DB.QueryRow("SELECT COUNT(*) FROM friends WHERE id = 1").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

// Test DELETE /friend/:id
func TestDeleteFriendRoute(t *testing.T) {

	mockRouter, mockFriendsHandler, err := setupTestEnvironment()

	err = insertMockFriend(mockFriendsHandler.DB, "1", "John Wick", "06/06/2023", "Nice guy")
	assert.NoError(t, err)

	response := performRequest(mockRouter, "DELETE", "/friends/1", nil)

	assert.Equal(t, http.StatusCreated, response.Code)

	var resp map[string]string
	err = json.Unmarshal(response.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Friend removed successfully", resp["message"])

	var count int
	err = mockFriendsHandler.DB.QueryRow("SELECT COUNT(*) FROM friends WHERE id = ?", "1").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

// Tests PUT /friends/:id
func TestPutFriend(t *testing.T) {
	mockFriend := models.Friend{
		ID:            "1",
		Name:          "John Wick",
		LastContacted: "06/06/2023",
		Notes:         "Nice guy",
	}

	mockRouter, mockFriendsHandler, err := setupTestEnvironment()

	err = insertMockFriend(mockFriendsHandler.DB,
		mockFriend.ID,
		mockFriend.Name,
		mockFriend.LastContacted,
		mockFriend.Notes,
	)
	assert.NoError(t, err)

	updatedFriend := models.Friend{
		Name:          "Master Chief",
		LastContacted: "15/01/2024",
		Notes:         "Doesn't talk much",
	}
	jsonValue, _ := json.Marshal(updatedFriend)

	response := performRequest(mockRouter, "PUT", "/friends/1", jsonValue)

	assert.Equal(t, http.StatusOK, response.Code)

	var resp map[string]string
	err = json.Unmarshal(response.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "1 updated successfully", resp["message"])

	var friend models.Friend
	err = mockFriendsHandler.DB.QueryRow("SELECT name, lastContacted, notes FROM friends WHERE id = ?", "1").Scan(&friend.Name, &friend.LastContacted, &friend.Notes)
	assert.NoError(t, err)
	assert.Equal(t, updatedFriend.Name, friend.Name)
	assert.Equal(t, updatedFriend.LastContacted, friend.LastContacted)
	assert.Equal(t, updatedFriend.Notes, friend.Notes)
}

// Tests PUT /friends/:id
// Tests the endpoint when only Notes field is provided
func TestPutNotesOnly(t *testing.T) {
	mockFriend := models.Friend{
		ID:            "1",
		Name:          "John Wick",
		LastContacted: "06/06/2023",
		Notes:         "Nice guy",
	}

	mockRouter, mockFriendsHandler, err := setupTestEnvironment()

	err = insertMockFriend(mockFriendsHandler.DB,
		mockFriend.ID,
		mockFriend.Name,
		mockFriend.LastContacted,
		mockFriend.Notes,
	)
	assert.NoError(t, err)

	updatedFriend := models.Friend{
		Notes: "Bro is Chuck Norris",
	}
	jsonValue, _ := json.Marshal(updatedFriend)

	response := performRequest(mockRouter, "PUT", "/friends/1", jsonValue)

	assert.Equal(t, http.StatusOK, response.Code)

	var resp map[string]string
	err = json.Unmarshal(response.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "1 updated successfully", resp["message"])

	var friend models.Friend
	err = mockFriendsHandler.DB.QueryRow("SELECT name, lastContacted, notes FROM friends WHERE id = ?", "1").Scan(&friend.Name, &friend.LastContacted, &friend.Notes)
	assert.NoError(t, err)
	assert.Equal(t, mockFriend.Name, friend.Name)
	assert.Equal(t, mockFriend.LastContacted, friend.LastContacted)
	assert.Equal(t, updatedFriend.Notes, friend.Notes)
}

// Tests PUT /friends/:id
// Tests the endpoint when only Name field is provided
func TestPutNameOnly(t *testing.T) {
	mockFriend := models.Friend{
		ID:            "1",
		Name:          "John Wick",
		LastContacted: "06/06/2023",
		Notes:         "Nice guy",
	}

	mockRouter, mockFriendsHandler, err := setupTestEnvironment()

	err = insertMockFriend(mockFriendsHandler.DB,
		mockFriend.ID,
		mockFriend.Name,
		mockFriend.LastContacted,
		mockFriend.Notes,
	)
	assert.NoError(t, err)

	updatedFriend := models.Friend{
		Name: "Winnie the Pooh",
	}
	jsonValue, _ := json.Marshal(updatedFriend)

	response := performRequest(mockRouter, "PUT", "/friends/1", jsonValue)

	assert.Equal(t, http.StatusOK, response.Code)

	var resp map[string]string
	err = json.Unmarshal(response.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "1 updated successfully", resp["message"])

	var friend models.Friend
	err = mockFriendsHandler.DB.QueryRow("SELECT name, lastContacted, notes FROM friends WHERE id = ?", "1").Scan(&friend.Name, &friend.LastContacted, &friend.Notes)
	assert.NoError(t, err)
	assert.Equal(t, updatedFriend.Name, friend.Name)
	assert.Equal(t, mockFriend.LastContacted, friend.LastContacted)
	assert.Equal(t, mockFriend.Notes, friend.Notes)
}

// Tests PUT /friends/:id
// Tests the endpoint when only Last Contacted field is provided
func TestPutLastContactedOnly(t *testing.T) {
	todaysDate := time.Now().Format("02/01/2006")
	mockFriend := models.Friend{
		ID:            "1",
		Name:          "John Wick",
		LastContacted: "06/06/2023",
		Notes:         "Nice guy",
	}

	mockRouter, mockFriendsHandler, err := setupTestEnvironment()

	err = insertMockFriend(mockFriendsHandler.DB,
		mockFriend.ID,
		mockFriend.Name,
		mockFriend.LastContacted,
		mockFriend.Notes,
	)
	assert.NoError(t, err)

	updatedFriend := models.Friend{
		LastContacted: todaysDate,
	}
	jsonValue, _ := json.Marshal(updatedFriend)

	response := performRequest(mockRouter, "PUT", "/friends/1", jsonValue)

	assert.Equal(t, http.StatusOK, response.Code)

	var resp map[string]string
	err = json.Unmarshal(response.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "1 updated successfully", resp["message"])

	var friend models.Friend
	err = mockFriendsHandler.DB.QueryRow("SELECT name, lastContacted, notes FROM friends WHERE id = ?", "1").Scan(&friend.Name, &friend.LastContacted, &friend.Notes)
	assert.NoError(t, err)
	assert.Equal(t, mockFriend.Name, friend.Name)
	assert.Equal(t, todaysDate, friend.LastContacted)
	assert.Equal(t, mockFriend.Notes, friend.Notes)
}
