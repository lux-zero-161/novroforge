// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	wings_model "novroforge.xyz/models/wings"
	"novroforge.xyz/modules/base"
	"novroforge.xyz/modules/log"
	"novroforge.xyz/modules/setting"
	"novroforge.xyz/services/context"
)

const tplWingsNests base.TplName = "admin/wings_nests"

func WingsNests(ctx *context.Context) {
	nests, err := wings_model.ListNests(ctx)
	if err != nil {
		ctx.ServerError("ListWingsNests", err)
		return
	}
	eggs := make([]*wings_model.WingsEgg, 0)
	variables := make(map[int64][]*wings_model.WingsEggVariable)
	for _, nest := range nests {
		nestEggs, err := wings_model.ListEggs(ctx, nest.ID)
		if err != nil {
			ctx.ServerError("ListWingsEggs", err)
			return
		}
		eggs = append(eggs, nestEggs...)
		for _, egg := range nestEggs {
			vs, err := wings_model.ListEggVariables(ctx, egg.ID)
			if err != nil {
				ctx.ServerError("ListWingsEggVariables", err)
				return
			}
			variables[egg.ID] = vs
		}
	}
	ctx.Data["Title"] = ctx.Tr("admin.wings_nests")
	ctx.Data["PageIsAdminWingsNests"] = true
	ctx.Data["Nests"] = nests
	ctx.Data["Eggs"] = eggs
	ctx.Data["Variables"] = variables
	ctx.HTML(http.StatusOK, tplWingsNests)
}

func CreateWingsNest(ctx *context.Context) {
	name := ctx.FormString("name")
	description := ctx.FormString("description")
	if _, err := wings_model.CreateNest(ctx, name, description); err != nil {
		log.Error("CreateWingsNest: %v", err)
		ctx.Flash.Error(ctx.Tr("admin.wings_nests.create_failed"))
	} else {
		ctx.Flash.Success(ctx.Tr("admin.wings_nests.created"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nests")
}

func DeleteWingsNest(ctx *context.Context) {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.invalid_id"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nests")
		return
	}
	if err := wings_model.DeleteNest(ctx, id); err != nil {
		log.Error("DeleteWingsNest[%d]: %v", id, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_nests.delete_failed"))
	} else {
		ctx.Flash.Success(ctx.Tr("admin.wings_nests.deleted"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nests")
}

func CreateWingsEgg(ctx *context.Context) {
	nestID, err := strconv.ParseInt(ctx.FormString("nest_id"), 10, 64)
	if err != nil || nestID <= 0 {
		ctx.Flash.Error(ctx.Tr("admin.wings_nests.invalid_nest"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nests")
		return
	}
	_, err = wings_model.CreateEgg(ctx, nestID, ctx.FormString("name"), ctx.FormString("description"),
		ctx.FormString("docker_image"), ctx.FormString("startup"), wings_model.EggConfig{})
	if err != nil {
		log.Error("CreateWingsEgg: %v", err)
		ctx.Flash.Error(ctx.Tr("admin.wings_nests.egg_create_failed"))
	} else {
		ctx.Flash.Success(ctx.Tr("admin.wings_nests.egg_created"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nests")
}

func DeleteWingsEgg(ctx *context.Context) {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.invalid_id"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nests")
		return
	}
	if err := wings_model.DeleteEgg(ctx, id); err != nil {
		log.Error("DeleteWingsEgg[%d]: %v", id, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_nests.egg_delete_failed"))
	} else {
		ctx.Flash.Success(ctx.Tr("admin.wings_nests.egg_deleted"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nests")
}

func CreateWingsEggVariable(ctx *context.Context) {
	eggID, err := strconv.ParseInt(ctx.FormString("egg_id"), 10, 64)
	if err != nil || eggID <= 0 {
		ctx.Flash.Error(ctx.Tr("admin.wings_nests.invalid_egg"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nests")
		return
	}
	userViewable := ctx.FormString("user_viewable") == "true"
	userEditable := ctx.FormString("user_editable") == "true"
	if _, err := wings_model.CreateEggVariable(ctx, eggID, ctx.FormString("name"), ctx.FormString("description"),
		ctx.FormString("env_variable"), ctx.FormString("default_value"), userViewable, userEditable, ctx.FormString("rules")); err != nil {
		log.Error("CreateWingsEggVariable: %v", err)
		ctx.Flash.Error(ctx.Tr("admin.wings_nests.variable_create_failed"))
	} else {
		ctx.Flash.Success(ctx.Tr("admin.wings_nests.variable_created"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nests")
}

func DeleteWingsEggVariable(ctx *context.Context) {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.invalid_id"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nests")
		return
	}
	if err := wings_model.DeleteEggVariable(ctx, id); err != nil {
		log.Error("DeleteWingsEggVariable[%d]: %v", id, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_nests.variable_delete_failed"))
	} else {
		ctx.Flash.Success(ctx.Tr("admin.wings_nests.variable_deleted"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nests")
}

func ImportWingsEgg(ctx *context.Context) {
	nestID, err := strconv.ParseInt(ctx.FormString("nest_id"), 10, 64)
	if err != nil || nestID <= 0 {
		ctx.Flash.Error(ctx.Tr("admin.wings_nests.invalid_nest"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nests")
		return
	}
	raw := strings.TrimSpace(ctx.FormString("egg_json"))
	if raw == "" {
		ctx.Flash.Error(ctx.Tr("admin.wings_nests.import_empty"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nests")
		return
	}
	var payload struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		DockerImage string `json:"docker_image"`
		DockerImages map[string]string `json:"docker_images"`
		Startup     string `json:"startup"`
		Config      any    `json:"config"`
		Variables   []struct {
			Name         string `json:"name"`
			Description  string `json:"description"`
			EnvVariable  string `json:"env_variable"`
			DefaultValue string `json:"default_value"`
			UserViewable bool   `json:"user_viewable"`
			UserEditable bool   `json:"user_editable"`
			Rules        string `json:"rules"`
		} `json:"variables"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		log.Error("ImportWingsEgg: %v", err)
		ctx.Flash.Error(ctx.Tr("admin.wings_nests.import_invalid"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nests")
		return
	}
	if payload.Name == "" {
		payload.Name = "Imported egg"
	}
	if payload.DockerImage == "" && len(payload.DockerImages) > 0 {
		for _, v := range payload.DockerImages {
			payload.DockerImage = v
			break
		}
	}
	cfgBytes, _ := json.Marshal(payload.Config)
	var cfg wings_model.EggConfig
	_ = json.Unmarshal(cfgBytes, &cfg)
	egg, err := wings_model.CreateEgg(ctx, nestID, payload.Name, payload.Description, payload.DockerImage, payload.Startup, cfg)
	if err != nil {
		log.Error("ImportWingsEgg create: %v", err)
		ctx.Flash.Error(ctx.Tr("admin.wings_nests.egg_create_failed"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nests")
		return
	}
	for _, v := range payload.Variables {
		if v.EnvVariable == "" {
			continue
		}
		userViewable := v.UserViewable
		userEditable := v.UserEditable
		// Pterodactyl defaults: viewable/editable true if not specified? keep as parsed
		if _, err := wings_model.CreateEggVariable(ctx, egg.ID, v.Name, v.Description, v.EnvVariable, v.DefaultValue, userViewable, userEditable, v.Rules); err != nil {
			log.Error("ImportWingsEgg variable: %v", err)
		}
	}
	ctx.Flash.Success(ctx.Tr("admin.wings_nests.imported"))
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nests")
}
