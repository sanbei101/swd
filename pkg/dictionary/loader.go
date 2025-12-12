package dictionary

import (
	"bufio"
	"context"
	_ "embed"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kirklin/go-swd/pkg/core"
	"github.com/kirklin/go-swd/pkg/types/category"
)

// Loader 实现core.Loader接口
type Loader struct {
	words           sync.Map
	observers       sync.Map
	notifyBatchSize int
	lastNotifyTime  atomic.Value // time.Time
	notifyInterval  time.Duration
}

// NewLoader 创建新的加载器实例
func NewLoader() *Loader {
	l := &Loader{
		notifyBatchSize: 100,
		notifyInterval:  time.Millisecond * 100,
	}
	l.lastNotifyTime.Store(time.Now())
	return l
}

func (l *Loader) SetNotifyInterval(d time.Duration) {
	l.notifyInterval = d
}

func (l *Loader) SetNotifyBatchSize(size int) {
	if size > 0 {
		l.notifyBatchSize = size
	}
}

//go:embed default/political.txt
var politicalWords string

//go:embed default/pornography.txt
var pornographyWords string

//go:embed default/violence.txt
var violenceWords string

//go:embed default/gambling.txt
var gamblingWords string

//go:embed default/drugs.txt
var drugsWords string

//go:embed default/profanity.txt
var profanityWords string

//go:embed default/discrimination.txt
var discriminationWords string

//go:embed default/scam.txt
var scamWords string

//go:embed default/all.txt
var allWords string

// 严格词库（更通用的常见词）
//
//go:embed strict/political.txt
var strictPoliticalWords string

//go:embed strict/pornography.txt
var strictPornographyWords string

//go:embed strict/violence.txt
var strictViolenceWords string

//go:embed strict/gambling.txt
var strictGamblingWords string

//go:embed strict/drugs.txt
var strictDrugsWords string

//go:embed strict/profanity.txt
var strictProfanityWords string

//go:embed strict/discrimination.txt
var strictDiscriminationWords string

//go:embed strict/scam.txt
var strictScamWords string

//go:embed strict/all.txt
var strictAllWords string

// LoadDefaultWords 加载默认词库
func (l *Loader) LoadDefaultWords(ctx context.Context) error {
	// 加载所有分类词典
	categories := map[string]struct {
		content string
		cat     category.Category
	}{
		"political.txt":      {content: politicalWords, cat: category.Political},
		"pornography.txt":    {content: pornographyWords, cat: category.Pornography},
		"violence.txt":       {content: violenceWords, cat: category.Violence},
		"gambling.txt":       {content: gamblingWords, cat: category.Gambling},
		"drugs.txt":          {content: drugsWords, cat: category.Drugs},
		"profanity.txt":      {content: profanityWords, cat: category.Profanity},
		"discrimination.txt": {content: discriminationWords, cat: category.Discrimination},
		"scam.txt":           {content: scamWords, cat: category.Scam},
	}

	for filename, data := range categories {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := l.loadFromString(ctx, data.content, data.cat); err != nil {
				return fmt.Errorf("failed to load %s: %w", filename, err)
			}
		}
	}

	// 加载通用词典
	if err := l.loadFromString(ctx, allWords, category.None); err != nil {
		return fmt.Errorf("failed to load all.txt: %w", err)
	}

	// 加载严格词库（补全常见词）
	strict := map[string]struct {
		content string
		cat     category.Category
	}{
		"strict/political.txt":      {content: strictPoliticalWords, cat: category.Political},
		"strict/pornography.txt":    {content: strictPornographyWords, cat: category.Pornography},
		"strict/violence.txt":       {content: strictViolenceWords, cat: category.Violence},
		"strict/gambling.txt":       {content: strictGamblingWords, cat: category.Gambling},
		"strict/drugs.txt":          {content: strictDrugsWords, cat: category.Drugs},
		"strict/profanity.txt":      {content: strictProfanityWords, cat: category.Profanity},
		"strict/discrimination.txt": {content: strictDiscriminationWords, cat: category.Discrimination},
		"strict/scam.txt":           {content: strictScamWords, cat: category.Scam},
	}

	for filename, data := range strict {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := l.loadFromString(ctx, data.content, data.cat); err != nil {
				return fmt.Errorf("failed to load %s: %w", filename, err)
			}
		}
	}

	if err := l.loadFromString(ctx, strictAllWords, category.None); err != nil {
		return fmt.Errorf("failed to load strict/all.txt: %w", err)
	}

	l.notifyObserversIfNeeded(true)
	return nil
}

// LoadCustomWords 加载自定义词库
func (l *Loader) LoadCustomWords(ctx context.Context, words map[string]category.Category) error {
	const batchSize = 1000
	count := 0

	for word, cat := range words {
		if count%batchSize == 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}
		if err := l.addWordInternal(word, cat); err != nil {
			return err
		}
		count++
	}

	l.notifyObserversIfNeeded(true)
	return nil
}

// AddObserver 添加观察者
func (l *Loader) AddObserver(observer core.Observer) {
	l.observers.Store(observer, struct{}{})
}

// RemoveObserver 移除观察者
func (l *Loader) RemoveObserver(observer core.Observer) {
	l.observers.Delete(observer)
}

// notifyObserversIfNeeded 根据条件通知观察者
func (l *Loader) notifyObserversIfNeeded(force bool) {
	if !force {
		lastNotify := l.lastNotifyTime.Load().(time.Time)
		if time.Since(lastNotify) < l.notifyInterval {
			return
		}
	}

	words := l.GetWords()
	l.observers.Range(func(key, value interface{}) bool {
		if observer, ok := key.(core.Observer); ok {
			observer.OnWordsChanged(words)
		}
		return true
	})

	l.lastNotifyTime.Store(time.Now())
}

// AddWord 添加单个敏感词
func (l *Loader) AddWord(word string, cat category.Category) error {
	if err := l.addWordInternal(word, cat); err != nil {
		return err
	}
	l.notifyObserversIfNeeded(true)
	return nil
}

// addWordInternal 内部添加词方法
func (l *Loader) addWordInternal(word string, cat category.Category) error {
	if word = strings.TrimSpace(word); word == "" {
		return fmt.Errorf("word cannot be empty")
	}

	// 验证分类的有效性
	if !cat.IsValid() {
		return fmt.Errorf("invalid category: %v", cat)
	}

	// 如果词已存在且有效分类，且当前要设置的是 None 分类，则保留原有分类
	if val, exists := l.words.Load(word); exists {
		if existingCat, ok := val.(category.Category); ok && existingCat != category.None && cat == category.None {
			return nil
		}
	}

	l.words.Store(word, cat)
	return nil
}

// AddWords 批量添加敏感词
func (l *Loader) AddWords(words map[string]category.Category) error {
	for word, cat := range words {
		if err := l.addWordInternal(word, cat); err != nil {
			return err
		}
	}
	l.notifyObserversIfNeeded(true)
	return nil
}

// RemoveWord 移除单个敏感词
func (l *Loader) RemoveWord(word string) error {
	l.words.Delete(word)
	l.notifyObserversIfNeeded(true)
	return nil
}

// RemoveWords 批量移除敏感词
func (l *Loader) RemoveWords(words []string) error {
	for _, word := range words {
		l.words.Delete(word)
	}
	l.notifyObserversIfNeeded(true)
	return nil
}

// Clear 清空所有敏感词
func (l *Loader) Clear() error {
	l.words = sync.Map{}
	l.notifyObserversIfNeeded(true)
	return nil
}

// loadFromString 从字符串加载敏感词
func (l *Loader) loadFromString(ctx context.Context, content string, cat category.Category) error {
	reader := strings.NewReader(content)
	return l.loadFromReader(ctx, reader, cat)
}

// loadFromReader 从Reader加载敏感词
func (l *Loader) loadFromReader(ctx context.Context, reader io.Reader, cat category.Category) error {
	scanner := bufio.NewScanner(reader)
	const batchSize = 1000
	count := 0

	for scanner.Scan() {
		if count%batchSize == 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}

		word := strings.TrimSpace(scanner.Text())
		if word == "" || strings.HasPrefix(word, "#") {
			continue
		}
		if err := l.addWordInternal(word, cat); err != nil {
			return err
		}
		count++
	}
	return scanner.Err()
}

// GetWords 获取所有已加载的敏感词
func (l *Loader) GetWords() map[string]category.Category {
	words := make(map[string]category.Category)
	l.words.Range(func(key, value interface{}) bool {
		if k, ok := key.(string); ok {
			if v, ok := value.(category.Category); ok {
				words[k] = v
			}
		}
		return true
	})
	return words
}
