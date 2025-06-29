package journal

import (
	"net/http"
	"self-improve-ai/internal/ai"
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
	journalGroup.GET("/:id", handler.GetEntry)
	journalGroup.GET("/streak", handler.GetStreak)
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

	// Fix: Check if record exists, not just for error
	var existingEntry JournalEntry
	err := h.DB.Where("user_id = ? AND entry_date = ?", userId, today).First(&existingEntry).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "You've already written a journal for today"})
		return
	}
	if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	entry := JournalEntry{
		UserId:    userId,
		Content:   input.Content,
		Mood:      input.Mood,
		EntryDate: today,
	}

	if input.GoalID != nil {
		entry.GoalID = input.GoalID
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

func (h *Handler) GetStreak(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var entries []JournalEntry
	err := h.DB.Where("user_id = ?", userID).
		Order("entry_date desc").
		Find(&entries).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load entries"})
		return
	}

	// Handle empty entries case
	if len(entries) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"current_streak": 0,
			"last_entry":     nil,
			"today_logged":   false,
		})
		return
	}

	today := time.Now().Truncate(24 * time.Hour)
	streak := 0
	expectedDate := today

	for _, e := range entries {
		entryDay := e.EntryDate.Truncate(24 * time.Hour)
		if entryDay.Equal(expectedDate) {
			streak++
			expectedDate = expectedDate.AddDate(0, 0, -1) // move back 1 day
		} else if entryDay.Before(expectedDate) {
			// Gap found - check if we should continue counting from yesterday
			if streak == 0 && entryDay.Equal(today.AddDate(0, 0, -1)) {
				// No entry today, but there's one yesterday - start streak from yesterday
				expectedDate = today.AddDate(0, 0, -1)
				continue
			}
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"current_streak": streak,
		"last_entry":     entries[0].EntryDate.Format("2006-01-02"),
		"today_logged":   len(entries) > 0 && entries[0].EntryDate.Truncate(24*time.Hour).Equal(today),
	})
}

func (h *Handler) GenerateAIFeedBack(c *gin.Context) {
	var entry JournalEntry
	userId := c.MustGet("user_id")
	id := c.Param("id")
	if err := h.DB.Where("user_id = ? and id = ?", userId, id).First(&entry).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
		return
	}

	summary, feedback, err := ai.GetJournalFeedback(entry.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI feedback failed"})
		return
	}
	entry.AI_Summary = summary
	entry.AI_Feedback = feedback

	h.DB.Save(&entry)
	c.JSON(http.StatusOK, entry)
}
