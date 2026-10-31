// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package actions

import (
	"novroforge.xyz/modules/graceful"
	"novroforge.xyz/modules/log"
	"novroforge.xyz/modules/queue"
	"novroforge.xyz/modules/setting"
	notify_service "novroforge.xyz/services/notify"
)

func Init() {
	if !setting.Actions.Enabled {
		return
	}

	jobEmitterQueue = queue.CreateUniqueQueue(graceful.GetManager().ShutdownContext(), "actions_ready_job", jobEmitterQueueHandler)
	if jobEmitterQueue == nil {
		log.Fatal("Unable to create actions_ready_job queue")
	}
	go graceful.GetManager().RunWithCancel(jobEmitterQueue)

	notify_service.RegisterNotifier(NewNotifier())
}
