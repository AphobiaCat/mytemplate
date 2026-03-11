package model

import (
	"time"
)

// ExampleTable
type ExampleTable struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"` // key
	Name      string    `gorm:"size:100"`
	Email     string    `gorm:"size:100;uniqueIndex"`
	Password  string    `gorm:"size:100"`
	CreatedAt time.Time // auto add create time
	UpdatedAt time.Time // auto add update time
}

// TableName
func (ExampleTable) TableName() string {
	return "example_table"
}
