// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package forms

import (
	"net/http"

	"novroforge.xyz/modules/web/middleware"
	"novroforge.xyz/services/context"

	"code.novroforge.xyz/go-chi/binding"
)

type ServiceMessageForm struct {
	Title string `binding:"Required"`
	Text  string `binding:"Required"`
}

// Validate validates the fields
func (f *ServiceMessageForm) Validate(req *http.Request, errs binding.Errors) binding.Errors {
	ctx := context.GetValidateContext(req)
	return middleware.Validate(errs, ctx.Data, f, ctx.Locale)
}
