package swd

import (
	"context"
	"fmt"

	"github.com/kirklin/go-swd/pkg/algorithm"
	"github.com/kirklin/go-swd/pkg/core"
	"github.com/kirklin/go-swd/pkg/detector"
	"github.com/kirklin/go-swd/pkg/dictionary"
	"github.com/kirklin/go-swd/pkg/filter"
)

// DefaultFactory 默认组件工厂实现
type DefaultFactory struct {
	algoType core.AlgorithmType
}

// NewDefaultFactory 创建默认工厂实例
func NewDefaultFactory() ComponentFactory {
	return &DefaultFactory{algoType: core.AlgorithmAhoCorasick}
}

func NewDefaultFactoryWithAlgorithm(t core.AlgorithmType) ComponentFactory {
	return &DefaultFactory{algoType: t}
}

// CreateDetector 创建检测器实例
func (f *DefaultFactory) CreateDetector(options *core.SWDOptions, loader core.Loader) (core.Detector, error) {
	var algo core.Algorithm
	typ := f.algoType
	if options != nil && options.Algorithm != "" {
		typ = options.Algorithm
	}
	algo = algorithm.NewComposite(typ)
	d, err := detector.NewDetector(*options, loader, algo)
	if err != nil {
		return nil, fmt.Errorf("创建检测器失败: %w", err)
	}
	return d, nil
}

// CreateFilter 创建过滤器实例
func (f *DefaultFactory) CreateFilter(detector core.Detector) core.Filter {
	return filter.NewFilter(detector)
}

// CreateLoader 创建加载器实例
func (f *DefaultFactory) CreateLoader() core.Loader {
	loader := dictionary.NewLoader()
	return loader
}

// CreateComponents 创建并关联所有组件
func (f *DefaultFactory) CreateComponents(options *core.SWDOptions) (core.Detector, core.Filter, core.Loader, error) {
	// 创建加载器
	loader := f.CreateLoader()

	// 加载默认词库
	if err := loader.LoadDefaultWords(context.Background()); err != nil {
		return nil, nil, nil, fmt.Errorf("加载默认词库失败: %w", err)
	}

	// 创建检测器
	d, err := f.CreateDetector(options, loader)
	if err != nil {
		return nil, nil, nil, err
	}

	// 注册检测器为加载器的观察者
	if observer, ok := d.(core.Observer); ok {
		loader.(*dictionary.Loader).AddObserver(observer)
	}

	// 创建过滤器
	filter := f.CreateFilter(d)

	return d, filter, loader, nil
}
