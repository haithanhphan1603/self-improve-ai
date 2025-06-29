package journal

import (
	"net/http"
	"self-improve-ai/pkg/middleware"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	journalGroup := r.Group("/api/journals")
	handler := &Handler{DB: db}
	journalGroup.Use(middleware.AuthMiddleware())
	journalGroup.POST("/", handler.CreateEntry)
	journalGroup.GET("/", handler.ListEntries)
	journalGroup.POST("/:id", handler.GetEntry)
}

type JournalInput struct {
	Content string `json:"content" binding:"required"`
	Mood    string `json:"mood"`
	GoalID  *uint  `json:"goal_id"` // optional
}

func (h *Handler) CreateEntry(c *gin.Context) {
	userId := c.MustGet("user_id").(uint)
	var input JournalInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	today := time.Now().Truncate(24 * time.Hour)
	err := h.DB.Where("user_id = ? and entry_date = ?", userId, today).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "You’ve already written a journal for today"})
		return
	}
	entry := JournalEntry{
		UserId:    userId,
		Content:   input.Content,
		Mood:      input.Mood,
		EntryDate: today,
	}

	if err := h.DB.Create(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create journal entry"})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

func (h *Handler) ListEntries(c *gin.Context) {
	userId := c.MustGet("user_id").(uint)

	var entries []JournalEntry
	if err := h.DB.Where("user_id = ?", userId).Order("created_at desc").Find(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch entries"})
		return
	}

	c.JSON(http.StatusOK, entries)
}

func (h *Handler) GetEntry(c *gin.Context) {
	var entry JournalEntry
	userId := c.MustGet("user_id").(uint)
	id := c.Param("id")
	if err := h.DB.Where("id = ? and user_id = ?", id, userId).First(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch entry"})
		return
	}

	c.JSON(http.StatusOK, entry)
}
