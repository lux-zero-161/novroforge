// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package web

import (
	"net/http"
	"strings"

	"forgejo.org/modules/base"
	"forgejo.org/services/context"
)

func legalTemplate(ctx *context.Context, name string) base.TplName {
	if strings.HasPrefix(ctx.Locale.Language(), "de") {
		return base.TplName("legal/" + name)
	}
	return base.TplName("legal/" + name + "_en")
}

func LegalPrivacy(ctx *context.Context) {
	ctx.Data["Title"] = "Privacy Policy"
	if strings.HasPrefix(ctx.Locale.Language(), "de") {
		ctx.Data["Title"] = "Datenschutzerklärung"
	}
	ctx.Data["PageIsLegal"] = true
	ctx.HTML(http.StatusOK, legalTemplate(ctx, "privacy"))
}

func LegalTOS(ctx *context.Context) {
	ctx.Data["Title"] = "Terms of Service"
	if strings.HasPrefix(ctx.Locale.Language(), "de") {
		ctx.Data["Title"] = "Nutzungsbedingungen"
	}
	ctx.Data["PageIsLegal"] = true
	ctx.HTML(http.StatusOK, legalTemplate(ctx, "tos"))
}

func LegalImprint(ctx *context.Context) {
	ctx.Data["Title"] = "Imprint"
	if strings.HasPrefix(ctx.Locale.Language(), "de") {
		ctx.Data["Title"] = "Impressum"
	}
	ctx.Data["PageIsLegal"] = true
	ctx.HTML(http.StatusOK, legalTemplate(ctx, "imprint"))
}
