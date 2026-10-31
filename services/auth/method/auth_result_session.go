// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package method

import (
	user_model "novroforge.xyz/models/user"
	"novroforge.xyz/services/auth"
)

var _ auth.AuthenticationResult = &sessionAuthenticationResult{}

type sessionAuthenticationResult struct {
	*auth.BaseAuthenticationResult
	user *user_model.User
}

func (r *sessionAuthenticationResult) User() *user_model.User {
	return r.user
}
