package detector

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/kirklin/go-swd/pkg/algorithm"
	"github.com/kirklin/go-swd/pkg/core"
	"github.com/kirklin/go-swd/pkg/detector/preprocessor"
	"github.com/kirklin/go-swd/pkg/types/category"
)

// detector 实现敏感词检测器接口
type detector struct {
	activeAlgo   core.Algorithm
	preprocess   *preprocessor.Preprocessor
	mu           sync.RWMutex
	options      core.SWDOptions
	loader       core.Loader
	logger       core.Logger
	buildCount   uint64
	lastBuildMs  int64
	totalBuildMs int64
	algoType     core.AlgorithmType
	buildCh      chan struct{}
	buildMu      sync.Mutex
	buildWords   map[string]category.Category
}

func (d *detector) GetBuildStats() (count uint64, lastMs int64, totalMs int64) {
	return d.buildCount, d.lastBuildMs, d.totalBuildMs
}

// NewDetector 创建一个新的检测器实例（依赖注入）
func NewDetector(options core.SWDOptions, loader core.Loader, algo core.Algorithm) (core.Detector, error) {
	if loader == nil || algo == nil {
		return nil, fmt.Errorf("依赖未提供")
	}

	// 加载词典（确保外部已加载，否则此处补全）
	if len(loader.GetWords()) == 0 {
		if err := loader.LoadDefaultWords(context.Background()); err != nil {
			return nil, fmt.Errorf("加载默认词典失败: %w", err)
		}
	}

	words := loader.GetWords()
	if len(words) == 0 {
		return nil, fmt.Errorf("词典内容为空")
	}

	if err := algo.Build(words); err != nil {
		return nil, fmt.Errorf("构建算法失败: %w", err)
	}

	d := &detector{
		activeAlgo: algo,
		preprocess: preprocessor.NewPreprocessor(options),
		options:    options,
		loader:     loader,
		logger:     options.Logger,
		algoType:   algo.Type(),
		buildCh:    make(chan struct{}, 1),
	}

	// 启动后台构建协程
	go d.buildWorker()
	return d, nil
}

// OnWordsChanged 实现Observer接口,当词库变更时重建算法
func (d *detector) OnWordsChanged(words map[string]category.Category) {
	// 保存最新快照并触发构建信号（非阻塞，合并突发）
	d.buildMu.Lock()
	// 拷贝快照，避免外部修改影响
	snapshot := make(map[string]category.Category, len(words))
	for k, v := range words {
		snapshot[k] = v
	}
	d.buildWords = snapshot
	d.buildMu.Unlock()
	select {
	case d.buildCh <- struct{}{}:
	default:
	}
}

func (d *detector) buildWorker() {
	for range d.buildCh {
		// 读取最新快照
		d.buildMu.Lock()
		words := d.buildWords
		d.buildMu.Unlock()
		if words == nil {
			continue
		}

		start := time.Now()
		// 新建 Composite 自动机进行离线构建
		builder := algorithm.NewComposite(d.algoType)
		if err := builder.Build(words); err != nil {
			if d.logger != nil {
				d.logger.Error(fmt.Sprintf("后台构建失败: %v", err))
			}
			continue
		}
		// 原子切换读用自动机
		d.mu.Lock()
		d.activeAlgo = builder
		d.mu.Unlock()

		d.buildCount++
		d.lastBuildMs = time.Since(start).Milliseconds()
		d.totalBuildMs += d.lastBuildMs
		if d.logger != nil {
			d.logger.Info(fmt.Sprintf("后台构建完成: %d ms (累计: %d)", d.lastBuildMs, d.buildCount))
		}
	}
}

// SetOptions 更新检测器选项
func (d *detector) SetOptions(options core.SWDOptions) {
	d.mu.Lock()
	d.options = options
	d.preprocess = preprocessor.NewPreprocessor(options)
	d.mu.Unlock()
}

func (d *detector) gapMatchAll(processedText string, mapping []int) []core.SensitiveWord {
	if d.loader == nil || d.options.MaxDistance <= 0 {
		return nil
	}
	textRunes := []rune(processedText)
	words := d.loader.GetWords()
	var out []core.SensitiveWord
	for w, cat := range words {
		pw := d.preprocess.Process(w)
		wr := []rune(pw)
		if len(wr) == 0 || len(wr) > 8 {
			continue
		}
		for i := 0; i < len(textRunes); i++ {
			ti := i
			wi := 0
			gaps := 0
			start := -1
			for ti < len(textRunes) && wi < len(wr) {
				r := textRunes[ti]
				if r == wr[wi] {
					if start == -1 {
						start = ti
					}
					wi++
					ti++
					gaps = 0
					continue
				}
				if r == '*' || r == '·' || r == '-' || r == '_' || r == '.' || r == ' ' {
					if gaps < d.options.MaxDistance {
						gaps++
						ti++
						continue
					}
				}
				break
			}
			if wi == len(wr) && start >= 0 {
				s := start
				e := ti
				if mapping != nil && s >= 0 && e-1 < len(mapping) {
					s = mapping[s]
					e = mapping[e-1] + 1
				}
				out = append(out, core.SensitiveWord{Word: w, StartPos: s, EndPos: e, Category: cat})
			}
		}
	}
	return out
}

func (d *detector) regexMatches(text string) []core.SensitiveWord {
	var res []core.SensitiveWord
	if d.options.EnableURLCheck {
		urlRe := `(?i)\b((https?://)?([a-z0-9-]+\.)+[a-z]{2,}(/[\w\-./?%&=]*)?)\b`
		res = append(res, d.findByRegex(text, urlRe)...)
	}
	if d.options.EnableEmailCheck {
		emailRe := `(?i)\b[\w.%+\-]+@[a-z0-9.-]+\.[a-z]{2,}\b`
		res = append(res, d.findByRegex(text, emailRe)...)
	}
	return res
}

func (d *detector) findByRegex(text, pattern string) []core.SensitiveWord {
	re := regexp.MustCompile(pattern)
	var res []core.SensitiveWord
	locs := re.FindAllStringIndex(text, -1)
	for _, loc := range locs {
		w := text[loc[0]:loc[1]]
		res = append(res, core.SensitiveWord{Word: w, StartPos: loc[0], EndPos: loc[1], Category: category.Custom})
	}
	return res
}

// Detect 检查文本是否包含任何敏感词
func (d *detector) Detect(text string) bool {
	if text == "" {
		return false
	}

	// 预处理文本（含索引映射）
	processedText, mapping := d.preprocess.ProcessWithMap(text)

	// 使用读锁进行检测
	d.mu.RLock()
	match := d.activeAlgo.Match(processedText)
	d.mu.RUnlock()

	if match != nil {
		return true
	}
	if d.options.MaxDistance > 0 {
		ms := d.gapMatchAll(processedText, mapping)
		if len(ms) > 0 {
			return true
		}
	}
	if d.options.EnableURLCheck || d.options.EnableEmailCheck {
		if len(d.regexMatches(text)) > 0 {
			return true
		}
	}
	return false
}

// DetectIn 检查文本是否包含指定分类的敏感词
func (d *detector) DetectIn(text string, categories ...category.Category) bool {
	if text == "" || len(categories) == 0 {
		return false
	}

	// 预处理文本（含索引映射）
	processedText, mapping := d.preprocess.ProcessWithMap(text)

	d.mu.RLock()
	matches := d.activeAlgo.MatchAll(processedText)
	d.mu.RUnlock()
	for _, match := range matches {
		for _, cat := range categories {
			if cat.Contains(match.Category) {
				return true
			}
		}
	}
	if d.options.MaxDistance > 0 {
		ms := d.gapMatchAll(processedText, mapping)
		for _, match := range ms {
			for _, cat := range categories {
				if cat.Contains(match.Category) {
					return true
				}
			}
		}
	}
	if d.options.EnableURLCheck || d.options.EnableEmailCheck {
		extra := d.regexMatches(text)
		for _, match := range extra {
			for _, cat := range categories {
				if cat.Contains(match.Category) {
					return true
				}
			}
		}
	}
	return false
}

// Match 返回文本中找到的第一个敏感词
func (d *detector) Match(text string) *core.SensitiveWord {
	if text == "" {
		return nil
	}

	// 预处理文本（含索引映射）
	processedText, mapping := d.preprocess.ProcessWithMap(text)

	// 使用读锁进行检测
	d.mu.RLock()
	match := d.activeAlgo.Match(processedText)
	d.mu.RUnlock()

	if match != nil {
		if mapping != nil {
			if match.StartPos >= 0 && match.EndPos-1 < len(mapping) {
				match.StartPos = mapping[match.StartPos]
				match.EndPos = mapping[match.EndPos-1] + 1
			}
		}
		return match
	}
	if d.options.MaxDistance > 0 {
		ms := d.gapMatchAll(processedText, mapping)
		if len(ms) > 0 {
			m := ms[0]
			return &m
		}
	}
	if d.options.EnableURLCheck || d.options.EnableEmailCheck {
		extra := d.regexMatches(text)
		if len(extra) > 0 {
			m := extra[0]
			return &m
		}
	}
	return nil
}

// MatchIn 返回文本中找到的第一个指定分类的敏感词
func (d *detector) MatchIn(text string, categories ...category.Category) *core.SensitiveWord {
	if text == "" || len(categories) == 0 {
		return nil
	}

	// 预处理文本（含索引映射）
	processedText, mapping := d.preprocess.ProcessWithMap(text)

	// 使用读锁进行检测
	d.mu.RLock()
	matches := d.activeAlgo.MatchAll(processedText)
	d.mu.RUnlock()

	// 返回第一个匹配的分类
	for _, match := range matches {
		for _, cat := range categories {
			if cat.Contains(match.Category) {
				result := match
				if mapping != nil {
					if result.StartPos >= 0 && result.EndPos-1 < len(mapping) {
						result.StartPos = mapping[result.StartPos]
						result.EndPos = mapping[result.EndPos-1] + 1
					}
				}
				return &result
			}
		}
	}

	return nil
}

// MatchAll 返回文本中找到的所有敏感词
func (d *detector) MatchAll(text string) []core.SensitiveWord {
	if text == "" {
		return nil
	}

	// 预处理文本（含索引映射）
	processedText, mapping := d.preprocess.ProcessWithMap(text)

	d.mu.RLock()
	matches := d.activeAlgo.MatchAll(processedText)
	d.mu.RUnlock()
	if mapping != nil {
		for i := range matches {
			if matches[i].StartPos >= 0 && matches[i].EndPos-1 < len(mapping) {
				matches[i].StartPos = mapping[matches[i].StartPos]
				matches[i].EndPos = mapping[matches[i].EndPos-1] + 1
			}
		}
	}
	if d.options.MaxDistance > 0 {
		ms := d.gapMatchAll(processedText, mapping)
		matches = append(matches, ms...)
	}
	if d.options.EnableURLCheck || d.options.EnableEmailCheck {
		matches = append(matches, d.regexMatches(text)...)
	}
	return matches
}

// MatchAllIn 返回文本中找到的所有指定分类的敏感词
func (d *detector) MatchAllIn(text string, categories ...category.Category) []core.SensitiveWord {
	if text == "" || len(categories) == 0 {
		return nil
	}

	// 预处理文本（含索引映射）
	processedText, mapping := d.preprocess.ProcessWithMap(text)

	d.mu.RLock()
	allMatches := d.activeAlgo.MatchAll(processedText)
	d.mu.RUnlock()
	if mapping != nil {
		for i := range allMatches {
			if allMatches[i].StartPos >= 0 && allMatches[i].EndPos-1 < len(mapping) {
				allMatches[i].StartPos = mapping[allMatches[i].StartPos]
				allMatches[i].EndPos = mapping[allMatches[i].EndPos-1] + 1
			}
		}
	}

	// 过滤出指定分类的敏感词
	var matches []core.SensitiveWord
	for _, match := range allMatches {
		for _, cat := range categories {
			if cat.Contains(match.Category) {
				matches = append(matches, match)
				break // 避免同一个敏感词被多个分类匹配而重复添加
			}
		}
	}

	if d.options.MaxDistance > 0 {
		ms := d.gapMatchAll(processedText, mapping)
		for _, match := range ms {
			for _, cat := range categories {
				if cat.Contains(match.Category) {
					matches = append(matches, match)
					break
				}
			}
		}
	}
	if d.options.EnableURLCheck || d.options.EnableEmailCheck {
		extra := d.regexMatches(text)
		for _, match := range extra {
			for _, cat := range categories {
				if cat.Contains(match.Category) {
					matches = append(matches, match)
					break
				}
			}
		}
	}
	return matches
}
