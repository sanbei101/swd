package model

type SensitiveWord struct {
	ID   uint   `gorm:"column:id;primaryKey;autoIncrement"`
	Word string `gorm:"column:word;not null"`
	Type string `gorm:"column:type;not null;default:'default'"`
}

func (SensitiveWord) TableName() string {
	return "sensitive_words"
}
