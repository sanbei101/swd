package service

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"

	"swd-new/internal/model"
	"swd-new/internal/repository"
	"swd-new/pkg/response"

	"gorm.io/gorm"
)

type SensitiveWordMatch struct {
	Word     string `json:"word"`
	Category string `json:"category"`
	StartPos int    `json:"start_pos"`
	EndPos   int    `json:"end_pos"`
}

type SensitiveWordCheckResult struct {
	Contains     bool                 `json:"contains"`
	FilteredText string               `json:"filtered_text"`
	Matches      []SensitiveWordMatch `json:"matches"`
}

type CreateSensitiveWordInput struct {
	Word string `json:"word"`
	Type string `json:"type"`
}

type UpdateSensitiveWordInput struct {
	Word string `json:"word"`
	Type string `json:"type"`
}

type SensitiveWordService interface {
	Check(text string) (*SensitiveWordCheckResult, error)
	ListWords(ctx context.Context, pageNum, pageSize int) (*response.Page[[]model.SensitiveWord], error)
	CreateWord(ctx context.Context, input CreateSensitiveWordInput) (*model.SensitiveWord, error)
	UpdateWord(ctx context.Context, id uint, input UpdateSensitiveWordInput) (*model.SensitiveWord, error)
	DeleteWord(ctx context.Context, id uint) error
}

var ErrInvalidSensitiveWordID = errors.New("invalid sensitive word id")

type sensitiveWordTrieNode struct {
	children map[rune]*sensitiveWordTrieNode
	word     string
	category string
	terminal bool
}

type sensitiveWordService struct {
	*Service
	repository repository.SensitiveWordRepository
	mu         sync.RWMutex
	root       *sensitiveWordTrieNode
	maxWordLen int
}

func NewSensitiveWordService(service *Service, sensitiveWordRepository repository.SensitiveWordRepository) (SensitiveWordService, error) {
	svc := &sensitiveWordService{
		Service:    service,
		repository: sensitiveWordRepository,
	}
	if err := svc.reloadTrie(context.Background()); err != nil {
		return nil, err
	}
	return svc, nil
}

func (s *sensitiveWordService) ListWords(ctx context.Context, pageNum, pageSize int) (*response.Page[[]model.SensitiveWord], error) {
	pageNum, pageSize, offset, limit := response.PageOffset(pageNum, pageSize)
	words, total, err := s.repository.ListPage(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	return response.ParsePage(words, pageNum, pageSize, total), nil
}

func (s *sensitiveWordService) CreateWord(ctx context.Context, input CreateSensitiveWordInput) (*model.SensitiveWord, error) {
	word, category, err := normalizeSensitiveWordInput(input.Word, input.Type)
	if err != nil {
		return nil, err
	}

	entity := &model.SensitiveWord{
		Word: word,
		Type: category,
	}
	if err := s.repository.Create(ctx, entity); err != nil {
		return nil, err
	}
	if err := s.reloadTrie(ctx); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *sensitiveWordService) UpdateWord(ctx context.Context, id uint, input UpdateSensitiveWordInput) (*model.SensitiveWord, error) {
	word, category, err := normalizeSensitiveWordInput(input.Word, input.Type)
	if err != nil {
		return nil, err
	}

	entity := &model.SensitiveWord{
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
	if err := s.reloadTrie(ctx); err != nil {
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
	return s.reloadTrie(ctx)
}

func (s *sensitiveWordService) Check(text string) (*SensitiveWordCheckResult, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("text must not be empty")
	}

	s.mu.RLock()
	root := s.root
	maxWordLen := s.maxWordLen
	s.mu.RUnlock()

	textRunes := []rune(text)
	matches := matchAll(root, maxWordLen, textRunes)
	if len(matches) == 0 {
		return &SensitiveWordCheckResult{
			Contains:     false,
			FilteredText: text,
			Matches:      []SensitiveWordMatch{},
		}, nil
	}
	return &SensitiveWordCheckResult{
		Contains:     len(matches) > 0,
		FilteredText: replaceWithAsterisk(textRunes, matches),
		Matches:      matches,
	}, nil
}

func (s *sensitiveWordService) reloadTrie(ctx context.Context) error {
	words, err := s.repository.List(ctx)
	if err != nil {
		return err
	}
	if len(words) == 0 {
		return errors.New("no sensitive words loaded from postgres")
	}

	root, maxWordLen := buildSensitiveWordTrie(words)

	s.mu.Lock()
	s.root = root
	s.maxWordLen = maxWordLen
	s.mu.Unlock()
	return nil
}

func buildSensitiveWordTrie(words []model.SensitiveWord) (*sensitiveWordTrieNode, int) {
	root := &sensitiveWordTrieNode{}
	maxWordLen := 0

	for _, word := range words {
		wordRunes := []rune(word.Word)
		if len(wordRunes) == 0 {
			continue
		}
		if len(wordRunes) > maxWordLen {
			maxWordLen = len(wordRunes)
		}

		node := root
		for _, r := range wordRunes {
			if node.children == nil {
				node.children = make(map[rune]*sensitiveWordTrieNode)
			}
			child := node.children[r]
			if child == nil {
				child = &sensitiveWordTrieNode{}
				node.children[r] = child
			}
			node = child
		}

		node.terminal = true
		node.word = word.Word
		node.category = word.Type
	}

	return root, maxWordLen
}

func matchAll(root *sensitiveWordTrieNode, maxWordLen int, textRunes []rune) []SensitiveWordMatch {
	if len(textRunes) == 0 || root == nil || maxWordLen == 0 {
		return []SensitiveWordMatch{}
	}

	candidates := make([]SensitiveWordMatch, 0, 8)
	for start := range textRunes {
		node := root
		limit := len(textRunes)
		if maxEnd := start + maxWordLen; maxEnd < limit {
			limit = maxEnd
		}

		for end := start; end < limit; end++ {
			node = node.children[textRunes[end]]
			if node == nil {
				break
			}
			if node.terminal {
				candidates = append(candidates, SensitiveWordMatch{
					Word:     node.word,
					Category: node.category,
					StartPos: start,
					EndPos:   end + 1,
				})
			}
		}
	}

	if len(candidates) == 0 {
		return []SensitiveWordMatch{}
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		leftLen := candidates[i].EndPos - candidates[i].StartPos
		rightLen := candidates[j].EndPos - candidates[j].StartPos
		if leftLen != rightLen {
			return leftLen > rightLen
		}
		if candidates[i].StartPos != candidates[j].StartPos {
			return candidates[i].StartPos < candidates[j].StartPos
		}
		return candidates[i].EndPos < candidates[j].EndPos
	})

	matches := make([]SensitiveWordMatch, 0, len(candidates))
	occupied := make([]bool, len(textRunes))
	for _, candidate := range candidates {
		overlap := false
		for i := candidate.StartPos; i < candidate.EndPos; i++ {
			if occupied[i] {
				overlap = true
				break
			}
		}
		if overlap {
			continue
		}

		for i := candidate.StartPos; i < candidate.EndPos; i++ {
			occupied[i] = true
		}
		matches = append(matches, candidate)
	}

	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].StartPos != matches[j].StartPos {
			return matches[i].StartPos < matches[j].StartPos
		}
		return matches[i].EndPos < matches[j].EndPos
	})

	return matches
}

func replaceWithAsterisk(textRunes []rune, matches []SensitiveWordMatch) string {
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
		category = "default"
	}

	return word, category, nil
}
