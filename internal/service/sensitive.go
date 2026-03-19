package service

import (
	"context"
	"errors"
	"strings"

	"github.com/kirklin/go-swd"

	"swd-new/internal/model"
	"swd-new/internal/repository"
	"swd-new/pkg/response"

	"gorm.io/gorm"
)

const (
	DefaultSensitiveWordType = 0
)

type CreateSensitiveWordInput struct {
	Word string       `json:"word"`
	Type swd.Category `json:"type"`
}

type UpdateSensitiveWordInput struct {
	Word string       `json:"word"`
	Type swd.Category `json:"type"`
}

type SensitiveWordCheckResult struct {
	Contain      bool                `json:"contain"`
	FilteredText string              `json:"filtered_text"`
	Matches      []swd.SensitiveWord `json:"matches"`
}

type SensitiveWordService interface {
	Check(text string) (*SensitiveWordCheckResult, error)
	ListWords(ctx context.Context, pageNum, pageSize int) (*response.Page[[]model.Word], error)
	CreateWord(ctx context.Context, input CreateSensitiveWordInput) (*model.Word, error)
	UpdateWord(ctx context.Context, id uint, input UpdateSensitiveWordInput) (*model.Word, error)
	DeleteWord(ctx context.Context, id uint) error
}

var ErrInvalidSensitiveWordID = errors.New("invalid sensitive word id")

type sensitiveWordService struct {
	*Service
	repository repository.SensitiveWordRepository
	swd        *swd.SWD
}

func NewSensitiveWordService(service *Service, sensitiveWordRepository repository.SensitiveWordRepository) (SensitiveWordService, error) {
	detector, err := swd.New()
	if err != nil {
		return nil, err
	}
	svc := &sensitiveWordService{
		Service:    service,
		repository: sensitiveWordRepository,
		swd:        detector,
	}
	return svc, nil
}

func (s *sensitiveWordService) ListWords(ctx context.Context, pageNum, pageSize int) (*response.Page[[]model.Word], error) {
	pageNum, pageSize, offset, limit := response.PageOffset(pageNum, pageSize)
	words, total, err := s.repository.ListPage(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	return response.ParsePage(words, pageNum, pageSize, total), nil
}

func (s *sensitiveWordService) CreateWord(ctx context.Context, input CreateSensitiveWordInput) (*model.Word, error) {
	word, category, err := normalizeSensitiveWordInput(input.Word, input.Type)
	if err != nil {
		return nil, err
	}

	entity := &model.Word{
		Word: word,
		Type: category,
	}
	if err := s.repository.Create(ctx, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *sensitiveWordService) UpdateWord(ctx context.Context, id uint, input UpdateSensitiveWordInput) (*model.Word, error) {
	word, category, err := normalizeSensitiveWordInput(input.Word, input.Type)
	if err != nil {
		return nil, err
	}

	entity := &model.Word{
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
	return s.repository.GetByID(ctx, id)
}

func (s *sensitiveWordService) DeleteWord(ctx context.Context, id uint) error {
	if err := s.repository.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("sensitive word not found")
		}
		return err
	}
	return nil
}

func (s *sensitiveWordService) Check(text string) (*SensitiveWordCheckResult, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("text must not be empty")
	}

	contain := s.swd.Detect(text)
	if !contain {
		return &SensitiveWordCheckResult{
			Contain:      false,
			FilteredText: text,
			Matches:      nil,
		}, nil
	}
	words := s.swd.MatchAll(text)
	filtered := s.swd.ReplaceWithAsterisk(text)

	return &SensitiveWordCheckResult{
		Contain:      true,
		FilteredText: filtered,
		Matches:      words,
	}, nil
}

func normalizeSensitiveWordInput(word string, category swd.Category) (string, swd.Category, error) {
	word = strings.TrimSpace(word)
	if word == "" {
		return "", 0, errors.New("word must not be empty")
	}

	if category == 0 {
		category = DefaultSensitiveWordType
	}

	return word, category, nil
}
