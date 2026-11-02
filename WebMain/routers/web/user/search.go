// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package user

import (
	"net/http"

	"novroforge.xyz/models/db"
	user_model "novroforge.xyz/models/user"
	"novroforge.xyz/modules/optional"
	"novroforge.xyz/modules/setting"
	"novroforge.xyz/services/context"
	"novroforge.xyz/services/convert"
)

// SearchCandidates searches candidate users for dropdown list
func SearchCandidates(ctx *context.Context) {
	users, _, err := user_model.SearchUsers(ctx, &user_model.SearchUserOptions{
		Actor:       ctx.Doer,
		Keyword:     ctx.FormTrim("q"),
		Type:        user_model.UserTypeIndividual,
		IsActive:    optional.Some(true),
		ListOptions: db.ListOptions{PageSize: setting.UI.MembersPagingNum},
	})
	if err != nil {
		ctx.ServerError("Unable to search users", err)
		return
	}
	ctx.JSON(http.StatusOK, map[string]any{"data": convert.ToUsers(ctx, ctx.Doer, users)})
}
