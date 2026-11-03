// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package web

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	wings_model "novroforge.xyz/models/wings"
	"novroforge.xyz/modules/log"

	"github.com/go-chi/chi/v5"
)

const maxWingsBackupCallbackSize = 16 << 10

func WingsRemoteBackupStatus(w http.ResponseWriter, r *http.Request) {
	backupUUID := chi.URLParam(r, "backup")
	tokenID, token, ok := parseWingsRemoteAuthorization(r.Header.Get("Authorization"))
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err := wings_model.AuthenticateBackupCallback(r.Context(), backupUUID, tokenID, token); err != nil {
		if errors.Is(err, wings_model.ErrBackupCallbackUnauthorized) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		} else {
			log.Error("Authenticate Wings backup callback: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	var payload struct {
		Checksum     string `json:"checksum"`
		ChecksumType string `json:"checksum_type"`
		Size         *int64 `json:"size"`
		Successful   *bool  `json:"successful"`
		Parts        []struct {
			ETag       string `json:"etag"`
			PartNumber int    `json:"part_number"`
		} `json:"parts"`
	}
	if err := decodeWingsBackupCallback(w, r, &payload); err != nil {
		http.Error(w, "invalid callback payload", http.StatusBadRequest)
		return
	}
	if payload.Size == nil || payload.Successful == nil {
		http.Error(w, "invalid callback payload", http.StatusBadRequest)
		return
	}
	if err := wings_model.ApplyBackupStatus(r.Context(), backupUUID, *payload.Successful, payload.Checksum, payload.ChecksumType, *payload.Size); err != nil {
		if errors.Is(err, wings_model.ErrInvalidBackupStatus) {
			http.Error(w, "invalid backup status", http.StatusBadRequest)
		} else if errors.Is(err, wings_model.ErrBackupNotReady) || errors.Is(err, wings_model.ErrBackupNotFound) {
			http.Error(w, "backup status conflict", http.StatusConflict)
		} else {
			log.Error("Apply Wings backup status: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusNoContent)
}

func WingsRemoteBackupRestoreStatus(w http.ResponseWriter, r *http.Request) {
	backupUUID := chi.URLParam(r, "backup")
	tokenID, token, ok := parseWingsRemoteAuthorization(r.Header.Get("Authorization"))
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err := wings_model.AuthenticateBackupCallback(r.Context(), backupUUID, tokenID, token); err != nil {
		if errors.Is(err, wings_model.ErrBackupCallbackUnauthorized) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		} else {
			log.Error("Authenticate Wings backup restore callback: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	var payload struct {
		Successful *bool `json:"successful"`
	}
	if err := decodeWingsBackupCallback(w, r, &payload); err != nil {
		http.Error(w, "invalid callback payload", http.StatusBadRequest)
		return
	}
	if payload.Successful == nil {
		http.Error(w, "invalid callback payload", http.StatusBadRequest)
		return
	}
	if err := wings_model.SetBackupRestoreStatus(r.Context(), backupUUID, *payload.Successful); err != nil {
		if errors.Is(err, wings_model.ErrBackupNotReady) || errors.Is(err, wings_model.ErrBackupNotFound) {
			http.Error(w, "backup restore status conflict", http.StatusConflict)
		} else {
			log.Error("Set Wings backup restore status: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusNoContent)
}

// WingsRemoteServers returns the (currently empty) list of servers registered on
// this node. WebMain does not provision servers onto Wings, so the list is
// always empty; this endpoint exists so the Wings daemon can boot.
func WingsRemoteServers(w http.ResponseWriter, r *http.Request) {
	tokenID, token, ok := parseWingsRemoteAuthorization(r.Header.Get("Authorization"))
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	node, exists, err := wings_model.GetNodeByRemoteTokenID(r.Context(), tokenID)
	if err != nil {
		log.Error("Get Wings node by remote token ID: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if !exists || node == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	ok2, err := node.AuthenticateRemoteCallback(tokenID, token)
	if err != nil {
		log.Error("Authenticate Wings remote request: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if !ok2 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	resp := map[string]any{
		"data": []any{},
		"meta": map[string]any{
			"pagination": map[string]any{
				"current_page": 1,
				"from":         0,
				"last_page":    1,
				"per_page":     50,
				"to":           0,
				"total":        0,
			},
		},
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// WingsRemoteServersReset acknowledges Wings' request to reset server states.
func WingsRemoteServersReset(w http.ResponseWriter, r *http.Request) {
	tokenID, token, ok := parseWingsRemoteAuthorization(r.Header.Get("Authorization"))
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	node, exists, err := wings_model.GetNodeByRemoteTokenID(r.Context(), tokenID)
	if err != nil {
		log.Error("Get Wings node by remote token ID: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if !exists || node == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	ok2, err := node.AuthenticateRemoteCallback(tokenID, token)
	if err != nil || !ok2 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusNoContent)
}

func parseWingsRemoteAuthorization(header string) (string, string, bool) {
	scheme, credentials, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || strings.ContainsAny(credentials, " \t\r\n") {
		return "", "", false
	}
	tokenID, token, ok := strings.Cut(credentials, ".")
	if !ok || tokenID == "" || token == "" || strings.Contains(token, ".") {
		return "", "", false
	}
	return tokenID, token, true
}

func decodeWingsBackupCallback(w http.ResponseWriter, r *http.Request, payload any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxWingsBackupCallbackSize)
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(payload); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errors.New("callback payload must contain one JSON value")
	}
	return nil
}
