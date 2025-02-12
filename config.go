// SPDX-FileCopyrightText: 2019-2025 caixw
//
// SPDX-License-Identifier: MIT

//go:generate web locale -l=und -f=yaml ./
//go:generate web update-locale -src=./locales/und.yaml -dest=./locales/zh.yaml

// Package config 提供了对多种格式配置文件的支持
package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// Config 项目的配置文件管理
//
// 相对于 [Serializer]，Config 提供了对目录的保护，只能存储指定目录下的内容。
type Config struct {
	root *os.Root
	s    Serializer
}

// BuildDir 根据 dir 生成不同的 [Config]
//
// dir 为项目的配置文件目录，后续通过 [Config] 操作都被限制在此目录之下。可以带以下的特殊前缀：
//   - ~ 表示系统提供的配置文件目录，比如 Linux 的 XDG_CONFIG、Windows 的 AppData 等；
//   - @ 表示当前程序的主目录；
//   - ^ 表示绝对路径；
//   - # 表示工作路径，这是一个随着工作目录变化的值，使用时需要小心；
//   - 其它则是直接采用 [Dir] 初始化。
//
// 这是对 [SystemDir]、[AppDir]、[Dir] 和 [WDDir] 的合并处理。
func BuildDir(s Serializer, dir string) (*Config, error) {
	if len(dir) == 0 {
		return Dir(s, dir), nil
	}

	switch dir[0] {
	case '@':
		return AppDir(s, dir[1:])
	case '~':
		return SystemDir(s, dir[1:])
	case '^':
		return Dir(s, dir[1:]), nil
	case '#':
		return WDDir(s, dir[1:])
	default:
		return Dir(s, dir), nil
	}
}

// SystemDir 将系统提供的配置目录下的 dir 作为配置目录
//
// dir 相对的 [os.UserConfigDir] 目录名称；
func SystemDir(s Serializer, dir string) (*Config, error) {
	return New(s, dir, os.UserConfigDir)
}

// AppDir 将应用程序下的 dir 作为配置文件的保存目录
//
// dir 相对 [os.Executable] 的目录名称；
func AppDir(s Serializer, dir string) (*Config, error) {
	return New(s, dir, func() (string, error) {
		ex, err := os.Executable()
		return filepath.Dir(ex), err
	})
}

// WDDir 将工作目录作为配置文件的保存目录
//
// dir 相对 [os.Getwd] 的目录名称；
func WDDir(s Serializer, dir string) (*Config, error) {
	return New(s, dir, os.Getwd)
}

// Dir 以指定的目录作为配置文件的保存位置
func Dir(s Serializer, dir string) *Config {
	c, _ := New(s, dir, nil)
	return c
}

// New 声明 [Config] 对象
//
// dir 表示当前项目的配置文件存放的目录名称，如果 parent 不为空，为相对于 parent 返回值的路径；
// parent 表示获取系统中用于存放配置文件的路径，比如 Linux 中的 XDG_CONFIG 等目录。
// 用户可以根据自己的需求自行实现该方法，如果为 nil，表示直接将 dir 作为全路径进行处理。
func New(s Serializer, dir string, parent func() (string, error)) (*Config, error) {
	if parent != nil {
		p, err := parent()
		if err != nil {
			return nil, err
		}

		if p != "" {
			dir = filepath.Join(p, dir)
		}
	}

	if s == nil {
		s = make(Serializer, 5)
	}

	if _, err := os.Stat(dir); errors.Is(err, fs.ErrNotExist) {
		if err := os.MkdirAll(dir, fs.ModePerm); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}

	return &Config{
		root: root,
		s:    s,
	}, nil
}

// Exists 是否存在指定的文件
func (f *Config) Exists(name string) bool {
	_, err := f.Root().Stat(name)
	return !errors.Is(err, fs.ErrNotExist)
}

// Root 配置文件的目录
func (f *Config) Root() *os.Root { return f.root }

// Load 加载指定名称的文件内容至 v
//
// name 为文件名，相对于 [Config.Root]，根据文件扩展名决定采用什么编码方法；
// 如果 v 实现了 [Sanitizer]，在加载之后会调用该接口对数据进行处理；
func (f *Config) Load(name string, v any) error {
	if err := f.s.Unmarshal(f.Root().OpenFile, name, v); err != nil {
		return err
	}

	if s, ok := v.(Sanitizer); ok {
		if fe := s.SanitizeConfig(); fe != nil {
			fe.Path = name
			return fe
		}
	}
	return nil
}

// Read 读取文件的原始内容
func (f *Config) Read(name string) ([]byte, error) { return fs.ReadFile(f.Root().FS(), name) }

// Save 将 v 解码并保存至 name 中
//
// 根据文件扩展名决定采用什么编码方法；
// mode 表示文件的权限，仅对新建文件时有效；
// 如果 v 实现了 [Sanitizer]，在保存之前会调用该接口对数据进行处理；
func (f *Config) Save(name string, v any, mode fs.FileMode) error {
	if s, ok := v.(Sanitizer); ok {
		if fe := s.SanitizeConfig(); fe != nil {
			fe.Path = name
			return fe
		}
	}

	return f.s.Marshal(f.Root().OpenFile, name, v, mode)
}
