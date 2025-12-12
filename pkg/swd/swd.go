package swd

import (
	"context"

	"github.com/kirklin/go-swd/pkg/types/category"

	"github.com/kirklin/go-swd/pkg/core"
	"github.com/kirklin/go-swd/pkg/dictionary"
)

// ComponentFactory 定义了创建各种组件的工厂接口
type ComponentFactory interface {
	CreateDetector(options *core.SWDOptions, loader core.Loader) (core.Detector, error)
	CreateFilter(detector core.Detector) core.Filter
	CreateLoader() core.Loader
	CreateComponents(options *core.SWDOptions) (core.Detector, core.Filter, core.Loader, error)
}

// SWD 敏感词检测与过滤引擎的实现
type SWD struct {
	detector      core.Detector
	filter        core.Filter
	loader        core.Loader
	options       *core.SWDOptions
	detectCalls   uint64
	matchAllCalls uint64
	replaceCalls  uint64
}

// New 创建一个敏感词检测引擎
func New(factory ComponentFactory) (*SWD, error) {
	if factory == nil {
		return nil, ErrNoFactory
	}

	options := &core.SWDOptions{}

	// 使用工厂的CreateComponents方法创建并关联组件
	detector, filter, loader, err := factory.CreateComponents(options)
	if err != nil {
		return nil, err
	}

	if detector == nil {
		return nil, ErrNoDetector
	}
	if filter == nil {
		return nil, ErrNoFilter
	}
	if loader == nil {
		return nil, ErrNoLoader
	}

	return &SWD{
		detector: detector,
		filter:   filter,
		loader:   loader,
		options:  options,
	}, nil
}

// LoadDefaultWords 加载默认词库
func (swd *SWD) LoadDefaultWords(ctx context.Context) error {
	return swd.loader.LoadDefaultWords(ctx)
}

// LoadCustomWords 加载自定义词库
func (swd *SWD) LoadCustomWords(ctx context.Context, words map[string]category.Category) error {
	return swd.loader.LoadCustomWords(ctx, words)
}

// AddWord 添加单个敏感词
func (swd *SWD) AddWord(word string, category category.Category) error {
	return swd.loader.AddWord(word, category)
}

// AddWords 批量添加敏感词
func (swd *SWD) AddWords(words map[string]category.Category) error {
	return swd.loader.AddWords(words)
}

// RemoveWord 移除敏感词
func (swd *SWD) RemoveWord(word string) error {
	return swd.loader.RemoveWord(word)
}

// RemoveWords 批量移除敏感词
func (swd *SWD) RemoveWords(words []string) error {
	return swd.loader.RemoveWords(words)
}

// Clear 清空所有敏感词
func (swd *SWD) Clear() error {
	return swd.loader.Clear()
}

// Detect 检查文本是否包含敏感词
func (swd *SWD) Detect(text string) bool {
	swd.detectCalls++
	return swd.detector.Detect(text)
}

// DetectIn 检查文本是否包含指定分类的敏感词
func (swd *SWD) DetectIn(text string, categories ...category.Category) bool {
	return swd.detector.DetectIn(text, categories...)
}

// Match 返回文本中找到的第一个敏感词
func (swd *SWD) Match(text string) *core.SensitiveWord {
	if text == "" {
		return nil
	}
	return swd.detector.Match(text)
}

// MatchIn 返回文本中第一个指定分类的敏感词
func (swd *SWD) MatchIn(text string, categories ...category.Category) *core.SensitiveWord {
	return swd.detector.MatchIn(text, categories...)
}

// MatchAll 返回文本中所有敏感词
func (swd *SWD) MatchAll(text string) []core.SensitiveWord {
	swd.matchAllCalls++
	return swd.detector.MatchAll(text)
}

// MatchAllIn 返回文本中所有指定分类的敏感词
func (swd *SWD) MatchAllIn(text string, categories ...category.Category) []core.SensitiveWord {
	return swd.detector.MatchAllIn(text, categories...)
}

// Replace 使用指定的替换字符替换敏感词
func (swd *SWD) Replace(text string, replacement rune) string {
	swd.replaceCalls++
	return swd.filter.Replace(text, replacement)
}

// ReplaceIn 使用指定的替换字符替换指定分类的敏感词
func (swd *SWD) ReplaceIn(text string, replacement rune, categories ...category.Category) string {
	swd.replaceCalls++
	return swd.filter.ReplaceIn(text, replacement, categories...)
}

// ReplaceWithAsterisk 使用星号替换敏感词
func (swd *SWD) ReplaceWithAsterisk(text string) string {
	swd.replaceCalls++
	return swd.filter.ReplaceWithAsterisk(text)
}

// ReplaceWithAsteriskIn 使用星号替换指定分类的敏感词
func (swd *SWD) ReplaceWithAsteriskIn(text string, categories ...category.Category) string {
	swd.replaceCalls++
	return swd.filter.ReplaceWithAsteriskIn(text, categories...)
}

// ReplaceWithStrategy 使用自定义策略替换敏感词
func (swd *SWD) ReplaceWithStrategy(text string, strategy func(word core.SensitiveWord) string) string {
	swd.replaceCalls++
	return swd.filter.ReplaceWithStrategy(text, strategy)
}

// ReplaceWithStrategyIn 使用自定义策略替换指定分类的敏感词
func (swd *SWD) ReplaceWithStrategyIn(text string, strategy func(word core.SensitiveWord) string, categories ...category.Category) string {
	swd.replaceCalls++
	return swd.filter.ReplaceWithStrategyIn(text, strategy, categories...)
}

func (swd *SWD) Metrics() (detectCalls, matchAllCalls, replaceCalls uint64) {
	return swd.detectCalls, swd.matchAllCalls, swd.replaceCalls
}

func (swd *SWD) BuildMetrics() (buildCount uint64, lastBuildMs int64, avgBuildMs float64) {
	type withStats interface {
		GetBuildStats() (uint64, int64, int64)
	}
	if d, ok := swd.detector.(withStats); ok {
		count, last, total := d.GetBuildStats()
		avg := 0.0
		if count > 0 {
			avg = float64(total) / float64(count)
		}
		return count, last, avg
	}
	return 0, 0, 0
}

func (swd *SWD) DetectBatch(texts []string) []bool {
	res := make([]bool, len(texts))
	for i, t := range texts {
		res[i] = swd.Detect(t)
	}
	return res
}

func (swd *SWD) MatchAllBatch(texts []string) [][]core.SensitiveWord {
	res := make([][]core.SensitiveWord, len(texts))
	for i, t := range texts {
		res[i] = swd.MatchAll(t)
	}
	return res
}

func (swd *SWD) ReplaceBatch(texts []string, replacement rune) []string {
	res := make([]string, len(texts))
	for i, t := range texts {
		res[i] = swd.Replace(t, replacement)
	}
	return res
}

func (swd *SWD) DetectWithContext(ctx context.Context, text string) bool {
	select {
	case <-ctx.Done():
		return false
	default:
		return swd.Detect(text)
	}
}

// Close 释放资源并解除观察者
func (swd *SWD) Close() {
	if swd == nil {
		return
	}
	if observer, ok := swd.detector.(core.Observer); ok {
		if dl, ok := swd.loader.(*dictionary.Loader); ok {
			dl.RemoveObserver(observer)
		}
	}
}
