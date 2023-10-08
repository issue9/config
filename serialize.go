// SPDX-License-Identifier: MIT

package config

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/issue9/localeutil"
)

type UnmarshalFunc func([]byte, any) error

type MarshalFunc func(any) ([]byte, error)

type serializer struct {
	Marshal   MarshalFunc
	Unmarshal UnmarshalFunc
}

// Serializer 管理配置文件序列化的方法
//
// 根据配置文件的扩展查找相应的序列化方法，
// 扩展名必须以 . 开头，如果未带 .，则会自动加上。
type Serializer map[string]*serializer

// Serializer 返回管理配置文件序列化的对象
func (f *Config) Serializer() Serializer { return f.s }

// Add 添加新的序列方法
//
// ext 为文件扩展名，需要带 . 符号；
func (s Serializer) Add(m MarshalFunc, u UnmarshalFunc, ext ...string) Serializer {
	if len(ext) == 0 {
		panic("参数 ext 不能为空")
	}

	for _, e := range ext {
		e = buildExt(e)
		if s.Exists(e) {
			panic(fmt.Sprintf("已经存在同名的扩展名 %s", e))
		}
		s[e] = &serializer{Marshal: m, Unmarshal: u}
	}

	return s
}

// Exists 是否存在对指定扩展名的序列化方法
func (s Serializer) Exists(ext string) bool {
	_, found := s[buildExt(ext)]
	return found
}

// Delete 删除序列化方法
func (s Serializer) Delete(ext ...string) {
	for _, e := range ext {
		delete(s, buildExt(e))
	}
}

// Get 获取指定扩展名对应的序列化方法
//
// 如果不存在，则返回 nil。
func (s Serializer) Get(ext string) (MarshalFunc, UnmarshalFunc) {
	if ss, found := s[buildExt(ext)]; found {
		return ss.Marshal, ss.Unmarshal
	}
	return nil, nil
}

// GetByFilename 通过文件查找对应的序列化方法
func (s Serializer) GetByFilename(name string) (MarshalFunc, UnmarshalFunc) {
	return s.Get(filepath.Ext(name))
}

func (s Serializer) Len() int { return len(s) }

func buildExt(e string) string {
	if len(e) == 0 {
		panic("扩展名不能为空")
	}

	if e[0] != '.' {
		e = "." + e
	}
	return e
}

// Marshal 将 v 按 path 的后缀名序列化并保存
func (s Serializer) Marshal(path string, v any, mode fs.FileMode) error {
	if m, _ := s.GetByFilename(path); m != nil {
		data, err := m(v)
		if err != nil {
			return err
		}
		return os.WriteFile(path, data, mode)
	}

	return localeutil.Error("not found serializer for %s", path)
}

// Unmarshal 根据 path 后缀名序列化其内容至 v
func (s Serializer) Unmarshal(path string, v any) error {
	return s.unmarshal(path, v, func() ([]byte, error) { return os.ReadFile(path) })
}

// UnmarshalFS 根据 name 后缀名序列化其内容至 v
func (s Serializer) UnmarshalFS(fsys fs.FS, name string, v any) error {
	return s.unmarshal(name, v, func() ([]byte, error) { return fs.ReadFile(fsys, name) })
}

func (s Serializer) unmarshal(filename string, v any, read func() ([]byte, error)) error {
	if _, u := s.GetByFilename(filename); u != nil {
		data, err := read()
		if err != nil {
			return err
		}
		return u(data, v)
	}

	return localeutil.Error("not found serializer for %s", filename)
}
