// SPDX-License-Identifier: MIT

package config

import "fmt"

// Error 配置文件错误信息
type Error struct {
	File    string // 文件地址
	Field   string // 字段名称
	Message string // 错误信息
}

// Sanitizer 对配置项的作检测和清理
//
// 如果对象实现了该方法，那么在解析完之后，
// 会调用该接口的函数对数据进行修正和检测。
type Sanitizer interface {
	// 对对象的各个字段进行检测，如果可以调整，则调整。
	// 如果不能调整，则返回错误信息。返回的错误信息，
	// 尽可能是 *Error 对象。
	Sanitize() error
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
