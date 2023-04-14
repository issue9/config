// SPDX-License-Identifier: MIT

package config

import (
	"errors"
	"fmt"
	"path/filepath"
)

var errSerializerNotFound = errors.New("serializer not found")

type UnmarshalFunc func([]byte, interface{}) error

type MarshalFunc func(interface{}) ([]byte, error)

type serializer struct {
	Marshal   MarshalFunc
	Unmarshal UnmarshalFunc
}

// Serializer 管理配置文件序列化的方法
//
// 根据配置文件的扩展查找相应的序列化方法，
// 扩展名必须以 . 开头，如果未带 .，则会自动加上。
type Serializer map[string]*serializer

func ErrSerializerNotFound() error { return errSerializerNotFound }

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

func (s Serializer) Len() int { return len(s) }

func (s Serializer) searchByExt(name string) *serializer {
	return s[filepath.Ext(name)]
}

func buildExt(e string) string {
	if len(e) == 0 {
		panic("扩展名不能为空")
	}

	if e[0] != '.' {
		e = "." + e
	}
	return e
}
