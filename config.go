// SPDX-License-Identifier: MIT

// Package config 提供了对多种格式配置文件的支持
package config

import "fmt"

// UnmarshalFunc 定义了将文本内容解析到对象的函数原型。
type UnmarshalFunc func([]byte, interface{}) error

// Error 配置文件错误信息
type Error struct {
	File    string // 文件地址
	Field   string // 字段名称
	Message string // 错误信息
}

// NewError 声明新的 Error 对象
func NewError(file, field, message string) *Error {
	return &Error{
		File:    file,
		Field:   field,
		Message: message,
	}
}

func (err *Error) Error() string {
	return fmt.Sprintf("在 %s 中的 %s 出错了 %s", err.File, err.Field, err.Message)
}
