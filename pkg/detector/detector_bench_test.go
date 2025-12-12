package detector

import (
    "context"
    "testing"

    "github.com/kirklin/go-swd/pkg/algorithm"
    "github.com/kirklin/go-swd/pkg/core"
    "github.com/kirklin/go-swd/pkg/dictionary"
)

func BenchmarkDetectShort(b *testing.B) {
    loader := dictionary.NewLoader()
    _ = loader.LoadDefaultWords(context.Background())
    algo := algorithm.NewAhoCorasick()
    d, _ := NewDetector(core.SWDOptions{IgnoreCase: true}, loader, algo)
    text := "这是一段包含色情的文本"
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = d.Detect(text)
    }
}

func BenchmarkDetectLong(b *testing.B) {
    loader := dictionary.NewLoader()
    _ = loader.LoadDefaultWords(context.Background())
    algo := algorithm.NewAhoCorasick()
    d, _ := NewDetector(core.SWDOptions{IgnoreCase: true}, loader, algo)
    long := "这是一段很长的文本，包含了多个敏感词：色情、暴力、政府、赌博、毒品、脏话、歧视、诈骗。"
    for i := 0; i < 12; i++ {
        long += long
    }
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = d.Detect(long)
    }
}