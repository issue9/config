// SPDX-License-Identifier: MIT

// Package locales 提供本地化的数据
package locales

import "embed"

//go:embed *.yaml
var Locales embed.FS
