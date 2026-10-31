// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package setting

var (
	// NovroForge is English-only. Keep the i18n helper for code stability.
	Langs = []string{"en-US"}
	Names = []string{"English"}
)

func loadI18nFrom(_ ConfigProvider) {}
