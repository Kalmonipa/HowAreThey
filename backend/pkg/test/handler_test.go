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

func setupMockHandler() *handler.FriendsHandler {
	mockFriendsList := models.FriendsList{
		models.Friend{ID: "1", Name: "John Wick", LastContacted: "06/06/2023", Notes: "Nice guy"},
		models.Friend{ID: "2", Name: "Peter Parker", LastContacted: "12/12/2023", Notes: "I think he's Spiderman"},
	}
	mockFriendsHandler := &handler.FriendsHandler{
		FriendsList: mockFriendsList,
	}

	return mockFriendsHandler
}

// Test GET /friends/count
func TestFriendsCountRoute(t *testing.T) {

	mockFriendsHandler := setupMockHandler()

	router := handler.SetupRouter(mockFriendsHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/friends/count", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "2", w.Body.String())
}

// Test GET /friends
func TestFriendsListRoute(t *testing.T) {

	mockFriendsHandler := setupMockHandler()

	db, err := SetupTestDB()
	assert.NoError(t, err)
	defer db.Close()

	router := handler.SetupRouter(mockFriendsHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/friends", nil)
	router.ServeHTTP(w, req)

	expectedResult := `[{"ID":"1","Name":"John Wick","LastContacted":"06/06/2023","Notes":"Nice guy"},{"ID":"2","Name":"Peter Parker","LastContacted":"12/12/2023","Notes":"I think he's Spiderman"}]`

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, expectedResult, w.Body.String())
}

// Test GET /friends/id/:id
func TestFriendIDRoute(t *testing.T) {

	mockFriendsHandler := setupMockHandler()

	router := handler.SetupRouter(mockFriendsHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/friends/id/1", nil)
	router.ServeHTTP(w, req)

	expectedResult := `{"ID":"1","Name":"John Wick","LastContacted":"06/06/2023","Notes":"Nice guy"}`

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, expectedResult, w.Body.String())
}

// Test GET /friends/name/:name
func TestFriendNameRoute(t *testing.T) {

	mockFriendsHandler := setupMockHandler()

	router := handler.SetupRouter(mockFriendsHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/friends/name/john-wick", nil)
	router.ServeHTTP(w, req)

	expectedResult := `{"ID":"1","Name":"John Wick","LastContacted":"06/06/2023","Notes":"Nice guy"}`

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, expectedResult, w.Body.String())
}

// Test GET /friends/id/:id
func TestMissingFriendIDRoute(t *testing.T) {

	mockFriendsHandler := setupMockHandler()

	router := handler.SetupRouter(mockFriendsHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/friends/id/100", nil)
	router.ServeHTTP(w, req)

	expectedResult := `{"error":"friend not found"}`

	assert.Equal(t, 404, w.Code)
	assert.Equal(t, expectedResult, w.Body.String())
}

// GET /friends/random
func TestGetRandomFriend(t *testing.T) {
	mockFriendsList := models.FriendsList{
		models.Friend{ID: "1", Name: "John Wick", LastContacted: "06/06/2023", Notes: "Nice guy"},
		models.Friend{ID: "2", Name: "Peter Parker", LastContacted: "12/12/2023", Notes: "I think he's Spiderman"},
	}

	db, err := SetupTestDB()
	assert.NoError(t, err)
	defer db.Close()

	mockFriendsHandler := handler.NewFriendsHandler(mockFriendsList, db)

	mockRouter := handler.SetupRouter(mockFriendsHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/friends/random", nil)
	mockRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var friendResponse models.Friend
	err = json.Unmarshal(w.Body.Bytes(), &friendResponse)
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
	db, err := SetupTestDB()
	assert.NoError(t, err)
	defer db.Close()

	router := gin.New()
	mockFriendsHandler := &handler.FriendsHandler{DB: db}
	router.POST("/friends", mockFriendsHandler.PostNewFriend)

	newFriend := models.Friend{
		Name:          "Jane Doe",
		LastContacted: "15/01/2024",
		Notes:         "I don't think she's a real person",
	}
	jsonValue, _ := json.Marshal(newFriend)

	req, _ := http.NewRequest("POST", "/friends", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Jane Doe added successfully", resp["message"])

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM friends WHERE id = 1").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestDeleteFriendRoute(t *testing.T) {
	db, err := SetupTestDB()
	assert.NoError(t, err)
	defer db.Close()

	err = insertMockFriend(db, "1", "John Wick", "06/06/2023", "Nice guy")
	assert.NoError(t, err)

	router := gin.New()
	mockFriendsHandler := &handler.FriendsHandler{DB: db}
	router.DELETE("/friends/:id", mockFriendsHandler.DeleteFriend)

	req, _ := http.NewRequest("DELETE", "/friends/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Friend removed successfully", resp["message"])

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM friends WHERE id = ?", "1").Scan(&count)
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
	mockFriendsList := models.FriendsList{
		models.Friend{ID: "1", Name: "John Wick", LastContacted: "06/06/2023", Notes: "Nice guy"},
		models.Friend{ID: "2", Name: "Peter Parker", LastContacted: "12/12/2023", Notes: "I think he's Spiderman"},
	}

	db, err := SetupTestDB()
	assert.NoError(t, err)
	defer db.Close()

	err = insertMockFriend(db,
		mockFriend.ID,
		mockFriend.Name,
		mockFriend.LastContacted,
		mockFriend.Notes,
	)
	assert.NoError(t, err)

	router := gin.New()
	mockFriendsHandler := &handler.FriendsHandler{DB: db, FriendsList: mockFriendsList}
	router.PUT("/friends/:id", mockFriendsHandler.PutFriend)

	updatedFriend := models.Friend{
		Name:          "Master Chief",
		LastContacted: "15/01/2024",
		Notes:         "Doesn't talk much",
	}
	jsonValue, _ := json.Marshal(updatedFriend)

	req, _ := http.NewRequest("PUT", "/friends/1", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "1 updated successfully", resp["message"])

	var friend models.Friend
	err = db.QueryRow("SELECT name, lastContacted, notes FROM friends WHERE id = ?", "1").Scan(&friend.Name, &friend.LastContacted, &friend.Notes)
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
	mockFriendsList := models.FriendsList{
		models.Friend{ID: "1", Name: "John Wick", LastContacted: "06/06/2023", Notes: "Nice guy"},
		models.Friend{ID: "2", Name: "Peter Parker", LastContacted: "12/12/2023", Notes: "I think he's Spiderman"},
	}

	db, err := SetupTestDB()
	assert.NoError(t, err)
	defer db.Close()

	err = insertMockFriend(db,
		mockFriend.ID,
		mockFriend.Name,
		mockFriend.LastContacted,
		mockFriend.Notes,
	)
	assert.NoError(t, err)

	router := gin.New()
	mockFriendsHandler := &handler.FriendsHandler{DB: db, FriendsList: mockFriendsList}
	router.PUT("/friends/:id", mockFriendsHandler.PutFriend)

	updatedFriend := models.Friend{
		Notes: "Bro is Chuck Norris",
	}
	jsonValue, _ := json.Marshal(updatedFriend)

	req, _ := http.NewRequest("PUT", "/friends/1", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "1 updated successfully", resp["message"])

	var friend models.Friend
	err = db.QueryRow("SELECT name, lastContacted, notes FROM friends WHERE id = ?", "1").Scan(&friend.Name, &friend.LastContacted, &friend.Notes)
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
	mockFriendsList := models.FriendsList{
		models.Friend{ID: "1", Name: "John Wick", LastContacted: "06/06/2023", Notes: "Nice guy"},
		models.Friend{ID: "2", Name: "Peter Parker", LastContacted: "12/12/2023", Notes: "I think he's Spiderman"},
	}

	db, err := SetupTestDB()
	assert.NoError(t, err)
	defer db.Close()

	err = insertMockFriend(db,
		mockFriend.ID,
		mockFriend.Name,
		mockFriend.LastContacted,
		mockFriend.Notes,
	)
	assert.NoError(t, err)

	router := gin.New()
	mockFriendsHandler := &handler.FriendsHandler{DB: db, FriendsList: mockFriendsList}
	router.PUT("/friends/:id", mockFriendsHandler.PutFriend)

	updatedFriend := models.Friend{
		Name: "Winnie the Pooh",
	}
	jsonValue, _ := json.Marshal(updatedFriend)

	req, _ := http.NewRequest("PUT", "/friends/1", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "1 updated successfully", resp["message"])

	var friend models.Friend
	err = db.QueryRow("SELECT name, lastContacted, notes FROM friends WHERE id = ?", "1").Scan(&friend.Name, &friend.LastContacted, &friend.Notes)
	assert.NoError(t, err)
	assert.Equal(t, updatedFriend.Name, friend.Name)
	assert.Equal(t, mockFriend.LastContacted, friend.LastContacted)
	assert.Equal(t, mockFriend.Notes, friend.Notes)
}

// Tests PUT /friends/:id
// Tests the endpoint when only Name field is provided
func TestPutLastContactedOnly(t *testing.T) {
	mockFriend := models.Friend{
		ID:            "1",
		Name:          "John Wick",
		LastContacted: "06/06/2023",
		Notes:         "Nice guy",
	}
	mockFriendsList := models.FriendsList{
		models.Friend{ID: "1", Name: "John Wick", LastContacted: "06/06/2023", Notes: "Nice guy"},
		models.Friend{ID: "2", Name: "Peter Parker", LastContacted: "12/12/2023", Notes: "I think he's Spiderman"},
	}
	todaysDate := time.Now().Format("02/01/2006")

	db, err := SetupTestDB()
	assert.NoError(t, err)
	defer db.Close()

	err = insertMockFriend(db,
		mockFriend.ID,
		mockFriend.Name,
		mockFriend.LastContacted,
		mockFriend.Notes,
	)
	assert.NoError(t, err)

	router := gin.New()
	mockFriendsHandler := &handler.FriendsHandler{DB: db, FriendsList: mockFriendsList}
	router.PUT("/friends/:id", mockFriendsHandler.PutFriend)

	updatedFriend := models.Friend{
		LastContacted: todaysDate,
	}
	jsonValue, _ := json.Marshal(updatedFriend)

	req, _ := http.NewRequest("PUT", "/friends/1", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "1 updated successfully", resp["message"])

	var friend models.Friend
	err = db.QueryRow("SELECT name, lastContacted, notes FROM friends WHERE id = ?", "1").Scan(&friend.Name, &friend.LastContacted, &friend.Notes)
	assert.NoError(t, err)
	assert.Equal(t, mockFriend.Name, friend.Name)
	assert.Equal(t, todaysDate, friend.LastContacted)
	assert.Equal(t, mockFriend.Notes, friend.Notes)
}
