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

// BuildDir 根据 dir 生成不同的 Config
//
// dir 为配置目录的地址，可以带以下的特殊前缀：
//   - ~ 表示系统提供的配置文件目录，比如 Linux 的 XDG_CONFIG、Window 的 AppData 等；
//   - @ 表示当前程序的主目录；
//   - ^ 表示绝对路径；
//   - # 表示工作路径；
//   - 其它则是直接采用 [Dir] 初始化。
//
// 这是对 [SystemDir]、[AppDir]、[Dir] 和 [WDDir] 的合并处理。
func BuildDir(dir string) (*Config, error) {
	if len(dir) == 0 {
		return Dir(dir), nil
	}

	switch dir[0] {
	case '@':
		return AppDir(dir[1:])
	case '~':
		return SystemDir(dir[1:])
	case '^':
		return Dir(dir[1:]), nil
	case '#':
		return WDDir(dir[1:])
	default:
		return Dir(dir), nil
	}
}

// SystemDir 将系统提供的配置目录下的 dir 作为配置目录
//
// dir 相对的目录名称；
// 根据 [os.UserConfigDir] 定位目录。
func SystemDir(dir string) (*Config, error) { return New(dir, os.UserConfigDir) }

// AppDir 将应用程序下的 dir 作为配置文件的保存目录
//
// dir 相对的目录名称；
// 根据 [os.Executable] 定位目录。
func AppDir(dir string) (*Config, error) {
	return New(dir, func() (string, error) {
		ex, err := os.Executable()
		return filepath.Dir(ex), err
	})
}

// WDDir 将工作目录作为配置文件的保存目录
//
// dir 相对的目录名称；
// 根据 [os.Getwd] 定位目录。
func WDDir(dir string) (*Config, error) { return New(dir, os.Getwd) }

// Dir 以指定的目录作为配置文件的保存位置
func Dir(dir string) *Config {
	c, _ := New(dir, nil)
	return c
}

// New 声明 Config 对象
//
// dir 表示当前项目的配置文件存放的目录名称，一般和项目名称相同；
// parent 表示获取系统中用于存放配置文件的目录，比如 Linux 中的 XDG_CONFIG 等目录。
// 用户可以根据自己的需求自行实现该方法，如果为 nil，表示直接将 dir 作为全路径进行处理。
func New(dir string, parent func() (string, error)) (*Config, error) {
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

// Dir 配置文件的目录
func (f *Config) Dir() string { return f.dir }

// Load 加载指定名称的文件内容至 v
//
// 根据文件扩展名决定采用什么编码方法；
// name 为文件名，相对于项目的文件夹；
func (f *Config) Load(name string, v interface{}) error {
	return f.s.Unmarshal(filepath.Join(f.Dir(), name), v)
}

// Read 读取文件的原始内容
func (f *Config) Read(name string) ([]byte, error) { return fs.ReadFile(f, name) }

// Save 将 v 解码并保存至 name 中
//
// 根据文件扩展名决定采用什么编码方法；
// mode 表示文件的权限，仅对新建文件时有效；
func (f *Config) Save(name string, v interface{}, mode fs.FileMode) error {
	return f.s.Marshal(filepath.Join(f.Dir(), name), v, mode)
}
