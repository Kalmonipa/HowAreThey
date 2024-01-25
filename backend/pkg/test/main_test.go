package test

import (
	"database/sql"
	"howarethey/pkg/models"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// setupTestDB creates and returns a new database for testing
func SetupTestDB() (*sql.DB, error) {
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
	mockFriendsList := models.FriendsList{
		models.Friend{ID: "1", Name: "John Wick", LastContacted: "06/06/2023", Notes: "Nice guy"},
		models.Friend{ID: "2", Name: "Peter Parker", LastContacted: "12/12/2023", Notes: "I think he's Spiderman"},
	}

	for i := 0; i < 10; i++ {
		friend, err := models.PickRandomFriend(mockFriendsList)
		assert.NoError(t, err)

		if !containsFriend(mockFriendsList, friend) {
			t.Errorf("Chosen friend %+v not found in the friends list", friend)
		}
	}

	emptyFriends := models.FriendsList{}
	_, err := models.PickRandomFriend(emptyFriends)
	assert.Error(t, err)
}

func TestCalculateWeight(t *testing.T) {
	mockFriend := models.Friend{
		ID: "1", Name: "John Wick", LastContacted: "06/06/2023", Notes: "Nice guy",
	}

	expectedWeight := 200

	todaysDate := time.Date(2023, time.December, 23, 0, 0, 0, 0, time.UTC)

	weight, err := models.CalculateWeight(mockFriend.LastContacted, todaysDate)
	assert.NoError(t, err)

	assert.Equal(t, expectedWeight, weight)
}

func TestCalculateWeightFromFuture(t *testing.T) {
	futureFriend := models.Friend{
		ID:            "3",
		Name:          "Doctor Who",
		LastContacted: "25/12/2070",
		Notes:         "Lives in a phonebox",
	}

	todaysDate := time.Date(2070, time.December, 23, 0, 0, 0, 0, time.UTC)

	weight, err := models.CalculateWeight(futureFriend.LastContacted, todaysDate)

	assert.Equal(t, weight, 0)
	assert.NotNil(t, err)
}

func TestUpdateLastContact(t *testing.T) {
	mockFriend := models.Friend{
		ID: "1", Name: "John Wick", LastContacted: "06/06/2023", Notes: "Nice guy",
	}

	todaysDate := time.Date(2023, time.December, 31, 0, 0, 0, 0, time.Local)

	expectedResult := models.Friend{
		ID:            mockFriend.ID,
		Name:          mockFriend.Name,
		LastContacted: "31/12/2023",
		Notes:         mockFriend.Notes,
	}

	updatedFriend := models.UpdateLastContacted(mockFriend, todaysDate)

	assert.Equal(t, &expectedResult, updatedFriend)
}

func TestListFriendsNames(t *testing.T) {
	mockFriendsList := models.FriendsList{
		models.Friend{ID: "1", Name: "John Wick", LastContacted: "06/06/2023", Notes: "Nice guy"},
		models.Friend{ID: "2", Name: "Peter Parker", LastContacted: "12/12/2023", Notes: "I think he's Spiderman"},
	}

	expectedResult := []string{"John Wick", "Peter Parker"}
	unexpectedResult := []string{"John Wick", "Peter Parker", "Shouldn't Exist"}

	assert.Equal(t, expectedResult, models.ListFriendsNames(mockFriendsList))
	assert.NotEqual(t, unexpectedResult, models.ListFriendsNames(mockFriendsList))
}

func TestAddFriend(t *testing.T) {
	db, err := SetupTestDB()
	assert.NoError(t, err)
	defer db.Close()

	newFriend := models.Friend{
		Name:          "Zark Muckerberg",
		LastContacted: "15/01/2024",
		Notes:         "Definitely a lizard person",
	}

	err = models.AddFriend(db, newFriend)
	assert.NoError(t, err)

	var friendCount int
	err = db.QueryRow("SELECT COUNT(*) FROM friends WHERE id = 1").Scan(&friendCount)
	assert.NoError(t, err)
	assert.Equal(t, 1, friendCount, "Expected new friend to be added")
}

func TestDeleteFriend(t *testing.T) {

	db, err := SetupTestDB()
	assert.NoError(t, err)
	defer db.Close()

	err = insertMockFriend(db, "1", "John Wick", "06/06/2023", "Nice guy")
	assert.NoError(t, err)
	err = insertMockFriend(db, "2", "Jack Reacher", "06/06/2023", "Must be on steroids")
	assert.NoError(t, err)

	err = models.DeleteFriend(db, "2")
	assert.NoError(t, err)

	var friendCount int
	err = db.QueryRow("SELECT COUNT(*) FROM friends").Scan(&friendCount)
	assert.NoError(t, err)
	assert.Equal(t, 1, friendCount, "Expected new friend to be deleted")
}

func TestSqlUpdateFriend(t *testing.T) {

	db, err := SetupTestDB()
	assert.NoError(t, err)
	defer db.Close()

	err = insertMockFriend(db, "1", "John Wick", "06/06/2023", "Nice guy")
	assert.NoError(t, err)

	updatedFriend := models.Friend{
		Name:          "John Wick",
		LastContacted: "10/01/2024",
		Notes:         "Nice guy",
	}

	err = models.SqlUpdateFriend(db, "1", &updatedFriend)
	assert.NoError(t, err)

	var friend models.Friend
	err = db.QueryRow("SELECT name, lastContacted, notes FROM friends WHERE id = ?", "1").Scan(&friend.Name, &friend.LastContacted, &friend.Notes)
	assert.NoError(t, err)
	assert.Equal(t, updatedFriend.Name, friend.Name)
	assert.Equal(t, updatedFriend.LastContacted, friend.LastContacted)
	assert.Equal(t, updatedFriend.Notes, friend.Notes)
}

// containsFriend checks if the given friend is in the friends list.
func containsFriend(friends models.FriendsList, friend models.Friend) bool {
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
