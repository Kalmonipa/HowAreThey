package main

import (
	"net/http"
	"net/http/httptest"
	"os"
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

func TestReadFile(t *testing.T) {
	readFileResult, err := buildFriendsList("test/friends_test.yaml")
	if err != nil {
		t.Fatalf("Failed to read or parse YAML: %v", err)
	}

	assert.Equal(t, friendsList, readFileResult)
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

// Tests the function that saves the FriendsList to the yaml
func TestSaveFriendsListToYAML(t *testing.T) {
	friends := FriendsList{
		{ID: "1", Name: "John Wick", LastContacted: "06/06/2023"},
		{ID: "2", Name: "Peter Parker", LastContacted: "12/12/2023"},
	}

	testFilePath := "temp_friends.yaml"

	// Call the function to save the list to a file
	err := saveFriendsListToYAML(friends, testFilePath)
	if err != nil {
		t.Fatalf("Failed to save friends list to YAML: %v", err)
	}

	// Clean up: defer the deletion of the test file
	defer os.Remove(testFilePath)

	// Read the file
	data, err := os.ReadFile(testFilePath)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// Check if the file contains the expected content
	expectedContent := `- id: "1"
  name: John Wick
  lastContacted: 06/06/2023
- id: "2"
  name: Peter Parker
  lastContacted: 12/12/2023
`
	assert.Equal(t, expectedContent, string(data))
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

	router := setupRouter(friendsHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/friends/list", nil)
	router.ServeHTTP(w, req)

	expectedResult := `[{"ID":"1","Name":"John Wick","LastContacted":"06/06/2023"},{"ID":"2","Name":"Peter Parker","LastContacted":"12/12/2023"}]`

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

// containsFriend checks if the given friend is in the friends list.
func containsFriend(friends FriendsList, friend Friend) bool {
	for _, f := range friends {
		if reflect.DeepEqual(f, friend) {
			return true
		}
	}
	return false
}
