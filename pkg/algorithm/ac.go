package algorithm

import (
	"log"

	"github.com/kirklin/go-swd/pkg/core"
	"github.com/kirklin/go-swd/pkg/types/category"
)

// AhoCorasickNode Aho-Corasick算法节点
type AhoCorasickNode struct {
	children map[rune]*AhoCorasickNode    // 子节点映射
	failLink *AhoCorasickNode             // 失败指针
	isEnd    bool                         // 是否是单词结尾
	word     string                       // 如果是结尾节点，存储完整词
	category category.Category            // 敏感词分类
	parent   *AhoCorasickNode             // 父节点 (用于重建单词)
	depth    int                          // 在字典树中的深度
	outputs  map[string]category.Category // 输出集合：在当前位置可匹配的所有模式（包含沿 fail 链继承的后缀模式）
}

// newAhoCorasickNode 创建新的Aho-Corasick算法节点
// 返回值：初始化的节点，含空的 children 与 outputs 映射
func newAhoCorasickNode() *AhoCorasickNode {
	return &AhoCorasickNode{
		children: make(map[rune]*AhoCorasickNode),
		outputs:  make(map[string]category.Category),
	}
}

// AhoCorasick Aho-Corasick算法实现
type AhoCorasick struct {
	root  *AhoCorasickNode
	built bool // 是否已构建失败指针
}

// NewAhoCorasick 创建新的Aho-Corasick算法实例
// 返回值：根节点已初始化的 AC 自动机
func NewAhoCorasick() *AhoCorasick {
	return &AhoCorasick{
		root: newAhoCorasickNode(),
	}
}

// Type 返回算法类型
// 用于算法选择与工厂注入
func (ac *AhoCorasick) Type() core.AlgorithmType {
	return core.AlgorithmAhoCorasick
}

// Build 构建 Aho-Corasick 算法词库
// 参数：words 为 敏感词->分类 的映射
// 过程：插入所有词并一次性构建失败指针与输出集合（含 fail 链继承）
// 返回：构建错误（通常为 nil）
func (ac *AhoCorasick) Build(words map[string]category.Category) error {
	ac.root = newAhoCorasickNode()
	for word, category := range words {
		ac.insert(word, category)
	}

	ac.buildFailureLinks()
	return nil
}

// insert 向自动机中添加一个词
// 参数：word 敏感词，category 分类
// 说明：仅插入前缀路径并在终止节点登记 outputs，需后续统一构建失败指针
func (ac *AhoCorasick) insert(word string, category category.Category) {
	if word == "" {
		return
	}

	current := ac.root
	for i, char := range word {
		if _, exists := current.children[char]; !exists {
			current.children[char] = newAhoCorasickNode()
			current.children[char].parent = current
			current.children[char].depth = i + 1
		}
		current = current.children[char]
	}

	current.isEnd = true
	current.word = word
	current.category = category
	current.outputs[word] = category
	ac.built = false // 需要重新构建失败指针
}

// buildFailureLinks 构建失败指针
// 过程：BFS 设定各节点 failLink，并将 failLink.outputs 合并到当前节点的 outputs
// 作用：实现多模式后缀匹配，提高多词重叠与后缀场景的匹配精度
func (ac *AhoCorasick) buildFailureLinks() {
	if ac.built {
		return
	}

	// 使用BFS构建失败指针
	queue := make([]*AhoCorasickNode, 0)

	// 先处理根节点的子节点
	for _, child := range ac.root.children {
		child.failLink = ac.root
		queue = append(queue, child)
	}

	// 处理剩余节点
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for char, child := range current.children {
			queue = append(queue, child)

			// 寻找失败指针
			failNode := current.failLink
			for failNode != nil {
				if next, exists := failNode.children[char]; exists {
					child.failLink = next
					break
				}
				failNode = failNode.failLink
			}
			if failNode == nil {
				child.failLink = ac.root
			}
			if child.failLink != nil {
				for w, cat := range child.failLink.outputs {
					child.outputs[w] = cat
				}
			}
		}
	}

	ac.built = true
}

// Match 查找文本中的第一个匹配
// 参数：text 输入文本
// 返回：首个匹配的敏感词结构（包含词、起止位置与分类）；无匹配返回 nil
// 选择策略：优先更靠前起点，起点相同取更长终点
func (ac *AhoCorasick) Match(text string) *core.SensitiveWord {
	if !ac.built {
		ac.buildFailureLinks()
	}

	current := ac.root
	runes := []rune(text)

	for pos, char := range runes {
		// 查找下一个状态
		for current != ac.root && current.children[char] == nil {
			current = current.failLink
		}

		if next, exists := current.children[char]; exists {
			current = next
		} else {
			continue
		}

		if len(current.outputs) > 0 {
			var best *core.SensitiveWord
			for w, cat := range current.outputs {
				l := len([]rune(w))
				startPos := pos - l + 1
				m := core.SensitiveWord{Word: w, StartPos: startPos, EndPos: pos + 1, Category: cat}
				if best == nil || m.StartPos < best.StartPos || (m.StartPos == best.StartPos && m.EndPos > best.EndPos) {
					tmp := m
					best = &tmp
				}
			}
			if best != nil {
				return best
			}
		}
	}

	return nil
}

// MatchAll 返回文本中所有敏感词
// 参数：text 输入文本
// 返回：所有命中敏感词的列表（未去重与未裁剪重叠）
func (ac *AhoCorasick) MatchAll(text string) []core.SensitiveWord {
	if !ac.built {
		ac.buildFailureLinks()
	}

	var matches []core.SensitiveWord
	current := ac.root
	runes := []rune(text)

	for pos, char := range runes {
		// 查找下一个状态
		for current != ac.root && current.children[char] == nil {
			current = current.failLink
		}

		if next, exists := current.children[char]; exists {
			current = next
		} else {
			continue
		}

		if len(current.outputs) > 0 {
			for w, cat := range current.outputs {
				l := len([]rune(w))
				startPos := pos - l + 1
				match := core.SensitiveWord{Word: w, StartPos: startPos, EndPos: pos + 1, Category: cat}
				matches = append(matches, match)
			}
		}
	}

	return matches
}

// Replace 替换敏感词
// 参数：text 输入文本，replacement 替换用字符
// 返回：完成敏感词位置替换后的文本
func (ac *AhoCorasick) Replace(text string, replacement rune) string {
	matches := ac.MatchAll(text)
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
// 参数：text 输入文本
// 返回：是否命中任一敏感词
func (ac *AhoCorasick) Detect(text string) bool {
	return ac.Match(text) != nil
}

// OnWordsChanged 词库变更回调
// 参数：words 最新词库映射
// 行为：同步重建自动机；错误仅记录日志（回调不可中断调用方）
func (ac *AhoCorasick) OnWordsChanged(words map[string]category.Category) {
	if err := ac.Build(words); err != nil {
		// 这里只能记录错误,因为是回调方法
		log.Printf("重建算法失败: %v", err)
	}
}

// AddWord 增量添加词
// 参数：word 待添加敏感词，cat 分类
// 行为：沿路径增量创建节点并设置 failLink；合并 failLink.outputs 到新增节点的 outputs；在终止节点登记当前词输出
// 适用：在 Composite 增量构建中最小化重建开销
func (ac *AhoCorasick) AddWord(word string, cat category.Category) error {
	if word == "" {
		return nil
	}
	current := ac.root
	runes := []rune(word)
	for i, r := range runes {
		child := current.children[r]
		if child == nil {
			child = newAhoCorasickNode()
			child.parent = current
			child.depth = i + 1
			// 设置增量 fail 链
			failNode := current.failLink
			for failNode != nil {
				if next, ok := failNode.children[r]; ok {
					child.failLink = next
					break
				}
				failNode = failNode.failLink
			}
			if failNode == nil {
				child.failLink = ac.root
			}
			if child.failLink != nil {
				for w, c := range child.failLink.outputs {
					child.outputs[w] = c
				}
			}
			current.children[r] = child
		}
		current = child
	}
	current.isEnd = true
	current.word = word
	current.category = cat
	current.outputs[word] = cat
	return nil
}

// RemoveWord 增量删除词
// 参数：word 待删除敏感词
// 行为：清除终止标记与词内容；若末端为叶子则向上修剪不再使用的分支；不影响其他分支的 fail 链
func (ac *AhoCorasick) RemoveWord(word string) error {
	if word == "" {
		return nil
	}
	current := ac.root
	runes := []rune(word)
	path := make([]*AhoCorasickNode, 0, len(runes)+1)
	path = append(path, current)
	for _, r := range runes {
		child := current.children[r]
		if child == nil {
			return nil
		}
		current = child
		path = append(path, current)
	}
	if !current.isEnd {
		return nil
	}
	current.isEnd = false
	current.word = ""
	// 修剪无子女的叶子（不影响其他节点 fail 链）
	for i := len(runes) - 1; i >= 0; i-- {
		node := path[i+1]
		parent := path[i]
		if node.isEnd || len(node.children) > 0 {
			break
		}
		delete(parent.children, runes[i])
	}
	return nil
}
