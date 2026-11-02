// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	wings_model "novroforge.xyz/models/wings"
	"novroforge.xyz/modules/log"
	"novroforge.xyz/modules/setting"
	"novroforge.xyz/services/context"
)

const maxWingsConsoleConnectionsPerUser = 3

var wingsConsoleConnections = struct {
	sync.Mutex
	active map[string]int
}{active: make(map[string]int)}

var wingsConsoleUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	CheckOrigin: func(request *http.Request) bool {
		return isWingsConsoleOriginAllowed(request.Header.Get("Origin"))
	},
}

func WingsServerConsoleWebSocket(ctx *context.Context) {
	if !wingsConsoleUpgrader.CheckOrigin(ctx.Req) {
		ctx.Error(http.StatusForbidden, ctx.Locale.TrString("wings.console_origin_denied"))
		return
	}
	origin, err := wingsConsolePanelOrigin()
	if err != nil {
		ctx.ServerError("GetUserWingsServerConsoleWebSocketOrigin", err)
		return
	}

	server, _, err := getAuthorizedWingsServer(ctx)
	if errors.Is(err, wings_model.ErrServerAccessDenied) {
		ctx.NotFound("WingsServerConsoleWebSocket", nil)
		return
	}
	if err != nil {
		ctx.ServerError("GetUserWingsServerConsoleWebSocket", err)
		return
	}
	releaseConnection, acquired := acquireWingsConsoleConnection(server.ID, ctx.Doer.ID)
	if !acquired {
		ctx.Error(http.StatusTooManyRequests, ctx.Locale.TrString("wings.console_connection_limit"))
		return
	}
	defer releaseConnection()

	client, err := server.GetServerClient(ctx)
	if err != nil {
		ctx.ServerError("GetUserWingsServerConsoleWebSocketClient", err)
		return
	}
	userUUID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("novroforge-user:"+strconv.FormatInt(ctx.Doer.ID, 10))).String()
	nodeConnection, err := client.OpenConsoleWebSocket(ctx, server.UUID, userUUID, origin)
	if err != nil {
		log.Error("GetUserWingsServerConsoleWebSocket[%d]: %v", server.ID, err)
		ctx.ServerError("GetUserWingsServerConsoleWebSocket", err)
		return
	}
	defer nodeConnection.Close()

	browserConnection, err := wingsConsoleUpgrader.Upgrade(ctx.Resp, ctx.Req, nil)
	if err != nil {
		return
	}
	defer browserConnection.Close()
	browserConnection.SetReadLimit(1024)

	go func() {
		_, _, err := browserConnection.ReadMessage()
		if err == nil {
			_ = browserConnection.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "console stream is read-only"), time.Now().Add(10*time.Second))
		}
		_ = nodeConnection.Close()
	}()

	for {
		messageType, message, err := nodeConnection.ReadMessage()
		if err != nil {
			return
		}
		if messageType != websocket.TextMessage {
			continue
		}
		var event struct {
			Event string `json:"event"`
		}
		if err := json.Unmarshal(message, &event); err != nil || event.Event != "console output" {
			continue
		}
		_ = browserConnection.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := browserConnection.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}

func acquireWingsConsoleConnection(serverID, userID int64) (func(), bool) {
	key := strconv.FormatInt(serverID, 10) + ":" + strconv.FormatInt(userID, 10)
	wingsConsoleConnections.Lock()
	if wingsConsoleConnections.active[key] >= maxWingsConsoleConnectionsPerUser {
		wingsConsoleConnections.Unlock()
		return nil, false
	}
	wingsConsoleConnections.active[key]++
	wingsConsoleConnections.Unlock()

	var releaseOnce sync.Once
	return func() {
		releaseOnce.Do(func() {
			wingsConsoleConnections.Lock()
			if wingsConsoleConnections.active[key] <= 1 {
				delete(wingsConsoleConnections.active, key)
			} else {
				wingsConsoleConnections.active[key]--
			}
			wingsConsoleConnections.Unlock()
		})
	}, true
}

func wingsConsolePanelOrigin() (string, error) {
	panelURL, err := url.Parse(setting.AppURL)
	if err != nil || panelURL.User != nil || panelURL.Host == "" || (panelURL.Scheme != "http" && panelURL.Scheme != "https") {
		return "", errors.New("invalid panel URL for Wings console origin")
	}
	return (&url.URL{Scheme: panelURL.Scheme, Host: panelURL.Host}).String(), nil
}

func isWingsConsoleOriginAllowed(origin string) bool {
	expectedOrigin, err := wingsConsolePanelOrigin()
	if err != nil || origin == "" {
		return false
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.User != nil || parsed.Host == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
		return false
	}
	expected, err := url.Parse(expectedOrigin)
	return err == nil && strings.EqualFold(parsed.Scheme, expected.Scheme) && strings.EqualFold(parsed.Host, expected.Host)
}
