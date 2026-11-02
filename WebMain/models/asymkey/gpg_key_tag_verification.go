// Copyright 2024 The NovroForge Authors c/o Codeberg e.V.. All rights reserved.
// SPDX-License-Identifier: MIT

package asymkey

import (
	"context"

	"novroforge.xyz/modules/git"
)

func ParseTagWithSignature(ctx context.Context, gitRepo *git.Repository, t *git.Tag) *ObjectVerification {
	o := tagToGitObject(t, gitRepo)
	return ParseObjectWithSignature(ctx, &o)
}
