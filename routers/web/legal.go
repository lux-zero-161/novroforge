// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package web

import (
	"net/http"

	"forgejo.org/modules/base"
	"forgejo.org/services/context"
)

const (
	tplLegalPrivacy base.TplName = "legal/privacy"
	tplLegalTOS     base.TplName = "legal/tos"
	tplLegalImprint base.TplName = "legal/imprint"
)

func LegalPrivacy(ctx *context.Context) {
	ctx.Data["Title"] = "Datenschutzerklärung"
	ctx.Data["PageIsLegal"] = true
	ctx.HTML(http.StatusOK, tplLegalPrivacy)
}

func LegalTOS(ctx *context.Context) {
	ctx.Data["Title"] = "Nutzungsbedingungen"
	ctx.Data["PageIsLegal"] = true
	ctx.HTML(http.StatusOK, tplLegalTOS)
}

func LegalImprint(ctx *context.Context) {
	ctx.Data["Title"] = "Impressum"
	ctx.Data["PageIsLegal"] = true
	ctx.HTML(http.StatusOK, tplLegalImprint)
}
