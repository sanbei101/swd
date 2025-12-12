package core

import (
	"github.com/kirklin/go-swd/pkg/types/category"
)

// AlgorithmType 算法类型
type AlgorithmType string

const (
	AlgorithmTrie        AlgorithmType = "trie"
	AlgorithmAhoCorasick AlgorithmType = "aho-corasick"
	AlgorithmDFA         AlgorithmType = "dfa"
)

// Algorithm 敏感词匹配算法接口
type Algorithm interface {
	Observer

	// Type 返回算法类型
	Type() AlgorithmType

	// Build 构建算法所需的数据结构
	Build(words map[string]category.Category) error

	// Detect 检查文本是否包含敏感词
	Detect(text string) bool

	// Match 返回文本中第一个敏感词
	Match(text string) *SensitiveWord

	// MatchAll 返回文本中所有敏感词
	MatchAll(text string) []SensitiveWord

	// Replace 替换敏感词
	Replace(text string, replacement rune) string
}

// MutableAlgorithm 可增量更新的算法接口
type MutableAlgorithm interface {
	// AddWord 增量添加词
	AddWord(word string, cat category.Category) error
	// RemoveWord 增量删除词
	RemoveWord(word string) error
}
