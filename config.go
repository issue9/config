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

// SystemDir 将系统提供的配置目录下的 dir 作为配置目录
//
// dir 相对的目录名称；
// 根据 [os.UserConfigDir] 定位目录。
func SystemDir(dir string) (*Config, error) {
	return newConfig(dir, os.UserConfigDir)
}

// AppDir 将应用程序下的 dir 作为配置文件的保存目录
//
// dir 相对的目录名称；
// 根据 [os.Executable] 定位目录。
func AppDir(dir string) (*Config, error) {
	return newConfig(dir, os.Executable)
}

// Dir 以指定的目录作为配置文件的保存位置
func Dir(dir string) *Config {
	c, _ := newConfig(dir, nil)
	return c
}

func newConfig(dir string, parent func() (string, error)) (*Config, error) {
	if parent != nil {
		p, err := parent()
		if err != nil {
			return nil, err
		}

		if p != "" {
			dir = filepath.Join(p, dir)
		}
	}

	return &Config{
		fsys: os.DirFS(dir),
		dir:  dir,
		s:    make(Serializer, 5),
	}, nil
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
