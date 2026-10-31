// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package service_message

import (
	"novroforge.xyz/modules/timeutil"
	"novroforge.xyz/modules/util"
)

var (
	ErrServiceMessageNotExist    = util.NewNotExistErrorf("service message does not exist")
	ErrInputWasEmpty             = util.NewNotExistErrorf("input must not be empty")
	ErrInvalidServiceMessageType = util.NewInvalidArgumentErrorf("service message type was invalid")
)

type SMType string

const (
	SMModal SMType = "modal"
)

func (s SMType) Name() string {
	res := ""
	if s == SMModal {
		res = "modal"
	}
	return res
}

func (s SMType) Valid() bool {
	return SMType(s.Name()) == SMModal
}

// Holds: ServiceMessageType:[service_message_confirmed]
type ConfirmTimestamps map[SMType]timeutil.TimeStamp

type ServiceMessageOptions struct {
	Title string
	Text  string
	Type  string
}
