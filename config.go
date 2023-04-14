// SPDX-License-Identifier: MIT

// Package config 提供了对多种格式配置文件的支持
package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// Config 项目的配置文件管理
type Config struct {
	fsys fs.FS
	dir  string // 配置文件的根目录
	s    Serializer
}

// New 声明 Config 对象
//
// 这将会在 $CONFIG/vendor/name 目录下操作项目的配置文件。
// $CONFIG 为系统规定的配置目录，比如 $XDG_CONFIG 等，各系统有所不同。
//
// name 为应用的名称；
// vendor 表示厂商名称，可以为空；
func New(name, vendor string) (*Config, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	return newConfig(filepath.Join(dir, vendor, name)), nil
}

func newConfig(dir string) *Config {
	return &Config{
		fsys: os.DirFS(dir),
		dir:  dir,
		s:    make(Serializer, 5),
	}
}

// Exists 是否存在指定的文件
func (f *Config) Exists(name string) bool {
	_, err := fs.Stat(f.fsys, name)
	return !errors.Is(err, fs.ErrNotExist)
}

// Open 实现 [fs.FS] 接口
func (f *Config) Open(name string) (fs.File, error) { return f.fsys.Open(name) }

// Load 加载指定名称的文件内容至 v
//
// 根据文件扩展名决定采用什么编码方法；
// name 为文件名，相对于项目的文件夹；
func (f *Config) Load(name string, v interface{}) error {
	s := f.s.searchByExt(name)
	if s == nil {
		return ErrSerializerNotFound()
	}

	data, err := f.Read(name)
	if err != nil {
		return err
	}
	return s.Unmarshal(data, v)
}

// Read 读取文件的原始内容
func (f *Config) Read(name string) ([]byte, error) { return fs.ReadFile(f, name) }

// Save 将 v 解码并保存至 name 中
//
// 根据文件扩展名决定采用什么编码方法；
// mode 表示文件的权限，仅对新建文件时有效；
func (f *Config) Save(name string, v interface{}, mode fs.FileMode) error {
	s := f.s.searchByExt(name)
	if s == nil {
		return ErrSerializerNotFound()
	}

	data, err := s.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(f.dir, name), data, mode)
}
