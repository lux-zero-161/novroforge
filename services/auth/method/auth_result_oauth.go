// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package method

import (
	auth_model "novroforge.xyz/models/auth"
	user_model "novroforge.xyz/models/user"
	"novroforge.xyz/modules/optional"
	"novroforge.xyz/services/auth"
)

var _ auth.AuthenticationResult = &oAuth2JWTAuthenticationResult{}

type oAuth2JWTAuthenticationResult struct {
	*auth.BaseAuthenticationResult
	user        *user_model.User
	scope       optional.Option[auth_model.AccessTokenScope]
	grantScopes string
}

func (*oAuth2JWTAuthenticationResult) IsOAuth2JWTAuthentication() bool {
	return true
}

func (r *oAuth2JWTAuthenticationResult) User() *user_model.User {
	return r.user
}

func (r *oAuth2JWTAuthenticationResult) Scope() optional.Option[auth_model.AccessTokenScope] {
	return r.scope
}

func (r *oAuth2JWTAuthenticationResult) OAuth2GrantScopes() optional.Option[string] {
	return optional.Some(r.grantScopes)
}
