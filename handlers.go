package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type FriendsHandler struct {
	FriendsList FriendsList
}

func NewFriendsHandler(friendsList FriendsList) *FriendsHandler {
	return &FriendsHandler{
		FriendsList: friendsList,
	}
}

// GET /friends/list
func (h *FriendsHandler) GetFriends(c *gin.Context) {
	//friendsNames := ListFriendsNames(h.FriendsList)
	c.JSON(http.StatusOK, h.FriendsList)
}

// GET /friends/random
func (h *FriendsHandler) GetRandomFriend(c *gin.Context) {
	randomFriend, err := pickRandomFriend(h.FriendsList)
	if err != nil {
		LogMessage(LogLevelFatal, "error: %v", err)
		c.JSON(http.StatusNotFound, "failed to pick a friend")
	}

	c.JSON(http.StatusOK, randomFriend)
}
