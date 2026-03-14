package models

type Facility struct {
	BaseModel
	Name            string `gorm:"size:200;not null" json:"name"`
	Code            string `gorm:"size:20;uniqueIndex;not null" json:"code"`
	Level           string `gorm:"size:20" json:"level"`
	Subcounty       string `gorm:"size:100" json:"subcounty"`
	HSD             string `gorm:"size:100" json:"hsd"`
	District        string `gorm:"size:100" json:"district"`
	ClientCodePrefix string `gorm:"size:10;not null" json:"client_code_prefix"`
}
