package model

import "github.com/kirklin/go-swd"

type Word struct {
	ID   uint         `gorm:"column:id;primaryKey;autoIncrement"`
	Word string       `gorm:"column:word;not null"`
	Type swd.Category `gorm:"column:type;not null;type:int"`
}

func (Word) TableName() string {
	return "sensitive_words"
}
