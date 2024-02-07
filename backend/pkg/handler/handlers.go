package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"os"
	"regexp"
	"time"

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

	// Allow CORS for the frontend to access
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}

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

func isValidDate(dateStr string) bool {
	match, _ := regexp.MatchString(`^\d{2}/\d{2}/\d{4}$`, dateStr)
	if !match {
		return false
	}

	_, err := time.Parse("02/01/2006", dateStr)
	return err == nil
}

// DELETE /friends/:id
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
	url := os.Getenv("DISCORD_WEBHOOK")

	randomFriend, err := models.PickRandomFriend(h.FriendsList)
	if err != nil {
		logger.LogMessage(logger.LogLevelFatal, "Failed to get a random friend: %v", err)
		c.JSON(http.StatusNotFound, "failed to pick a friend")
	}

	logger.LogMessage(logger.LogLevelInfo, randomFriend.Name+"has been chosen")

	if url != "" {
		models.SendNotification(randomFriend, url)
	}

	updatedFriend := models.UpdateLastContacted(randomFriend, time.Now())

	err = models.SqlUpdateFriend(h.DB, updatedFriend.ID, updatedFriend)
	if err != nil {
		logger.LogMessage(logger.LogLevelFatal, "Failed to update friend: %v", err)
		c.JSON(http.StatusNotFound, "failed to update a friend")
	}

	h.FriendsList, err = models.UpdateFriend(h.FriendsList, updatedFriend)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, randomFriend)
}

// GET /friends/count
func (h *FriendsHandler) GetFriendCount(c *gin.Context) {
	c.JSON(http.StatusOK, models.GetFriendCount(h.FriendsList))
}

// GET /friends/id/:id
func (h *FriendsHandler) GetFriendByID(c *gin.Context) {
	friendID := c.Param("id")
	friend, err := models.GetFriendByID(friendID, h.FriendsList)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, friend)
}

// GET /friends/name/:name
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

// PUT /friends/:id
func (h *FriendsHandler) PutFriend(c *gin.Context) {
	id := c.Param("id")

	currentFriend, err := models.GetFriendByID(id, h.FriendsList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Friend not found"})
		return
	}

	var updatedFriend models.Friend
	if err := c.ShouldBindJSON(&updatedFriend); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if updatedFriend.Name != "" {
		logger.LogMessage(logger.LogLevelDebug, "Setting name to "+updatedFriend.Name)
		currentFriend.Name = updatedFriend.Name
	}
	updatedLastContacted := updatedFriend.LastContacted
	if updatedLastContacted != "" {
		if !isValidDate(updatedLastContacted) {
			err = errors.New("Date must be in DD/MM/YYYY format." + updatedLastContacted + "does not match.")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		logger.LogMessage(logger.LogLevelDebug, "Setting last contacted to "+updatedFriend.LastContacted)
		currentFriend.LastContacted = updatedFriend.LastContacted
	}
	if updatedFriend.Notes != "" {
		logger.LogMessage(logger.LogLevelDebug, "Setting notes to "+updatedFriend.Notes)
		currentFriend.Notes = updatedFriend.Notes
	}

	if err := models.SqlUpdateFriend(h.DB, id, currentFriend); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.FriendsList, err = models.BuildFriendsList(h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	successMsg := id + " updated successfully"

	c.JSON(http.StatusOK, gin.H{"message": successMsg})
}
