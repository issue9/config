// SPDX-FileCopyrightText: 2019-2025 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/issue9/assert/v4"
)

type conf struct {
	XMLName struct{} `json:"-" xml:"config"`
	Debug   bool     `json:"debug" xml:"debug,attr"`
	Port    int      `json:"port" xml:"port,attr"`
	Cert    string   `json:"cert" xml:"cert"`
}

func TestAppDir(t *testing.T) {
	a := assert.New(t, false)
	c, err := AppDir(nil, "config")
	a.NotError(err).NotNil(c).
		NotNil(c.Root())
}

func TestConfig(t *testing.T) {
	a := assert.New(t, false)

	c := Dir(nil, "./testdata")
	a.NotNil(c)
	a.True(c.Exists("config.xml")).False(c.Exists("not-exists.xml"))

	data, err := c.Read("config.xml")
	a.NotError(err).NotNil(data)

	obj := &conf{}
	a.ErrorString(c.Load("config.json", obj), "not found serializer for ")

	c.Serializer().Add(xml.Marshal, xml.Unmarshal, ".xml").
		Add(json.Marshal, json.Unmarshal, ".json", ".js")

	obj1 := &conf{}
	a.NotError(c.Load("config.json", obj))
	obj2 := &conf{}
	a.NotError(c.Load("config.xml", obj))
	a.Equal(obj1, obj2)
}
