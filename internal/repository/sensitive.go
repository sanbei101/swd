package repository

import (
	"context"
	"swd-new/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SensitiveWordRepository interface {
	List(ctx context.Context) ([]model.Word, error)
	ListPage(ctx context.Context, offset, limit int) ([]model.Word, int64, error)
	Create(ctx context.Context, word *model.Word) error
	Update(ctx context.Context, word *model.Word) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*model.Word, error)
}

type sensitiveWordRepository struct {
	*Repository
}

func NewSensitiveWordRepository(repository *Repository) (SensitiveWordRepository, error) {
	return &sensitiveWordRepository{
		Repository: repository,
	}, nil
}

func (r *sensitiveWordRepository) List(ctx context.Context) ([]model.Word, error) {
	words := make([]model.Word, 0, 1024)
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&words).Error; err != nil {
		return nil, err
	}

	r.logger.Info("sensitive words loaded", zap.Int("count", len(words)))
	return words, nil
}

func (r *sensitiveWordRepository) ListPage(ctx context.Context, offset, limit int) ([]model.Word, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Word{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	words := make([]model.Word, 0, limit)
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

func (r *sensitiveWordRepository) Create(ctx context.Context, word *model.Word) error {
	return r.db.WithContext(ctx).Create(word).Error
}

func (r *sensitiveWordRepository) Update(ctx context.Context, word *model.Word) error {
	result := r.db.WithContext(ctx).Model(&model.Word{}).Where("id = ?", word.ID).Updates(map[string]interface{}{
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
	result := r.db.WithContext(ctx).Delete(&model.Word{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *sensitiveWordRepository) GetByID(ctx context.Context, id uint) (*model.Word, error) {
	var word model.Word
	if err := r.db.WithContext(ctx).First(&word, id).Error; err != nil {
		return nil, err
	}
	return &word, nil
}
