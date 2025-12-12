package main

import (
	"fmt"
	"log"

	"github.com/kirklin/go-swd"
)

func main() {
	// 1. 创建实例
	detector, err := swd.New()
	if err != nil {
		log.Fatal(err)
	}
	defer detector.Close()

	// 2. 添加自定义敏感词（可选）
	customWords := map[string]swd.Category{
		"涉黄":    swd.Pornography,    // 涉黄分类
		"涉政":    swd.Political,      // 涉政分类
		"赌博词汇":  swd.Gambling,       // 赌博分类
		"毒品词汇":  swd.Drugs,          // 毒品分类
		"脏话词汇":  swd.Profanity,      // 脏话分类
		"歧视词汇":  swd.Discrimination, // 歧视分类
		"诈骗词汇":  swd.Scam,           // 诈骗分类
		"自定义词汇": swd.Custom,         // 自定义分类
	}
	if err := detector.AddWords(customWords); err != nil {
		log.Fatal(err)
	}

	// 3. 基本检测
	text := "这是一段包含敏感词涉黄和涉政的文本"
	fmt.Println("是否包含敏感词:", detector.Detect(text))

	// 4. 检测指定分类
	fmt.Println("是否包含涉黄内容:", detector.DetectIn(text, swd.Pornography))
	fmt.Println("是否包含涉政内容:", detector.DetectIn(text, swd.Political))
	fmt.Println("是否包含赌博内容:", detector.DetectIn(text, swd.Gambling))
	fmt.Println("是否包含毒品内容:", detector.DetectIn(text, swd.Drugs))

	// 5. 检测多个分类
	fmt.Println("是否包含涉黄或涉政内容:", detector.DetectIn(text, swd.Pornography, swd.Political))
	fmt.Println("是否包含任意预定义分类:", detector.DetectIn(text, swd.All))

	// 6. 获取匹配结果
	if word := detector.Match(text); word != nil {
		fmt.Printf("首个敏感词: %s (分类: %s)\n", word.Word, word.Category)
	}

	// 7. 获取所有匹配
	words := detector.MatchAll(text)
	for _, word := range words {
		fmt.Printf("敏感词: %s (分类: %s, 位置: %d-%d)\n",
			word.Word, word.Category, word.StartPos, word.EndPos)
	}

	// 8. 敏感词过滤
	filtered := detector.ReplaceWithAsterisk(text) // 使用 * 替换
	fmt.Println("过滤后的文本:", filtered)

	// 9. 自定义替换策略
	customFiltered := detector.ReplaceWithStrategy(text, func(word swd.SensitiveWord) string {
		return fmt.Sprintf("[%s]", word.Category) // 替换为分类名
	})
	fmt.Println("自定义替换后的文本:", customFiltered)

	// 10. 移除敏感词
	if err := detector.RemoveWord("自定义敏感词1"); err != nil {
		log.Printf("移除敏感词失败: %v", err)
	}

	// 11. 清空词库
	if err := detector.Clear(); err != nil {
		log.Printf("清空词库失败: %v", err)
	}
}
