// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package issue

import (
	"context"

	issues_model "novroforge.xyz/models/issues"
	user_model "novroforge.xyz/models/user"
	notify_service "novroforge.xyz/services/notify"
)

// ChangeContent changes issue content, as the given user.
func ChangeContent(ctx context.Context, issue *issues_model.Issue, doer *user_model.User, content string, contentVersion int) (err error) {
	oldContent := issue.Content

	if err := issues_model.ChangeIssueContent(ctx, issue, doer, content, contentVersion); err != nil {
		return err
	}

	notify_service.IssueChangeContent(ctx, doer, issue, oldContent)

	return nil
}
