// Copyright 2023 The NovroForge Authors
// SPDX-License-Identifier: MIT

package migrations

import (
	"novroforge.xyz/modules/structs"
)

func init() {
	RegisterDownloaderFactory(&NovroForgeDownloaderFactory{})
}

type NovroForgeDownloaderFactory struct {
	GiteaDownloaderFactory
}

func (f *NovroForgeDownloaderFactory) GitServiceType() structs.GitServiceType {
	return structs.NovroForgeService
}
