// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package explore

import (
	"net/http"

	"novroforge.xyz/models/db"
	repo_model "novroforge.xyz/models/repo"
	api "novroforge.xyz/modules/structs"
	"novroforge.xyz/services/context"
	"novroforge.xyz/services/convert"
)

// TopicSearch search for creating topic
func TopicSearch(ctx *context.Context) {
	opts := &repo_model.FindTopicOptions{
		Keyword: ctx.FormString("q"),
		ListOptions: db.ListOptions{
			Page:     ctx.FormInt("page"),
			PageSize: convert.ToCorrectPageSize(ctx.FormInt("limit")),
		},
	}

	topics, total, err := repo_model.FindTopics(ctx, opts)
	if err != nil {
		ctx.Error(http.StatusInternalServerError)
		return
	}

	topicResponses := make([]*api.TopicResponse, len(topics))
	for i, topic := range topics {
		topicResponses[i] = convert.ToTopicResponse(topic)
	}

	ctx.SetTotalCountHeader(total)
	ctx.JSON(http.StatusOK, map[string]any{
		"topics": topicResponses,
	})
}
