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
func (h *FriendsHandler) ListFriendsNames(c *gin.Context) {
	friendsNames := ListFriendsNames(h.FriendsList)
	c.JSON(http.StatusOK, friendsNames)
}
