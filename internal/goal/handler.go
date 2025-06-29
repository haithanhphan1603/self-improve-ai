package goal

import (
	"net/http"
	"self-improve-ai/pkg/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	goalGroup := r.Group("/api/goals")
	handler := &Handler{DB: db}
	goalGroup.Use(middleware.AuthMiddleware())
	goalGroup.POST("/", handler.CreateGoal)
	goalGroup.GET("/", handler.ListGoals)
	goalGroup.PUT("/:id", handler.UpdateGoal)
	goalGroup.DELETE("/:id", handler.DeleteGoal)
}

type GoalInput struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Category    string `json:"category"`
	TargetDate  string `json:"target_date"`
}

func (h *Handler) CreateGoal(c *gin.Context) {
	var input GoalInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.MustGet("user_id").(uint)
	goal := Goal{
		UserID:      userId,
		Title:       input.Title,
		Description: input.Description,
		Category:    input.Category,
		TargetDate:  input.TargetDate,
		Progress:    0,
	}
	if err := h.DB.Create(&goal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create goal"})
		return
	}
	c.JSON(http.StatusCreated, goal)
}

func (h *Handler) ListGoals(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var goals []Goal
	if err := h.DB.Where("user_id = ?", userID).Find(&goals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch goals"})
		return
	}

	c.JSON(http.StatusOK, goals)
}

type UpdateInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
	Progress    *int    `json:"progress"` // Optional update
	TargetDate  *string `json:"target_date"`
}

func (h *Handler) UpdateGoal(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("user_id").(uint)

	var goal Goal
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&goal).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Goal not found"})
		return
	}

	var input UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Title != nil {
		goal.Title = *input.Title
	}
	if input.Description != nil {
		goal.Description = *input.Description
	}
	if input.Category != nil {
		goal.Category = *input.Category
	}
	if input.TargetDate != nil {
		goal.TargetDate = *input.TargetDate
	}
	if input.Progress != nil {
		goal.Progress = *input.Progress
	}

	if err := h.DB.Save(&goal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update goal"})
		return
	}

	c.JSON(http.StatusOK, goal)
}

func (h *Handler) DeleteGoal(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("user_id").(uint)

	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&Goal{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete goal"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Goal deleted"})
}
