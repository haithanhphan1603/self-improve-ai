package journal

import (
	"time"

	"gorm.io/gorm"
)

type JournalEntry struct {
	gorm.Model
	UserId      uint   `json:"-"`
	Content     string `json:"content"`
	Mood        string `json:"mood"`
	AI_Summary  string `json:"ai_summary"`
	AI_Feedback string `json:"ai_feedback"`
	GoalID      *uint  `json:"goal_id"`
	// 📅 Ensure 1 journal per day
	EntryDate time.Time `json:"entry_date" gorm:"index"` // YYYY-MM-DD
}
