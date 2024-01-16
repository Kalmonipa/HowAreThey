package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var friendsList = FriendsList{
	Friend{ID: "1", Name: "John Wick", LastContacted: "06/06/2023"},
	Friend{ID: "2", Name: "Peter Parker", LastContacted: "12/12/2023"},
}

var futureFriendsList = FriendsList{
	Friend{ID: "3", Name: "Doctor Who", LastContacted: "25/12/2070"},
}

var friendsHandler = &FriendsHandler{FriendsList: friendsList}

// setupTestDB creates and returns a new database for testing
func setupTestDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:") // Use in-memory database for tests
	if err != nil {
		return nil, err
	}

	// Create the friends table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS friends (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		lastContacted TEXT NOT NULL
	);`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestPickRandom(t *testing.T) {
	// Call pickRandomFriend a few times to ensure it doesn't return an out-of-bounds error or panic
	for i := 0; i < 10; i++ {
		friend, err := pickRandomFriend(friendsList)
		if err != nil {
			t.Errorf("RandomFriend returned an error: %v", err)
		}

		// Checks to make sure that the friend that gets returned is in the FriendsList
		if !containsFriend(friendsList, friend) {
			t.Errorf("Chosen friend %+v not found in the friends list", friend)
		}
	}

	// Test with an empty friends list
	emptyFriends := FriendsList{}
	_, err := pickRandomFriend(emptyFriends)
	if err == nil {
		t.Errorf("RandomFriend should return an error when the slice is empty")
	}
}

func TestCalculateWeight(t *testing.T) {
	expectedWeight := 200

	todaysDate := time.Date(2023, time.December, 23, 0, 0, 0, 0, time.UTC)

	weight, err := calculateWeight(friendsList[0].LastContacted, todaysDate)
	if err != nil {
		t.Errorf("Failed to calculate the weight of %s", friendsList[0].Name)
	}

	assert.Equal(t, expectedWeight, weight)
}

func TestCalculateWeightFromFuture(t *testing.T) {
	todaysDate := time.Date(2070, time.December, 23, 0, 0, 0, 0, time.UTC)

	weight, err := calculateWeight(futureFriendsList[0].LastContacted, todaysDate)

	assert.Equal(t, weight, 0)
	assert.NotNil(t, err)
}

func TestUpdateLastContact(t *testing.T) {
	friend := Friend{ID: "1", Name: "Jimmy Neutron", LastContacted: "10/10/2023"}
	todaysDate := time.Date(2023, time.December, 31, 0, 0, 0, 0, time.Local)

	expectedResult := Friend{ID: "1", Name: "Jimmy Neutron", LastContacted: "31/12/2023"}

	updatedFriend := updateLastContacted(friend, todaysDate)

	assert.Equal(t, updatedFriend, expectedResult)
}

// Tests the function that grabs the names of the friends list supplied
func TestListFriendsNames(t *testing.T) {
	friends := FriendsList{
		{ID: "1", Name: "John Wick", LastContacted: "06/06/2023"},
		{ID: "2", Name: "Peter Parker", LastContacted: "12/12/2023"},
	}

	expectedResult := []string{"John Wick", "Peter Parker"}
	unexpectedResult := []string{"John Wick", "Peter Parker", "Shouldn't Exist"}

	assert.Equal(t, expectedResult, listFriendsNames(friends))
	assert.NotEqual(t, unexpectedResult, listFriendsNames(friends))
}

// Testing the routes
func TestFriendsCountRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := setupRouter(friendsHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/friends/count", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "2", w.Body.String())
}

func TestFriendsListRoute(t *testing.T) {

	gin.SetMode(gin.TestMode)

	// Setup test database
	db, err := setupTestDB()
	assert.NoError(t, err)
	defer db.Close()

	// Insert mock data
	_, err = db.Exec(`INSERT INTO friends (id, name, lastContacted) VALUES
	(1, 'John Wick', '06/06/2023'),
	(2, 'Jack Reacher', '06/06/2023');`)
	assert.NoError(t, err)

	// Setup router with test DB
	friendsList, err := buildFriendsList(db)
	assert.NoError(t, err)
	friendsHandler := NewFriendsHandler(friendsList, db)
	router := setupRouter(friendsHandler)

	// Perform the test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/friends", nil)
	router.ServeHTTP(w, req)

	expectedResult := `[{"ID":"1","Name":"John Wick","LastContacted":"06/06/2023"},{"ID":"2","Name":"Jack Reacher","LastContacted":"06/06/2023"}]`

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, expectedResult, w.Body.String())
}

func TestFriendIDRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := setupRouter(friendsHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/friends/id/1", nil)
	router.ServeHTTP(w, req)

	expectedResult := `{"ID":"1","Name":"John Wick","LastContacted":"06/06/2023"}`

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, expectedResult, w.Body.String())
}

func TestFriendNameRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := setupRouter(friendsHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/friends/name/john-wick", nil)
	router.ServeHTTP(w, req)

	expectedResult := `{"ID":"1","Name":"John Wick","LastContacted":"06/06/2023"}`

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, expectedResult, w.Body.String())
}

func TestMissingFriendIDRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := setupRouter(friendsHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/friends/id/100", nil)
	router.ServeHTTP(w, req)

	expectedResult := `{"error":"friend not found"}`

	assert.Equal(t, 404, w.Code)
	assert.Equal(t, expectedResult, w.Body.String())
}

func TestAddFriend(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	defer db.Close()

	newFriend := Friend{
		ID:            "3",
		Name:          "Zark Muckerberg",
		LastContacted: "15/01/2024",
	}

	// Test the addFriend function
	err = addFriend(db, newFriend)
	assert.NoError(t, err)

	// Verify that the friend was added
	var friendCount int
	err = db.QueryRow("SELECT COUNT(*) FROM friends WHERE id = ?", newFriend.ID).Scan(&friendCount)
	assert.NoError(t, err)
	assert.Equal(t, 1, friendCount, "Expected new friend to be added")
}

// Tests that the deleteFriend function removes a friend based on the ID provided
func TestDeleteFriend(t *testing.T) {

	gin.SetMode(gin.TestMode)
	db, err := setupTestDB()
	assert.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`INSERT INTO friends (id, name, lastContacted) VALUES
	(1, 'John Wick', '06/06/2023'),
	(2, 'Jack Reacher', '06/06/2023');`)
	assert.NoError(t, err)

	err = deleteFriend(db, "2")
	assert.NoError(t, err)

	var friendCount int
	err = db.QueryRow("SELECT COUNT(*) FROM friends").Scan(&friendCount)
	assert.NoError(t, err)
	assert.Equal(t, 1, friendCount, "Expected new friend to be deleted")
}

// containsFriend checks if the given friend is in the friends list.
func containsFriend(friends FriendsList, friend Friend) bool {
	for _, f := range friends {
		if reflect.DeepEqual(f, friend) {
			return true
		}
	}
	return false
}
