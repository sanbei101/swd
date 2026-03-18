package repository

import (
	"context"

	"swd-new/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SensitiveWordRepository interface {
	List(ctx context.Context) ([]model.SensitiveWord, error)
	ListPage(ctx context.Context, offset, limit int) ([]model.SensitiveWord, int64, error)
	Create(ctx context.Context, word *model.SensitiveWord) error
	Update(ctx context.Context, word *model.SensitiveWord) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*model.SensitiveWord, error)
}

type sensitiveWordRepository struct {
	*Repository
}

func NewSensitiveWordRepository(repository *Repository) (SensitiveWordRepository, error) {
	return &sensitiveWordRepository{
		Repository: repository,
	}, nil
}

func (r *sensitiveWordRepository) List(ctx context.Context) ([]model.SensitiveWord, error) {
	words := make([]model.SensitiveWord, 0, 1024)
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&words).Error; err != nil {
		return nil, err
	}

	r.logger.Info("sensitive words loaded", zap.Int("count", len(words)))
	return words, nil
}

func (r *sensitiveWordRepository) ListPage(ctx context.Context, offset, limit int) ([]model.SensitiveWord, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.SensitiveWord{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	words := make([]model.SensitiveWord, 0, limit)
	if total == 0 {
		return words, 0, nil
	}

	if err := r.db.WithContext(ctx).
		Order("id ASC").
		Offset(offset).
		Limit(limit).
		Find(&words).Error; err != nil {
		return nil, 0, err
	}

	return words, total, nil
}

func (r *sensitiveWordRepository) Create(ctx context.Context, word *model.SensitiveWord) error {
	return r.db.WithContext(ctx).Create(word).Error
}

func (r *sensitiveWordRepository) Update(ctx context.Context, word *model.SensitiveWord) error {
	result := r.db.WithContext(ctx).Model(&model.SensitiveWord{}).Where("id = ?", word.ID).Updates(map[string]interface{}{
		"word": word.Word,
		"type": word.Type,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *sensitiveWordRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.SensitiveWord{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *sensitiveWordRepository) GetByID(ctx context.Context, id uint) (*model.SensitiveWord, error) {
	var word model.SensitiveWord
	if err := r.db.WithContext(ctx).First(&word, id).Error; err != nil {
		return nil, err
	}
	return &word, nil
}
