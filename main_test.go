package main

import (
	"reflect"
	"testing"

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

// containsFriend checks if the given friend is in the friends list.
func containsFriend(friends FriendsList, friend Friend) bool {
	for _, f := range friends {
		if reflect.DeepEqual(f, friend) {
			return true
		}
	}
	return false
}
