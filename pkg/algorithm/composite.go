package algorithm

import (
	"sort"

	"github.com/kirklin/go-swd/pkg/core"
	"github.com/kirklin/go-swd/pkg/types/category"
)

type composite struct {
	typ       core.AlgorithmType                        // 组合内部所选具体算法类型（Trie/AC）
	shards    map[category.Category]core.Algorithm      // 按分类拆分的分片自动机，用于增量构建与分类隔离
	snapshots map[category.Category]map[string]struct{} // 每个分类的词快照（用于计算增删差异）
	global    core.Algorithm                            // 全局自动机（读路径快车道：Detect/Match/MatchAll）
}

// NewComposite 创建组合算法
// 参数：t 指定内部采用的基础算法类型
// 返回：实现 core.Algorithm 的组合实例（含分片与快照初始化）
func NewComposite(t core.AlgorithmType) core.Algorithm {
	return &composite{
		typ:       t,
		shards:    make(map[category.Category]core.Algorithm),
		snapshots: make(map[category.Category]map[string]struct{}),
	}
}

// Type 返回组合内部所选的基础算法类型
func (c *composite) Type() core.AlgorithmType { return c.typ }

// Build 构建组合算法
// 参数：words 为 敏感词->分类 的映射
// 过程：
// 1) 按分类拆分词表；
// 2) 根据分类快照计算增删差异，能增量则对分片调用 Add/Remove，否则重建分片；
// 3) 同步重建全局自动机，作为读路径快车道。
// 返回：构建错误（通常为 nil）
func (c *composite) Build(words map[string]category.Category) error {
	// 按分类拆分词表
	perCat := make(map[category.Category]map[string]category.Category)
	for w, cat := range words {
		per := perCat[cat]
		if per == nil {
			per = make(map[string]category.Category)
			perCat[cat] = per
		}
		per[w] = cat
	}
	// 对每个分类：若快照一致则跳过，否则增量应用或重建
	for cat, wmap := range perCat {
		snap := c.snapshots[cat]
		// 计算差异集
		added := make([]string, 0)
		removed := make([]string, 0)
		seen := make(map[string]struct{}, len(wmap))
		for w := range wmap {
			seen[w] = struct{}{}
		}
		for w := range seen {
			if snap == nil || snap[w] == (struct{}{}) { /* noop to appease linter */
			}
		}
		if snap == nil {
			for w := range wmap {
				added = append(added, w)
			}
		} else {
			for w := range wmap {
				if _, ok := snap[w]; !ok {
					added = append(added, w)
				}
			}
			for w := range snap {
				if _, ok := wmap[w]; !ok {
					removed = append(removed, w)
				}
			}
		}

		if len(added) == 0 && len(removed) == 0 {
			continue
		}

		// 创建或获取子算法（分片自动机）
		algo := c.shards[cat]
		if algo == nil {
			switch c.typ {
			case core.AlgorithmTrie:
				algo = NewTrie()
			default:
				algo = NewAhoCorasick()
			}
			c.shards[cat] = algo
		}

		// 尝试增量，否则重建
		if ma, ok := algo.(core.MutableAlgorithm); ok {
			for _, w := range added {
				_ = ma.AddWord(w, cat)
			}
			for _, w := range removed {
				_ = ma.RemoveWord(w)
			}
		} else {
			if err := algo.Build(wmap); err != nil {
				return err
			}
		}

		// 更新快照
		newSnap := make(map[string]struct{}, len(wmap))
		for w := range wmap {
			newSnap[w] = struct{}{}
		}
		c.snapshots[cat] = newSnap
	}
	// 全量构建全局自动机（用于 Detect/Match/MatchAll 快路径）
	if c.global == nil {
		switch c.typ {
		case core.AlgorithmTrie:
			c.global = NewTrie()
		default:
			c.global = NewAhoCorasick()
		}
	}
	if err := c.global.Build(words); err != nil {
		return err
	}
	return nil
}

// Detect 检查文本是否包含敏感词
// 优先走全局自动机快路径；无全局时遍历分片
func (c *composite) Detect(text string) bool {
	if c.global != nil {
		return c.global.Detect(text)
	}
	for _, algo := range c.shards {
		if algo.Detect(text) {
			return true
		}
	}
	return false
}

// Match 查找首个匹配
// 策略：优先更靠前起点，起点相同取更长终点
func (c *composite) Match(text string) *core.SensitiveWord {
	if c.global != nil {
		return c.global.Match(text)
	}
	var best *core.SensitiveWord
	for _, algo := range c.shards {
		if m := algo.Match(text); m != nil {
			if best == nil || m.StartPos < best.StartPos || (m.StartPos == best.StartPos && m.EndPos > best.EndPos) {
				best = m
			}
		}
	}
	return best
}

// MatchAll 返回所有命中敏感词（未裁剪重叠）
// 无全局时聚合分片结果并做稳定排序（起点升序、终点降序）
func (c *composite) MatchAll(text string) []core.SensitiveWord {
	if c.global != nil {
		return c.global.MatchAll(text)
	}
	var out []core.SensitiveWord
	for _, algo := range c.shards {
		out = append(out, algo.MatchAll(text)...)
	}
	if len(out) == 0 {
		return out
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].StartPos == out[j].StartPos {
			return out[i].EndPos > out[j].EndPos
		}
		return out[i].StartPos < out[j].StartPos
	})
	return out
}

// Replace 将命中区间替换为给定字符
// 过程：先 MatchAll 获取区间，再按区间逐位替换
func (c *composite) Replace(text string, replacement rune) string {
	matches := c.MatchAll(text)
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

// OnWordsChanged 词库变更回调
// 行为：复用 Build 逻辑以保持增量与全局重建一致性
func (c *composite) OnWordsChanged(words map[string]category.Category) {
	_ = c.Build(words)
}
