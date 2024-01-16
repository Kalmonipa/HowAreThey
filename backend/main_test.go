package main

import (
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var friendsList = FriendsList{
	Friend{ID: "1", Name: "John Wick", LastContacted: "06/06/2023", Notes: "Nice guy"},
	Friend{ID: "2", Name: "Peter Parker", LastContacted: "12/12/2023", Notes: "I think he's Spiderman"},
}

var futureFriendsList = FriendsList{
	Friend{ID: "3", Name: "Doctor Who", LastContacted: "25/12/2070", Notes: "Lives in a phonebox"},
}

var friendsHandler = &FriendsHandler{FriendsList: friendsList}

// setupTestDB creates and returns a new database for testing
func setupTestDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	// Create the friends table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS friends (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		lastContacted TEXT NOT NULL,
		notes TEXT NOT NULL
	);`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestPickRandom(t *testing.T) {
	for i := 0; i < 10; i++ {
		friend, err := pickRandomFriend(friendsList)
		if err != nil {
			t.Errorf("RandomFriend returned an error: %v", err)
		}

		if !containsFriend(friendsList, friend) {
			t.Errorf("Chosen friend %+v not found in the friends list", friend)
		}
	}

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
	friend := Friend{ID: "1", Name: "Jimmy Neutron", LastContacted: "10/10/2023", Notes: "Pretty smart kid"}
	todaysDate := time.Date(2023, time.December, 31, 0, 0, 0, 0, time.Local)

	expectedResult := Friend{ID: "1", Name: "Jimmy Neutron", LastContacted: "31/12/2023", Notes: "Pretty smart kid"}

	updatedFriend := updateLastContacted(friend, todaysDate)

	assert.Equal(t, updatedFriend, expectedResult)
}

func TestListFriendsNames(t *testing.T) {
	expectedResult := []string{"John Wick", "Peter Parker"}
	unexpectedResult := []string{"John Wick", "Peter Parker", "Shouldn't Exist"}

	assert.Equal(t, expectedResult, listFriendsNames(friendsList))
	assert.NotEqual(t, unexpectedResult, listFriendsNames(friendsList))
}

func TestAddFriend(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	defer db.Close()

	newFriend := Friend{
		Name:          "Zark Muckerberg",
		LastContacted: "15/01/2024",
		Notes:         "Definitely a lizard person",
	}

	// Test the addFriend function
	err = addFriend(db, newFriend)
	assert.NoError(t, err)

	// Verify that the friend was added
	var friendCount int
	err = db.QueryRow("SELECT COUNT(*) FROM friends WHERE id = 1").Scan(&friendCount)
	assert.NoError(t, err)
	assert.Equal(t, 1, friendCount, "Expected new friend to be added")
}

func TestDeleteFriend(t *testing.T) {

	gin.SetMode(gin.TestMode)
	db, err := setupTestDB()
	assert.NoError(t, err)
	defer db.Close()

	err = insertMockFriend(db, "1", "John Wick", "06/06/2023", "Nice guy")
	assert.NoError(t, err)
	err = insertMockFriend(db, "2", "Jack Reacher", "06/06/2023", "Must be on steroids")
	assert.NoError(t, err)

	err = deleteFriend(db, "2")
	assert.NoError(t, err)

	var friendCount int
	err = db.QueryRow("SELECT COUNT(*) FROM friends").Scan(&friendCount)
	assert.NoError(t, err)
	assert.Equal(t, 1, friendCount, "Expected new friend to be deleted")
}

func TestUpdateFriend(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := setupTestDB()
	assert.NoError(t, err)
	defer db.Close()

	err = insertMockFriend(db, "1", "John Wick", "06/06/2023", "Nice guy")
	assert.NoError(t, err)

	updatedFriend := Friend{
		Name:          "John Wick",
		LastContacted: "10/01/2024",
		Notes:         "Nice guy",
	}

	err = updateFriend(db, "1", updatedFriend)
	assert.NoError(t, err)

	var friend Friend
	err = db.QueryRow("SELECT name, lastContacted, notes FROM friends WHERE id = ?", "1").Scan(&friend.Name, &friend.LastContacted, &friend.Notes)
	assert.NoError(t, err)
	assert.Equal(t, updatedFriend.Name, friend.Name)
	assert.Equal(t, updatedFriend.LastContacted, friend.LastContacted)
	assert.Equal(t, updatedFriend.Notes, friend.Notes)
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

// insertMockFriend inserts a mock friend into the test database.
func insertMockFriend(db *sql.DB, id string, name string, lastContacted string, notes string) error {
	stmt, err := db.Prepare("INSERT INTO friends (id, name, lastContacted, notes) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id, name, lastContacted, notes)
	return err
}
