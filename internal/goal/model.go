package goal

import "gorm.io/gorm"

type Goal struct {
	gorm.Model
	UserID      uint   `json:"-"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`    // fitness, learning, etc.
	Progress    int    `json:"progress"`    // 0–100
	TargetDate  string `json:"target_date"` // ISO format
}
