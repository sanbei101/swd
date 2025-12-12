package swd

import "github.com/kirklin/go-swd/pkg/core"

// WithOptions 设置所有配置选项
func (swd *SWD) WithOptions(options core.SWDOptions) *SWD {
	if swd.options == nil {
		swd.options = &core.SWDOptions{}
	}
	*swd.options = options
	return swd
}

// ApplyOptions 将当前选项应用到检测器
func (swd *SWD) ApplyOptions() *SWD {
	if swd.options == nil {
		swd.options = &core.SWDOptions{}
	}
	if swd.detector != nil {
		swd.detector.SetOptions(*swd.options)
	}
	return swd
}

// WithSkipWhitespace 设置是否忽略空白字符
func (swd *SWD) WithSkipWhitespace(skip bool) *SWD {
	if swd.options == nil {
		swd.options = &core.SWDOptions{}
	}
	swd.options.SkipWhitespace = skip
	return swd
}

// WithIgnoreCase 设置是否忽略大小写
func (swd *SWD) WithIgnoreCase(ignore bool) *SWD {
	if swd.options == nil {
		swd.options = &core.SWDOptions{}
	}
	swd.options.IgnoreCase = ignore
	return swd
}

// WithIgnoreWidth 设置是否忽略全角和半角字符差异
func (swd *SWD) WithIgnoreWidth(ignore bool) *SWD {
	if swd.options == nil {
		swd.options = &core.SWDOptions{}
	}
	swd.options.IgnoreWidth = ignore
	return swd
}

// WithMaxDistance 设置字符间最大距离
func (swd *SWD) WithMaxDistance(distance int) *SWD {
	if swd.options == nil {
		swd.options = &core.SWDOptions{}
	}
	swd.options.MaxDistance = distance
	return swd
}

// EnablePinyin 启用拼音检测
func (swd *SWD) EnablePinyin() *SWD {
	if swd.options == nil {
		swd.options = &core.SWDOptions{}
	}
	swd.options.EnablePinyin = true
	return swd
}

// DisablePinyin 禁用拼音检测
func (swd *SWD) DisablePinyin() *SWD {
	if swd.options == nil {
		swd.options = &core.SWDOptions{}
	}
	swd.options.EnablePinyin = false
	return swd
}

// EnableHomophone 启用同音字检测
func (swd *SWD) EnableHomophone() *SWD {
	if swd.options == nil {
		swd.options = &core.SWDOptions{}
	}
	swd.options.EnableHomophone = true
	return swd
}

// DisableHomophone 禁用同音字检测
func (swd *SWD) DisableHomophone() *SWD {
	if swd.options == nil {
		swd.options = &core.SWDOptions{}
	}
	swd.options.EnableHomophone = false
	return swd
}

// EnableNumCheck 启用数字检测
func (swd *SWD) EnableNumCheck() *SWD {
	if swd.options == nil {
		swd.options = &core.SWDOptions{}
	}
	swd.options.EnableNumCheck = true
	return swd
}

// DisableNumCheck 禁用数字检测
func (swd *SWD) DisableNumCheck() *SWD {
	if swd.options == nil {
		swd.options = &core.SWDOptions{}
	}
	swd.options.EnableNumCheck = false
	return swd
}

// EnableURLCheck 启用URL检测
func (swd *SWD) EnableURLCheck() *SWD {
	if swd.options == nil {
		swd.options = &core.SWDOptions{}
	}
	swd.options.EnableURLCheck = true
	return swd
}

// DisableURLCheck 禁用URL检测
func (swd *SWD) DisableURLCheck() *SWD {
	if swd.options == nil {
		swd.options = &core.SWDOptions{}
	}
	swd.options.EnableURLCheck = false
	return swd
}

// EnableEmailCheck 启用Email检测
func (swd *SWD) EnableEmailCheck() *SWD {
	if swd.options == nil {
		swd.options = &core.SWDOptions{}
	}
	swd.options.EnableEmailCheck = true
	return swd
}

// DisableEmailCheck 禁用Email检测
func (swd *SWD) DisableEmailCheck() *SWD {
	if swd.options == nil {
		swd.options = &core.SWDOptions{}
	}
	swd.options.EnableEmailCheck = false
	return swd
}
