// SPDX-FileCopyrightText: 2019-2024 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/issue9/assert/v4"
)

func TestSerializer(t *testing.T) {
	a := assert.New(t, false)
	s := Serializer{}

	a.Equal(0, s.Len())

	t.Run("Add", func(t *testing.T) {
		a := assert.New(t, false)

		s.Add(json.Marshal, json.Unmarshal, ".json", "js")
		a.True(s.Exists("json")).
			True(s.Exists(".js")).
			Equal(2, s.Len())

		a.PanicString(func() {
			s.Add(xml.Marshal, xml.Unmarshal)
		}, "参数 ext 不能为空")

		a.PanicString(func() {
			s.Add(xml.Marshal, xml.Unmarshal, "")
		}, "扩展名不能为空")

		a.PanicString(func() {
			s.Add(xml.Marshal, xml.Unmarshal, "json")
		}, "已经存在同名的扩展名 .json")
	})

	t.Run("Delete", func(t *testing.T) {
		a := assert.New(t, false)

		s.Delete(".not-exists")
		a.Equal(2, s.Len())

		s.Delete(".json", ".mjs")
		a.Equal(1, s.Len())
		a.NotNil(s.GetByFilename("abc.js")).
			Nil(s.GetByFilename(".json"))
	})
}
