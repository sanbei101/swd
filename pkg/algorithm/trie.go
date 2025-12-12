package algorithm

import (
	"log"

	"github.com/kirklin/go-swd/pkg/core"
	"github.com/kirklin/go-swd/pkg/types/category"
)

// TrieNode Trie树节点
type TrieNode struct {
	children map[rune]*TrieNode // 子节点映射
	isEnd    bool               // 是否是单词结尾
	word     string             // 如果是结尾节点，存储完整词
	category category.Category  // 敏感词分类
}

// newTrieNode 创建新的Trie节点
func newTrieNode() *TrieNode {
	return &TrieNode{
		children: make(map[rune]*TrieNode),
	}
}

// Trie 字典树实现
type Trie struct {
	root *TrieNode
}

// NewTrie 创建新的Trie树
func NewTrie() *Trie {
	return &Trie{
		root: newTrieNode(),
	}
}

// Type 返回算法类型
func (t *Trie) Type() core.AlgorithmType {
	return core.AlgorithmTrie
}

// Build 构建敏感词字典树
func (t *Trie) Build(words map[string]category.Category) error {
	t.root = newTrieNode()
	for word, category := range words {
		current := t.root
		for _, char := range word {
			if _, exists := current.children[char]; !exists {
				current.children[char] = newTrieNode()
			}
			current = current.children[char]
		}
		current.isEnd = true
		current.word = word
		current.category = category
	}
	return nil
}

// Match 返回文本中第一个敏感词
func (t *Trie) Match(text string) *core.SensitiveWord {
	runes := []rune(text)
	for i := range runes {
		if match := t.matchFromPosition(text, i); match != nil {
			return match
		}
	}
	return nil
}

// matchFromPosition 从指定位置开始匹配
func (t *Trie) matchFromPosition(text string, start int) *core.SensitiveWord {
	current := t.root
	runes := []rune(text)
	if start >= len(runes) {
		return nil
	}

	for i, char := range runes[start:] {
		next, exists := current.children[char]
		if !exists {
			break
		}
		current = next
		if current.isEnd {
			return &core.SensitiveWord{
				Word:     current.word,
				StartPos: start,
				EndPos:   start + i + 1,
				Category: current.category,
			}
		}
	}
	return nil
}

// MatchAll 返回文本中所有敏感词
func (t *Trie) MatchAll(text string) []core.SensitiveWord {
	var matches []core.SensitiveWord
	runes := []rune(text)
	for i := range runes {
		if match := t.matchFromPosition(text, i); match != nil {
			matches = append(matches, *match)
		}
	}
	return matches
}

// Replace 替换敏感词
func (t *Trie) Replace(text string, replacement rune) string {
	matches := t.MatchAll(text)
	if len(matches) == 0 {
		return text
	}

	runes := []rune(text)
	for _, match := range matches {
		for i := match.StartPos; i < match.EndPos; i++ {
			runes[i] = replacement
		}
	}
	return string(runes)
}

// Detect 检查文本是否包含敏感词
func (t *Trie) Detect(text string) bool {
	return t.Match(text) != nil
}

// OnWordsChanged 实现 Observer 接口,当词库变更时重建算法
func (t *Trie) OnWordsChanged(words map[string]category.Category) {
	if err := t.Build(words); err != nil {
		// 这里只能记录错误,因为是回调方法
		log.Printf("重建算法失败: %v", err)
	}
}

// AddWord 增量添加词
func (t *Trie) AddWord(word string, cat category.Category) error {
	if word == "" {
		return nil
	}
	current := t.root
	for _, r := range []rune(word) {
		next := current.children[r]
		if next == nil {
			next = newTrieNode()
			current.children[r] = next
		}
		current = next
	}
	current.isEnd = true
	current.word = word
	current.category = cat
	return nil
}

// RemoveWord 增量删除词
func (t *Trie) RemoveWord(word string) error {
	if word == "" {
		return nil
	}
	// 路径栈用于回溯修剪
	stack := make([]*TrieNode, 0, len([]rune(word))+1)
	current := t.root
	stack = append(stack, current)
	for _, r := range []rune(word) {
		next := current.children[r]
		if next == nil {
			return nil
		}
		current = next
		stack = append(stack, current)
	}
	if !current.isEnd {
		return nil
	}
	current.isEnd = false
	current.word = ""
	// 自底向上修剪无用叶子
	runes := []rune(word)
	for i := len(runes) - 1; i >= 0; i-- {
		node := stack[i+1]
		parent := stack[i]
		if node.isEnd || len(node.children) > 0 {
			break
		}
		delete(parent.children, runes[i])
	}
	return nil
}
