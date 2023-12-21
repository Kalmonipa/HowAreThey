package main

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var friendsList = FriendsList{
	Friend{Name: "John Wick", LastContacted: "06/06/2023"},
	Friend{Name: "Peter Parker", LastContacted: "12/12/2023"},
	Friend{Name: "The Grinch", LastContacted: "25/12/2022"},
}

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

// containsFriend checks if the given friend is in the friends list.
func containsFriend(friends FriendsList, friend Friend) bool {
	for _, f := range friends {
		if reflect.DeepEqual(f, friend) {
			return true
		}
	}
	return false
}
