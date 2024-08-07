package test

import (
	"database/sql"
	"howarethey/pkg/handler"
	"howarethey/pkg/logger"
	"howarethey/pkg/models"
	"os"
	"reflect"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

var (
	// Mock friends list used in the tests
	mockFriendsList = models.FriendsList{
		models.Friend{ID: "1", Name: "John Wick", LastContacted: "2023-06-06", Birthday: "1996-02-23", Notes: "Nice guy"},
		models.Friend{ID: "2", Name: "Peter Parker", LastContacted: "2023-12-12", Birthday: "1996-10-20", Notes: "I think he's Spiderman"},
	}
)

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

func TestPickRandom(t *testing.T) {
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
	mockFriend := mockFriendsList[0]

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
		LastContacted: "2070-12-25",
		Birthday:      "1996-02-23",
		Notes:         "Lives in a phonebox",
	}

	todaysDate := time.Date(2070, time.December, 23, 0, 0, 0, 0, time.UTC)

	weight, err := models.CalculateWeight(futureFriend.LastContacted, todaysDate)

	assert.Equal(t, weight, 0)
	assert.NotNil(t, err)
}

func TestCheckBirthday(t *testing.T) {
	os.Setenv("TEST_ENV", "true")
	logger.SetupLogger()

	mockTodaysDate := time.Date(2020, time.February, 23, 0, 0, 0, 0, time.UTC)

	result := models.CheckBirthdays(mockFriendsList, mockTodaysDate)

	assert.Equal(t, 1, len(result))
	assert.Equal(t, "John Wick", result[0].Name)
}

func TestCheckBirthdayNoResults(t *testing.T) {
	mockTodaysDate := time.Date(2020, time.January, 10, 0, 0, 0, 0, time.UTC)

	result := models.CheckBirthdays(mockFriendsList, mockTodaysDate)

	assert.Equal(t, len(result), 0)
}

func TestUpdateLastContact(t *testing.T) {
	mockFriend := mockFriendsList[0]
	todaysDate := time.Date(2023, time.December, 31, 0, 0, 0, 0, time.Local)

	expectedResult := models.Friend{
		ID:            mockFriend.ID,
		Name:          mockFriend.Name,
		LastContacted: "2023-12-31",
		Birthday:      "1996-02-23",
		Notes:         mockFriend.Notes,
	}

	updatedFriend := models.UpdateLastContacted(mockFriend, todaysDate)

	assert.Equal(t, &expectedResult, updatedFriend)
}

func TestListFriendsNames(t *testing.T) {
	expectedResult := []string{"John Wick", "Peter Parker"}
	unexpectedResult := []string{"John Wick", "Peter Parker", "Shouldn't Exist"}

	assert.Equal(t, expectedResult, models.ListFriendsNames(mockFriendsList))
	assert.NotEqual(t, unexpectedResult, models.ListFriendsNames(mockFriendsList))
}

func TestAddFriend(t *testing.T) {
	os.Setenv("TEST_ENV", "true")
	logger.SetupLogger()

	db, err := setupTestDB()
	assert.NoError(t, err)
	defer db.Close()

	newFriend := models.Friend{
		Name:          "Zark Muckerberg",
		LastContacted: "2024-01-15",
		Birthday:      "1996-02-23",
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
	db, err := setupTestDB()
	assert.NoError(t, err)
	defer db.Close()

	friendOne := mockFriendsList[0]
	friendTwo := mockFriendsList[1]

	err = insertMockFriend(db, friendOne.ID, friendOne.Name, friendOne.LastContacted, friendOne.Birthday, friendOne.Notes)
	assert.NoError(t, err)
	err = insertMockFriend(db, friendTwo.ID, friendTwo.Name, friendTwo.LastContacted, friendTwo.Birthday, friendTwo.Notes)
	assert.NoError(t, err)

	err = models.DeleteFriend(db, friendTwo)
	assert.NoError(t, err)

	var friendCount int
	err = db.QueryRow("SELECT COUNT(*) FROM friends").Scan(&friendCount)
	assert.NoError(t, err)
	assert.Equal(t, 1, friendCount, "Expected new friend to be deleted")
}

func TestSqlUpdateFriend(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	defer db.Close()

	err = insertMockFriend(db, "1", "John Wick", "2023-06-06", "1996-02-23", "Nice guy")
	assert.NoError(t, err)

	updatedFriend := models.Friend{
		Name:          "John Wick",
		LastContacted: "2024-01-10",
		Birthday:      "1996-02-23",
		Notes:         "Nice guy",
	}

	err = models.SqlUpdateFriend(db, "1", &updatedFriend)
	assert.NoError(t, err)

	var friend models.Friend
	err = db.QueryRow("SELECT name, lastContacted, birthday, notes FROM friends WHERE id = ?", "1").Scan(&friend.Name, &friend.LastContacted, &friend.Birthday, &friend.Notes)
	assert.NoError(t, err)
	assert.Equal(t, updatedFriend.Name, friend.Name)
	assert.Equal(t, updatedFriend.LastContacted, friend.LastContacted)
	assert.Equal(t, updatedFriend.Notes, friend.Notes)
}

func TestIsValidDate(t *testing.T) {
	assert.True(t, handler.IsValidDate("2024-03-02"))
}

func TestInvalidDate(t *testing.T) {
	assert.False(t, handler.IsValidDate("02/03/2024"))
	assert.False(t, handler.IsValidDate("24.04.05"))
	assert.False(t, handler.IsValidDate("invalid"))
	assert.False(t, handler.IsValidDate("this should fail"))
}

func containsFriend(friends models.FriendsList, friend models.Friend) bool {
	for _, f := range friends {
		if reflect.DeepEqual(f, friend) {
			return true
		}
	}
	return false
}

func insertMockFriend(db *sql.DB, id string, name string, lastContacted string, birthday string, notes string) error {
	stmt, err := db.Prepare("INSERT INTO friends (id, name, lastContacted, birthday, notes) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id, name, lastContacted, birthday, notes)
	return err
}
