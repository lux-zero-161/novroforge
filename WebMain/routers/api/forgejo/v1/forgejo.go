// SPDX-License-Identifier: MIT

package v1

import (
	"net/http"

	"novroforge.xyz/modules/json"
	"novroforge.xyz/modules/setting"
)

type NovroForge struct{}

var _ ServerInterface = &NovroForge{}

func NewNovroForge() *NovroForge {
	return &NovroForge{}
}

func (f *NovroForge) GetVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(Version{&setting.NovroForgeVersion})
}
