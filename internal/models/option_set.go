package models

type OptionSet struct {
	BaseModel
	Category    string `gorm:"size:50;not null;index" json:"category"`
	Code        string `gorm:"size:20;not null" json:"code"`
	Label       string `gorm:"size:200;not null" json:"label"`
	Description string `gorm:"size:500" json:"description,omitempty"`
	SortOrder   int    `gorm:"default:0" json:"sort_order"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
}
