package service

import (
	"context"
	"errors"
	"strings"

	"github.com/sanbei101/swd"

	"swd-new/internal/repository"
	"swd-new/pkg/response"

	"gorm.io/gorm"
)

const (
	DefaultSensitiveWordType = "default"
)

type CreateSensitiveWordInput struct {
	Word string `json:"word"`
	Type string `json:"type"`
}

type UpdateSensitiveWordInput struct {
	Word string `json:"word"`
	Type string `json:"type"`
}

type SensitiveWordService interface {
	Check(text string) (*swd.SensitiveWordCheckResult, error)
	ListWords(ctx context.Context, pageNum, pageSize int) (*response.Page[[]swd.SensitiveWord], error)
	CreateWord(ctx context.Context, input CreateSensitiveWordInput) (*swd.SensitiveWord, error)
	UpdateWord(ctx context.Context, id uint, input UpdateSensitiveWordInput) (*swd.SensitiveWord, error)
	DeleteWord(ctx context.Context, id uint) error
}

var ErrInvalidSensitiveWordID = errors.New("invalid sensitive word id")

type sensitiveWordService struct {
	*Service
	repository repository.SensitiveWordRepository
	swd        *swd.Swd
}

func NewSensitiveWordService(service *Service, sensitiveWordRepository repository.SensitiveWordRepository) (SensitiveWordService, error) {
	words, err := sensitiveWordRepository.List(context.Background())
	if err != nil {
		return nil, err
	}
	if len(words) == 0 {
		return nil, errors.New("no sensitive words loaded from postgres")
	}
	swd := swd.NewSwd(words)
	svc := &sensitiveWordService{
		Service:    service,
		repository: sensitiveWordRepository,
		swd:        swd,
	}
	return svc, nil
}

func (s *sensitiveWordService) ListWords(ctx context.Context, pageNum, pageSize int) (*response.Page[[]swd.SensitiveWord], error) {
	pageNum, pageSize, offset, limit := response.PageOffset(pageNum, pageSize)
	words, total, err := s.repository.ListPage(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	return response.ParsePage(words, pageNum, pageSize, total), nil
}

func (s *sensitiveWordService) CreateWord(ctx context.Context, input CreateSensitiveWordInput) (*swd.SensitiveWord, error) {
	word, category, err := normalizeSensitiveWordInput(input.Word, input.Type)
	if err != nil {
		return nil, err
	}

	entity := &swd.SensitiveWord{
		Word: word,
		Type: category,
	}
	if err := s.repository.Create(ctx, entity); err != nil {
		return nil, err
	}
	if err := s.reloadSwd(ctx); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *sensitiveWordService) UpdateWord(ctx context.Context, id uint, input UpdateSensitiveWordInput) (*swd.SensitiveWord, error) {
	word, category, err := normalizeSensitiveWordInput(input.Word, input.Type)
	if err != nil {
		return nil, err
	}

	entity := &swd.SensitiveWord{
		ID:   id,
		Word: word,
		Type: category,
	}
	if err := s.repository.Update(ctx, entity); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("sensitive word not found")
		}
		return nil, err
	}
	if err := s.reloadSwd(ctx); err != nil {
		return nil, err
	}
	return s.repository.GetByID(ctx, id)
}

func (s *sensitiveWordService) DeleteWord(ctx context.Context, id uint) error {
	if err := s.repository.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("sensitive word not found")
		}
		return err
	}
	return s.reloadSwd(ctx)
}

func (s *sensitiveWordService) Check(text string) (*swd.SensitiveWordCheckResult, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("text must not be empty")
	}

	return s.swd.Check(text)
}

func (s *sensitiveWordService) reloadSwd(ctx context.Context) error {
	words, err := s.repository.List(ctx)
	if err != nil {
		return err
	}
	s.swd = swd.NewSwd(words)
	return nil
}

func replaceWithAsterisk(textRunes []rune, matches []swd.SensitiveWordMatch) string {
	for _, match := range matches {
		for i := match.StartPos; i < match.EndPos; i++ {
			textRunes[i] = '*'
		}
	}
	return string(textRunes)
}

func normalizeSensitiveWordInput(word, category string) (string, string, error) {
	word = strings.TrimSpace(word)
	if word == "" {
		return "", "", errors.New("word must not be empty")
	}

	category = strings.TrimSpace(category)
	if category == "" {
		category = DefaultSensitiveWordType
	}

	return word, category, nil
}
