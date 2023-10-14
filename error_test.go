// SPDX-License-Identifier: MIT

package config

import (
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

var (
	_ error               = &FieldError{}
	_ localeutil.Stringer = &FieldError{}
)

func TestNewFieldError(t *testing.T) {
	a := assert.New(t, false)

	err1 := NewFieldError("f1", "err1")
	a.NotNil(err1)

	err2 := NewFieldError("f2", err1)
	a.NotNil(err2).
		Equal(err2.Field, "f2.f1").
		Equal(err1.Field, "f2.f1")
}

func TestFieldError_LocaleString(t *testing.T) {
	a := assert.New(t, false)
	hans := language.MustParse("cmn-hans")
	hant := language.MustParse("cmn-hant")

	b := catalog.NewBuilder()
	b.SetString(hans, "%s at %s:%s", "位于 %[2]s:%[3]s 发生了 %[1]s")
	b.SetString(hant, "%s at %s:%s", "位于 %[2]s:%[3]s 发生了 %[1]s")

	a.NotError(b.SetString(hans, "k1", "cn1"))
	a.NotError(b.SetString(hant, "k1", "tw1"))

	cnp := message.NewPrinter(hans, message.Catalog(b))
	twp := message.NewPrinter(hant, message.Catalog(b))

	ferr := NewFieldError("", localeutil.Phrase("k1"))
	ferr.Path = "path"
	a.Equal("位于 path: 发生了 cn1", ferr.LocaleString(cnp))
	a.Equal("位于 path: 发生了 tw1", ferr.LocaleString(twp))
	a.Equal("k1 at path:", ferr.Error())
}

func TestFieldError_SetFieldParent(t *testing.T) {
	a := assert.New(t, false)

	err := NewFieldError("f1", "error")
	err.AddFieldParent("f2")
	a.Equal(err.Field, "f2.f1")
	err.AddFieldParent("f3")
	a.Equal(err.Field, "f3.f2.f1")
	err.AddFieldParent("")
	a.Equal(err.Field, "f3.f2.f1")

	err = NewFieldError("", "error")
	err.AddFieldParent("f2")
	a.Equal(err.Field, "f2")
	err.AddFieldParent("f3")
	a.Equal(err.Field, "f3.f2")
}
