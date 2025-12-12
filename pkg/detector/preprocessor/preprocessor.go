package preprocessor

import (
	"unicode"

	"github.com/kirklin/go-swd/pkg/core"
)

// Preprocessor 文本预处理器
type Preprocessor struct {
	options core.SWDOptions
}

// NewPreprocessor 创建新的预处理器实例
func NewPreprocessor(options core.SWDOptions) *Preprocessor {
	return &Preprocessor{
		options: options,
	}
}

// Process 处理文本
func (p *Preprocessor) Process(text string) string {
	if text == "" {
		return text
	}

	// 转换为rune切片以正确处理Unicode字符
	runes := []rune(text)
	result := make([]rune, 0, len(runes))

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		// 忽略大小写
		if p.options.IgnoreCase {
			r = unicode.ToLower(r)
		}

		// 忽略空白字符
		if p.options.SkipWhitespace && unicode.IsSpace(r) {
			continue
		}

		// 全角转半角
		if p.options.IgnoreWidth {
			if r > 0xFF00 && r < 0xFF5F {
				r = r - 0xFEE0
			}
		}

		// 数字样式统一
		if p.options.IgnoreNumStyle && (unicode.IsNumber(r) || isChineseNumber(r)) {
			r = p.normalizeNumber(r)
		}

		result = append(result, r)
	}

	return string(result)
}

func (p *Preprocessor) ProcessWithMap(text string) (string, []int) {
	if text == "" {
		return text, nil
	}
	if !p.options.IgnoreCase && !p.options.SkipWhitespace && !p.options.IgnoreWidth && !p.options.IgnoreNumStyle {
		return text, nil
	}
	runes := []rune(text)
	result := make([]rune, 0, len(runes))
	mapping := make([]int, 0, len(runes))
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if p.options.IgnoreCase {
			r = unicode.ToLower(r)
		}
		if p.options.SkipWhitespace && unicode.IsSpace(r) {
			continue
		}
		if p.options.IgnoreWidth {
			if r > 0xFF00 && r < 0xFF5F {
				r = r - 0xFEE0
			}
		}
		if p.options.IgnoreNumStyle && (unicode.IsNumber(r) || isChineseNumber(r)) {
			r = p.normalizeNumber(r)
		}
		result = append(result, r)
		mapping = append(mapping, i)
	}
	return string(result), mapping
}

// isChineseNumber 判断是否是中文数字
func isChineseNumber(r rune) bool {
	chineseNumbers := map[rune]bool{
		'零': true, '〇': true,
		'一': true, '二': true, '三': true, '四': true, '五': true,
		'六': true, '七': true, '八': true, '九': true, '十': true,
	}
	return chineseNumbers[r]
}

// normalizeNumber 将各种数字字符统一为ASCII数字
func (p *Preprocessor) normalizeNumber(r rune) rune {
	switch {
	case r >= '0' && r <= '9':
		return r
	case r >= '０' && r <= '９': // 全角数字
		return r - '０' + '0'
	case r >= '⓪' && r <= '⑨': // 带圈数字
		return r - '⓪' + '0'
	case r >= '①' && r <= '⑨': // 带圈数字（另一种）
		return r - '①' + '1'
	case r >= '㈠' && r <= '㈩': // 带括号汉字数字
		return r - '㈠' + '1'
	case r == '零' || r == '〇':
		return '0'
	case r == '一':
		return '1'
	case r == '二':
		return '2'
	case r == '三':
		return '3'
	case r == '四':
		return '4'
	case r == '五':
		return '5'
	case r == '六':
		return '6'
	case r == '七':
		return '7'
	case r == '八':
		return '8'
	case r == '九':
		return '9'
	case r == '十':
		return r
	default:
		return r
	}
}
