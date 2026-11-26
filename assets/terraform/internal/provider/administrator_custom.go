package provider

import (
	"context"

	"github.com/PaloAltoNetworks/pango/device/administrator"
	"github.com/PaloAltoNetworks/pango/util"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

type AdministratorCustom struct{}

func NewAdministratorCustom(_ *ProviderData) (*AdministratorCustom, error) {
	return &AdministratorCustom{}, nil
}

func (o *AdministratorResource) ValidateConfigCustom(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config AdministratorResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !config.PasswordWo.IsNull() && !config.Password.IsNull() {
		resp.Diagnostics.AddError("Conflicting passwords", "Configure either password or password_wo, not both.")
	}
	if config.PasswordWo.IsNull() != config.PasswordWoVersion.IsNull() {
		resp.Diagnostics.AddError("Incomplete write-only password", "password_wo and password_wo_version must be configured together.")
	}
	if !config.PasswordWoVersion.IsNull() && !config.PasswordWoVersion.IsUnknown() && config.PasswordWoVersion.ValueInt64() <= 0 {
		resp.Diagnostics.AddError("Invalid rotation version", "password_wo_version must be positive.")
	}
}

func (o *AdministratorResource) PreCreate(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse, state *AdministratorResourceModel, _ administrator.Location, obj *administrator.Entry, _ *EncryptedValuesManager) {
	var config AdministratorResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || config.PasswordWoVersion.IsNull() {
		return
	}
	if config.PasswordWo.IsUnknown() || config.PasswordWo.IsNull() || config.PasswordWo.ValueString() == "" {
		resp.Diagnostics.AddError("Missing password", "password_wo must be a known non-empty value at apply time.")
		return
	}
	hash, err := o.client.RequestPasswordHash(ctx, config.PasswordWo.ValueString())
	if err != nil || hash == "" {
		resp.Diagnostics.AddError("Password hashing failed", "Panorama did not return an administrator password hash.")
		return
	}
	obj.Phash = &hash
}

func (o *AdministratorResource) PreUpdate(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse, plan, state *AdministratorResourceModel, _ administrator.Location, obj *administrator.Entry, _ *EncryptedValuesManager) {
	if plan.PasswordWoVersion.IsNull() || plan.PasswordWoVersion.Equal(state.PasswordWoVersion) {
		return
	}
	var config AdministratorResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if config.PasswordWo.IsUnknown() || config.PasswordWo.IsNull() || config.PasswordWo.ValueString() == "" {
		resp.Diagnostics.AddError("Missing password", "password_wo must be a known non-empty value at apply time.")
		return
	}
	hash, err := o.client.RequestPasswordHash(ctx, config.PasswordWo.ValueString())
	if err != nil || hash == "" {
		resp.Diagnostics.AddError("Password hashing failed", "Panorama did not return an administrator password hash.")
		return
	}
	obj.Phash = &hash
}

func administratorCreatePasswordHash(ctx context.Context, client util.PangoClient, password string) (string, error) {
	return client.RequestPasswordHash(ctx, password)
}
