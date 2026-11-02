// Copyright 2025 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package templates

import (
	"context"

	"novroforge.xyz/modules/translation"
)

type Context struct {
	context.Context
	Locale      translation.Locale
	AvatarUtils *AvatarUtils
	Data        map[string]any
}

var _ context.Context = Context{}

func NewContext(ctx context.Context) *Context {
	return &Context{Context: ctx}
}
