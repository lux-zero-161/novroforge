// Copyright 2026 The Forgejo Authors
// SPDX-License-Identifier: MIT

package testimport

// ensure the init() function of those modules are called in a test
// environment that may not include them. It matters when the engine
// is trying to figure out the ordering of foreign keys, for instance

import ( //revive:disable:blank-imports
	_ "novroforge.xyz/models/actions"
	_ "novroforge.xyz/models/activities"
	_ "novroforge.xyz/models/auth"
	_ "novroforge.xyz/models/forgefed"
	_ "novroforge.xyz/models/perm/access"
	_ "novroforge.xyz/models/repo"
)
