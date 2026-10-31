// Copyright 2023 The NovroForge Authors
// SPDX-License-Identifier: MIT

package migrations

import (
	"testing"

	"novroforge.xyz/modules/structs"

	"github.com/stretchr/testify/require"
)

func TestNovroForgeDownload(t *testing.T) {
	require.NotNil(t, getFactoryFromServiceType(structs.NovroForgeService))
}
