// Copyright 2024 The NovroForge Authors c/o Codeberg e.V.. All rights reserved.
// SPDX-License-Identifier: MIT

package structs

// ReplaceFlagsOption options when replacing the flags of a repository
type ReplaceFlagsOption struct {
	Flags []string `json:"flags"`
}
