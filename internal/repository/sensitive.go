package repository

import (
	"context"

	"github.com/sanbei101/swd"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SensitiveWordRepository interface {
	List(ctx context.Context) ([]swd.SensitiveWord, error)
	ListPage(ctx context.Context, offset, limit int) ([]swd.SensitiveWord, int64, error)
	Create(ctx context.Context, word *swd.SensitiveWord) error
	Update(ctx context.Context, word *swd.SensitiveWord) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*swd.SensitiveWord, error)
}

type sensitiveWordRepository struct {
	*Repository
}

func NewSensitiveWordRepository(repository *Repository) (SensitiveWordRepository, error) {
	return &sensitiveWordRepository{
		Repository: repository,
	}, nil
}

func (r *sensitiveWordRepository) List(ctx context.Context) ([]swd.SensitiveWord, error) {
	words := make([]swd.SensitiveWord, 0, 1024)
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&words).Error; err != nil {
		return nil, err
	}

	r.logger.Info("sensitive words loaded", zap.Int("count", len(words)))
	return words, nil
}

func (r *sensitiveWordRepository) ListPage(ctx context.Context, offset, limit int) ([]swd.SensitiveWord, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&swd.SensitiveWord{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	words := make([]swd.SensitiveWord, 0, limit)
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

func (r *sensitiveWordRepository) Create(ctx context.Context, word *swd.SensitiveWord) error {
	return r.db.WithContext(ctx).Create(word).Error
}

func (r *sensitiveWordRepository) Update(ctx context.Context, word *swd.SensitiveWord) error {
	result := r.db.WithContext(ctx).Model(&swd.SensitiveWord{}).Where("id = ?", word.ID).Updates(map[string]interface{}{
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
	result := r.db.WithContext(ctx).Delete(&swd.SensitiveWord{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *sensitiveWordRepository) GetByID(ctx context.Context, id uint) (*swd.SensitiveWord, error) {
	var word swd.SensitiveWord
	if err := r.db.WithContext(ctx).First(&word, id).Error; err != nil {
		return nil, err
	}
	return &word, nil
}
