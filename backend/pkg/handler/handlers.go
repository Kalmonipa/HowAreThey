package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"howarethey/pkg/logger"
	"howarethey/pkg/models"
)

type FriendsHandler struct {
	FriendsList models.FriendsList
	DB          *sql.DB
}

func NewFriendsHandler(friendsList models.FriendsList, db *sql.DB) *FriendsHandler {
	return &FriendsHandler{
		FriendsList: friendsList,
		DB:          db,
	}
}

func SetupRouter(handler *FriendsHandler) *gin.Engine {

	gin.SetMode(gin.ReleaseMode)

	// TODO: Figure out if .Default() is what I need or something else
	r := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	// You can also add more settings here, like config.AllowMethods, config.AllowHeaders, etc.

	r.Use(cors.New(config))

	r.DELETE("/friends/:id", handler.DeleteFriend)
	r.GET("/friends", handler.GetFriends)
	r.GET("/friends/random", handler.GetRandomFriend)
	r.GET("/friends/count", handler.GetFriendCount)
	r.GET("/friends/id/:id", handler.GetFriendByID)
	r.GET("/friends/name/:name", handler.GetFriendByName)
	r.POST("/friends", handler.PostNewFriend)
	r.PUT("/friends/:id", handler.PutFriend)

	return r
}

// DELETE /friends/{ID}
func (h *FriendsHandler) DeleteFriend(c *gin.Context) {
	friendID := c.Param("id")

	err := models.DeleteFriend(h.DB, friendID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	friendsList, err := models.BuildFriendsList(h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.FriendsList = friendsList

	c.JSON(http.StatusCreated, gin.H{"message": "Friend removed successfully"})
}

// GET /friends
func (h *FriendsHandler) GetFriends(c *gin.Context) {
	c.JSON(http.StatusOK, h.FriendsList)
}

// GET /friends/random
func (h *FriendsHandler) GetRandomFriend(c *gin.Context) {
	randomFriend, err := models.PickRandomFriend(h.FriendsList)
	if err != nil {
		logger.LogMessage(logger.LogLevelFatal, "Failed to get a random friend: %v", err)
		c.JSON(http.StatusNotFound, "failed to pick a friend")
	}

	c.JSON(http.StatusOK, randomFriend)
}

// GET /friends/count
func (h *FriendsHandler) GetFriendCount(c *gin.Context) {
	c.JSON(http.StatusOK, models.GetFriendCount(h.FriendsList))
}

// GET /friends/id/{ID}
func (h *FriendsHandler) GetFriendByID(c *gin.Context) {
	friendID := c.Param("id")
	friend, err := models.GetFriendByID(friendID, h.FriendsList)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, friend)
}

// GET /friends/name/{NAME}
func (h *FriendsHandler) GetFriendByName(c *gin.Context) {
	friendName := c.Param("name")
	friend, err := models.GetFriendByName(friendName, h.FriendsList)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, friend)
}

// POST /friends
func (h *FriendsHandler) PostNewFriend(c *gin.Context) {
	var newFriend models.Friend
	if err := c.ShouldBindJSON(&newFriend); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := models.AddFriend(h.DB, newFriend)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	friendsList, err := models.BuildFriendsList(h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.FriendsList = friendsList

	successMsg := newFriend.Name + " added successfully"

	c.JSON(http.StatusCreated, gin.H{"message": successMsg})
}

// PUT /friends/{ID}
func (h *FriendsHandler) PutFriend(c *gin.Context) {
	id := c.Param("id")

	var updatedFriend models.Friend
	if err := c.ShouldBindJSON(&updatedFriend); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := models.UpdateFriend(h.DB, id, updatedFriend); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	friendsList, err := models.BuildFriendsList(h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.FriendsList = friendsList

	successMsg := updatedFriend.ID + " updated successfully"

	c.JSON(http.StatusOK, gin.H{"message": successMsg})
}
