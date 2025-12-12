// Package swd 提供了敏感词检测和过滤功能
package swd

import (
	"github.com/kirklin/go-swd/pkg/core"
	"github.com/kirklin/go-swd/pkg/swd"
	"github.com/kirklin/go-swd/pkg/types/category"
)

// 导出核心类型
type (
	// SensitiveWord 表示一个敏感词及其相关信息
	SensitiveWord = core.SensitiveWord
	// Category 表示敏感词的分类
	Category = category.Category
	// SWD 是敏感词检测引擎的主要实现
	SWD = swd.SWD
)

// 导出分类常量
const (
	None           = category.None           // 未分类
	Pornography    = category.Pornography    // 涉黄
	Political      = category.Political      // 涉政
	Violence       = category.Violence       // 暴力
	Gambling       = category.Gambling       // 赌博
	Drugs          = category.Drugs          // 毒品
	Profanity      = category.Profanity      // 脏话
	Discrimination = category.Discrimination // 歧视
	Scam           = category.Scam           // 诈骗
	Custom         = category.Custom         // 自定义
)

// All 所有预定义分类的组合
var All = category.All

// New 创建一个新的敏感词检测引擎
func New() (*SWD, error) {
	factory := swd.NewDefaultFactory()
	return swd.New(factory)
}

// NewWithFactory 使用自定义工厂创建敏感词检测引擎
func NewWithFactory(factory swd.ComponentFactory) (*SWD, error) {
	return swd.New(factory)
}

// ComponentFactory 定义了创建各种组件的工厂接口
type ComponentFactory = swd.ComponentFactory
