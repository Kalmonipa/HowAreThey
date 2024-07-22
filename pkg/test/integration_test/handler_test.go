package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"howarethey/pkg/handler"
	"howarethey/pkg/logger"
	"howarethey/pkg/models"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func insertMockFriend(db *sql.DB, id string, name string, lastContacted string, birthday string, notes string) error {
	stmt, err := db.Prepare("INSERT INTO friends (id, name, lastContacted, birthday, notes) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id, name, lastContacted, birthday, notes)
	return err
}

func setupTestDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS friends (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		lastContacted TEXT NOT NULL,
		birthday TEXT NOT NULL,
		notes TEXT NOT NULL
	);`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func performHandlerRequest(r http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)
	return recorder
}

func setupMockHandler(mockFriendsList models.FriendsList) *handler.FriendsHandler {
	// TODO: The test DB doesn't actually contain the above friends
	// At the moment, that's not an issue but probably worth adding them to the DB as well
	mockDb, _ := setupTestDB()

	mockFriendsHandler := &handler.FriendsHandler{
		FriendsList: mockFriendsList,
		DB:          mockDb,
	}

	return mockFriendsHandler
}

func setupTestEnvironment(isFriendsListPopulated bool) (*gin.Engine, *handler.FriendsHandler, error) {
	var mockFriendsList models.FriendsList
	if isFriendsListPopulated {
		mockFriendsList = models.FriendsList{
			models.Friend{ID: "1", Name: "John Wick", LastContacted: "2023-06-06", Birthday: "1996-02-23", Notes: "Nice guy"},
			models.Friend{ID: "2", Name: "Peter Parker", LastContacted: "2023-12-12", Birthday: "1996-02-23", Notes: "I think he's Spiderman"},
		}
	} else {
		mockFriendsList = models.FriendsList{}
	}

	mockFriendsHandler := setupMockHandler(mockFriendsList)
	mockRouter := handler.SetupRouter(mockFriendsHandler)

	return mockRouter, mockFriendsHandler, nil
}

// Test GET /birthdays
func TestFriendsBirthday(t *testing.T) {
	os.Setenv("TEST_ENV", "true")
	logger.SetupLogger()

	mockRouter, _, err := setupTestEnvironment(true)
	assert.NoError(t, err)

	response := performHandlerRequest(mockRouter, "GET", "/birthdays", nil)

	todaysDate := time.Now().Format("01-02")

	assert.Equal(t, http.StatusOK, response.Code)

	if todaysDate == "02-23" {
		expectedResult, err := json.Marshal(mockFriendsList)
		if err != nil {
			fmt.Println(err)
			return
		}
		assert.Equal(t, string(expectedResult), response.Body.String())
	} else {
		assert.Equal(t, "[]", response.Body.String())
	}

}

// Test GET /friends/count
func TestFriendsCountRoute(t *testing.T) {
	mockRouter, _, err := setupTestEnvironment(true)
	assert.NoError(t, err)

	response := performHandlerRequest(mockRouter, "GET", "/friends/count", nil)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "2", response.Body.String())
}

// Test GET /friends
func TestFriendsListRoute(t *testing.T) {
	router, _, err := setupTestEnvironment(true)
	assert.NoError(t, err)

	response := performHandlerRequest(router, "GET", "/friends", nil)

	expectedResult, err := json.Marshal(mockFriendsList)
	if err != nil {
		fmt.Println(err)
		return
	}

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, string(expectedResult), response.Body.String())
}

// Test GET /friends
// If the friends list is empty, it should return []
func TestEmptyFriendsList(t *testing.T) {
	router, _, err := setupTestEnvironment(false)
	assert.NoError(t, err)

	response := performHandlerRequest(router, "GET", "/friends", nil)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "[]", response.Body.String())
}

// Test GET /friends/id/:id
func TestFriendIDRoute(t *testing.T) {
	router, _, err := setupTestEnvironment(true)
	assert.NoError(t, err)

	response := performHandlerRequest(router, "GET", "/friends/id/1", nil)

	expectedResult, err := json.Marshal(mockFriendsList[0])
	if err != nil {
		fmt.Println(err)
		return
	}

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, string(expectedResult), response.Body.String())
}

// Test GET /friends/name/:name
func TestFriendNameRoute(t *testing.T) {
	router, _, err := setupTestEnvironment(true)
	assert.NoError(t, err)

	response := performHandlerRequest(router, "GET", "/friends/name/john-wick", nil)

	expectedResult, err := json.Marshal(mockFriendsList[0])
	if err != nil {
		fmt.Println(err)
		return
	}

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, string(expectedResult), response.Body.String())
}

// Test GET /friends/id/:id
// Searches for an id that shouldn't exist
func TestMissingFriendIDRoute(t *testing.T) {
	router, _, err := setupTestEnvironment(true)
	assert.NoError(t, err)

	response := performHandlerRequest(router, "GET", "/friends/id/100", nil)

	expectedResult := `{"error":"friend not found"}`

	assert.Equal(t, 404, response.Code)
	assert.Equal(t, expectedResult, response.Body.String())
}

// Test GET /friends/random
func TestGetRandomFriend(t *testing.T) {
	os.Setenv("TEST_ENV", "true")
	logger.SetupLogger()

	mockRouter, mockFriendsHandler, err := setupTestEnvironment(true)
	assert.NoError(t, err)

	response := performHandlerRequest(mockRouter, "GET", "/friends/random", nil)

	assert.Equal(t, http.StatusOK, response.Code)

	var friendResponse models.Friend
	err = json.Unmarshal(response.Body.Bytes(), &friendResponse)
	assert.NoError(t, err)

	found := false
	today := time.Now().Format("2006-01-02")
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
	os.Setenv("TEST_ENV", "true")
	logger.SetupLogger()

	newFriend := models.Friend{
		Name:          "Jane Doe",
		LastContacted: "2024-01-15",
		Birthday:      "1996-02-23",
		Notes:         "I don't think she's a real person",
	}
	jsonValue, _ := json.Marshal(newFriend)

	mockRouter, _, err := setupTestEnvironment(true)
	assert.NoError(t, err)

	response := performHandlerRequest(mockRouter, "POST", "/friends", jsonValue)

	assert.Equal(t, http.StatusCreated, response.Code)

	var resp map[string]string
	err = json.Unmarshal(response.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Jane Doe added successfully", resp["message"])

}

// Test POST /friends
// Using no data so it should fail
func TestAddFriendRouteNoData(t *testing.T) {
	os.Setenv("TEST_ENV", "true")
	logger.SetupLogger()

	newFriend := models.Friend{}
	jsonValue, _ := json.Marshal(newFriend)

	mockRouter, _, err := setupTestEnvironment(false)
	assert.NoError(t, err)

	response := performHandlerRequest(mockRouter, "POST", "/friends", jsonValue)

	assert.Equal(t, http.StatusInternalServerError, response.Code)

	var resp map[string]string
	err = json.Unmarshal(response.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "name must not be blank", resp["error"])
}

// Test POST /friends
// Using bad LastContacted data so it should fail
func TestAddFriendRouteBadLastContactedData(t *testing.T) {
	os.Setenv("TEST_ENV", "true")
	logger.SetupLogger()

	newFriend := models.Friend{
		Name:          "Jane Doe",
		LastContacted: "15",
		Birthday:      "1990-01-23",
		Notes:         "I don't think she's a real person",
	}
	jsonValue, _ := json.Marshal(newFriend)

	mockRouter, _, err := setupTestEnvironment(true)
	assert.NoError(t, err)

	response := performHandlerRequest(mockRouter, "POST", "/friends", jsonValue)

	assert.Equal(t, http.StatusInternalServerError, response.Code)

	var resp map[string]string
	err = json.Unmarshal(response.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "last Contacted date must be in yyyy-mm-dd format. 15 does not match", resp["error"])

}

// Test POST /friends
// Using bad Birthday data so it should fail
func TestAddFriendRouteBadBirthdayData(t *testing.T) {
	os.Setenv("TEST_ENV", "true")
	logger.SetupLogger()

	newFriend := models.Friend{
		Name:          "Jane Doe",
		LastContacted: "1990-01-15",
		Birthday:      "23",
		Notes:         "I don't think she's a real person",
	}
	jsonValue, _ := json.Marshal(newFriend)

	mockRouter, _, err := setupTestEnvironment(true)
	assert.NoError(t, err)

	response := performHandlerRequest(mockRouter, "POST", "/friends", jsonValue)

	assert.Equal(t, http.StatusInternalServerError, response.Code)

	var resp map[string]string
	err = json.Unmarshal(response.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "birthday must be in yyyy-mm-dd format. 23 does not match", resp["error"])

}

// Test DELETE /friend/:id
func TestDeleteFriendRoute(t *testing.T) {
	os.Setenv("TEST_ENV", "true")
	logger.SetupLogger()

	mockRouter, mockFriendsHandler, err := setupTestEnvironment(true)
	assert.NoError(t, err)

	err = insertMockFriend(mockFriendsHandler.DB, "1", "John Wick", "2023-06-06", "1996-02-23", "Nice guy")
	assert.NoError(t, err)

	response := performHandlerRequest(mockRouter, "DELETE", "/friends/1", nil)

	assert.Equal(t, http.StatusOK, response.Code)

	var resp map[string]string
	err = json.Unmarshal(response.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "John Wick removed successfully", resp["message"])

	var count int
	err = mockFriendsHandler.DB.QueryRow("SELECT COUNT(*) FROM friends WHERE id = ?", "1").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

// Tests PUT /friends/:id
func TestPutFriend(t *testing.T) {
	os.Setenv("TEST_ENV", "true")
	logger.SetupLogger()

	mockFriend := mockFriendsList[0]

	mockRouter, mockFriendsHandler, err := setupTestEnvironment(true)
	assert.NoError(t, err)

	err = insertMockFriend(mockFriendsHandler.DB,
		mockFriend.ID,
		mockFriend.Name,
		mockFriend.LastContacted,
		mockFriend.Birthday,
		mockFriend.Notes,
	)
	assert.NoError(t, err)

	updatedFriend := models.Friend{
		Name:          "Master Chief",
		LastContacted: "2024-01-15",
		Birthday:      "1996-02-23",
		Notes:         "Doesn't talk much",
	}
	jsonValue, _ := json.Marshal(updatedFriend)

	response := performHandlerRequest(mockRouter, "PUT", "/friends/1", jsonValue)

	assert.Equal(t, http.StatusOK, response.Code)

	var resp map[string]string
	err = json.Unmarshal(response.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "1 updated successfully", resp["message"])

	var friend models.Friend
	err = mockFriendsHandler.DB.QueryRow("SELECT name, lastContacted, birthday, notes FROM friends WHERE id = ?", "1").Scan(&friend.Name, &friend.LastContacted, &friend.Birthday, &friend.Notes)
	assert.NoError(t, err)
	assert.Equal(t, updatedFriend.Name, friend.Name)
	assert.Equal(t, updatedFriend.LastContacted, friend.LastContacted)
	assert.Equal(t, updatedFriend.Birthday, friend.Birthday)
	assert.Equal(t, updatedFriend.Notes, friend.Notes)
}

// Tests PUT /friends/:id
// Tests the endpoint when only Notes field is provided
func TestPutNotesOnly(t *testing.T) {
	os.Setenv("TEST_ENV", "true")
	logger.SetupLogger()

	mockFriend := mockFriendsList[0]

	mockRouter, mockFriendsHandler, err := setupTestEnvironment(true)
	assert.NoError(t, err)

	err = insertMockFriend(mockFriendsHandler.DB,
		mockFriend.ID,
		mockFriend.Name,
		mockFriend.LastContacted,
		mockFriend.Birthday,
		mockFriend.Notes,
	)
	assert.NoError(t, err)

	updatedFriend := models.Friend{
		Notes: "Bro is Chuck Norris",
	}
	jsonValue, _ := json.Marshal(updatedFriend)

	response := performHandlerRequest(mockRouter, "PUT", "/friends/1", jsonValue)

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
	os.Setenv("TEST_ENV", "true")
	logger.SetupLogger()

	mockFriend := mockFriendsList[0]

	mockRouter, mockFriendsHandler, err := setupTestEnvironment(true)
	assert.NoError(t, err)

	err = insertMockFriend(mockFriendsHandler.DB,
		mockFriend.ID,
		mockFriend.Name,
		mockFriend.LastContacted,
		mockFriend.Birthday,
		mockFriend.Notes,
	)
	assert.NoError(t, err)

	updatedFriend := models.Friend{
		Name: "Winnie the Pooh",
	}
	jsonValue, _ := json.Marshal(updatedFriend)

	response := performHandlerRequest(mockRouter, "PUT", "/friends/1", jsonValue)

	assert.Equal(t, http.StatusOK, response.Code)

	var resp map[string]string
	err = json.Unmarshal(response.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "1 updated successfully", resp["message"])

	var friend models.Friend
	err = mockFriendsHandler.DB.QueryRow("SELECT name, lastContacted, birthday, notes FROM friends WHERE id = ?", "1").Scan(&friend.Name, &friend.LastContacted, &friend.Birthday, &friend.Notes)
	assert.NoError(t, err)
	assert.Equal(t, updatedFriend.Name, friend.Name)
	assert.Equal(t, mockFriend.LastContacted, friend.LastContacted)
	assert.Equal(t, mockFriend.Birthday, friend.Birthday)
	assert.Equal(t, mockFriend.Notes, friend.Notes)
}

// Tests PUT /friends/:id
// Tests the endpoint when only Last Contacted field is provided
func TestPutLastContactedOnly(t *testing.T) {
	os.Setenv("TEST_ENV", "true")
	logger.SetupLogger()

	todaysDate := time.Now().Format("2006-01-02")
	mockFriend := mockFriendsList[0]

	mockRouter, mockFriendsHandler, err := setupTestEnvironment(true)
	assert.NoError(t, err)

	err = insertMockFriend(mockFriendsHandler.DB,
		mockFriend.ID,
		mockFriend.Name,
		mockFriend.LastContacted,
		mockFriend.Birthday,
		mockFriend.Notes,
	)
	assert.NoError(t, err)

	updatedFriend := models.Friend{
		LastContacted: todaysDate,
	}
	jsonValue, _ := json.Marshal(updatedFriend)

	response := performHandlerRequest(mockRouter, "PUT", "/friends/1", jsonValue)

	assert.Equal(t, http.StatusOK, response.Code)

	var resp map[string]string
	err = json.Unmarshal(response.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "1 updated successfully", resp["message"])

	var friend models.Friend
	err = mockFriendsHandler.DB.QueryRow("SELECT name, lastContacted, birthday, notes FROM friends WHERE id = ?", "1").Scan(&friend.Name, &friend.LastContacted, &friend.Birthday, &friend.Notes)
	assert.NoError(t, err)
	assert.Equal(t, mockFriend.Name, friend.Name)
	assert.Equal(t, todaysDate, friend.LastContacted)
	assert.Equal(t, mockFriend.Birthday, friend.Birthday)
	assert.Equal(t, mockFriend.Notes, friend.Notes)
}

// Tests PUT /friends/:id
// Tests the endpoint when only Birthday field is provided
func TestPutBirthdayOnly(t *testing.T) {
	os.Setenv("TEST_ENV", "true")
	logger.SetupLogger()

	todaysDate := time.Now().Format("2006-01-02")
	mockFriend := mockFriendsList[0]

	mockRouter, mockFriendsHandler, err := setupTestEnvironment(true)
	assert.NoError(t, err)

	err = insertMockFriend(mockFriendsHandler.DB,
		mockFriend.ID,
		mockFriend.Name,
		mockFriend.LastContacted,
		mockFriend.Birthday,
		mockFriend.Notes,
	)
	assert.NoError(t, err)

	updatedFriend := models.Friend{
		Birthday: todaysDate,
	}
	jsonValue, _ := json.Marshal(updatedFriend)

	response := performHandlerRequest(mockRouter, "PUT", "/friends/1", jsonValue)

	assert.Equal(t, http.StatusOK, response.Code)

	var resp map[string]string
	err = json.Unmarshal(response.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "1 updated successfully", resp["message"])

	var friend models.Friend
	err = mockFriendsHandler.DB.QueryRow("SELECT name, lastContacted, birthday, notes FROM friends WHERE id = ?", "1").Scan(&friend.Name, &friend.LastContacted, &friend.Birthday, &friend.Notes)
	assert.NoError(t, err)
	assert.Equal(t, mockFriend.Name, friend.Name)
	assert.Equal(t, mockFriend.LastContacted, friend.LastContacted)
	assert.Equal(t, todaysDate, friend.Birthday)
	assert.Equal(t, mockFriend.Notes, friend.Notes)
}
