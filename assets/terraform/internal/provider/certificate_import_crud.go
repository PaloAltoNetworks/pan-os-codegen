package provider

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/PaloAltoNetworks/pango/device/certificate"
	sdkerrors "github.com/PaloAltoNetworks/pango/errors"
	"github.com/PaloAltoNetworks/pango/xmlapi"
	sdkmanager "github.com/PaloAltoNetworks/terraform-provider-panos/internal/manager"
)

var (
	certificateExportErrorPrivateKey = errors.New("unable to export certificate private key")
	certificateExportErrorOther      = errors.New("unexpected server error")
)

func (o *CertificateImportResource) importCertificate(ctx context.Context, state *CertificateImportResourceModel, template string, vsys string) error {
	command := &xmlapi.Import{}
	command.Extras = url.Values{}

	if template != "" {
		command.Extras.Set("target-tpl", template)
	}

	if vsys != "" && vsys != "shared" {
		command.Extras.Set("target-tpl-vsys", vsys)
	}

	command.Extras.Set("certificate-name", state.Name.ValueString())

	if state.Local != nil {
		local := state.Local
		if local.Pem != nil {
			command.Extras.Add("format", "pem")

			command.Category = "certificate"
			certificate := local.Pem.Certificate.ValueString()
			certificate_fname := local.Pem.CertificateFilename.ValueString()

			_, _, err := o.client.ImportFile(ctx, command, certificate, certificate_fname, "file", false, nil)
			if err != nil {
				return fmt.Errorf("Failed to import PEM certificate into PAN-OS: %w", err)
			}

			command.Category = "private-key"
			privateKey := local.Pem.PrivateKey.ValueStringPointer()
			if privateKey != nil {
				passphrase := local.Pem.Passphrase.ValueString()
				if passphrase == "" {
					passphrase = "dummy-passphrase"
				}
				privatekey_fname := local.Pem.PrivateKeyFilename.ValueString()

				command.Extras.Add("passphrase", passphrase)
				_, _, err := o.client.ImportFile(ctx, command, *privateKey, privatekey_fname, "file", false, nil)
				if err != nil {
					return fmt.Errorf("Failed to import PEM private key into PAN-OS: %w", err)
				}
			}
		} else if local.Pkcs12 != nil {
			command.Extras.Add("format", "pkcs12")

			command.Category = "certificate"
			certificate := local.Pkcs12.Certificate.ValueString()

			passphrase := local.Pkcs12.Passphrase.ValueString()
			if passphrase != "" {
				command.Extras.Add("passphrase", "passphrase")
			}

			_, _, err := o.client.ImportFile(ctx, command, certificate, "cert.pkcs12", "file", false, nil)
			if err != nil {
				return fmt.Errorf("Failed to import PKCS12 certificate into PAN-OS: %w", err)
			}
		}
	}

	return nil
}

func (o *CertificateImportResource) terraformToPangoLocation(ctx context.Context, source CertificateImportLocation) (*certificate.Location, diag.Diagnostics) {
	var location certificate.Location

	var diags diag.Diagnostics

	{
		if !source.Panorama.IsNull() {
			location.Panorama = &certificate.PanoramaLocation{}
			var innerLocation CertificateImportPanoramaLocation
			diags.Append(source.Panorama.As(ctx, &innerLocation, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return nil, diags
			}
		}

		if !source.Vsys.IsNull() {
			location.Vsys = &certificate.VsysLocation{}
			var innerLocation CertificateImportVsysLocation
			diags.Append(source.Vsys.As(ctx, &innerLocation, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return nil, diags
			}
			location.Vsys.NgfwDevice = innerLocation.NgfwDevice.ValueString()
			location.Vsys.Vsys = innerLocation.Name.ValueString()
		}

		if !source.Template.IsNull() {
			location.Template = &certificate.TemplateLocation{}
			var innerLocation CertificateImportTemplateLocation
			diags.Append(source.Template.As(ctx, &innerLocation, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return nil, diags
			}

			location.Template.PanoramaDevice = innerLocation.PanoramaDevice.ValueString()
			location.Template.Template = innerLocation.Name.ValueString()
		}

		if !source.TemplateVsys.IsNull() {
			location.TemplateVsys = &certificate.TemplateVsysLocation{}
			var innerLocation CertificateImportTemplateVsysLocation
			diags.Append(source.Template.As(ctx, &innerLocation, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return nil, diags
			}
			location.TemplateVsys.NgfwDevice = innerLocation.NgfwDevice.ValueString()
			location.TemplateVsys.PanoramaDevice = innerLocation.PanoramaDevice.ValueString()
			location.TemplateVsys.Template = innerLocation.Template.ValueString()
			location.TemplateVsys.Vsys = innerLocation.Vsys.ValueString()
		}

		if !source.TemplateStack.IsNull() {
			location.TemplateStack = &certificate.TemplateStackLocation{}
			var innerLocation CertificateImportTemplateStackLocation
			diags.Append(source.TemplateStack.As(ctx, &innerLocation, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return nil, diags
			}
			location.TemplateStack.PanoramaDevice = innerLocation.PanoramaDevice.ValueString()
			location.TemplateStack.TemplateStack = innerLocation.Name.ValueString()
		}

		if !source.TemplateStackVsys.IsNull() {
			location.TemplateStackVsys = &certificate.TemplateStackVsysLocation{}
			var innerLocation CertificateImportTemplateStackVsysLocation
			diags.Append(source.TemplateStack.As(ctx, &innerLocation, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return nil, diags
			}
			location.TemplateStackVsys.Vsys = innerLocation.Vsys.ValueString()
			location.TemplateStackVsys.NgfwDevice = innerLocation.NgfwDevice.ValueString()
			location.TemplateStackVsys.PanoramaDevice = innerLocation.PanoramaDevice.ValueString()
			location.TemplateStackVsys.TemplateStack = innerLocation.TemplateStack.ValueString()
		}
	}

	return &location, diags
}

func (o *CertificateImportResource) getImportLocationExtras(ctx context.Context, state CertificateImportResourceModel) (string, string, diag.Diagnostics) {
	var location CertificateImportLocation
	diags := state.Location.As(ctx, &location, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return "", "", diags
	}

	if !location.Template.IsNull() {
		var innerLocation CertificateImportTemplateLocation
		diags.Append(location.Template.As(ctx, &innerLocation, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return "", "", diags
		}
		return innerLocation.Name.ValueString(), "", nil
	} else if !location.TemplateStack.IsNull() {
		var innerLocation CertificateImportTemplateStackLocation
		diags.Append(location.Template.As(ctx, &innerLocation, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return "", "", diags
		}
		return innerLocation.Name.ValueString(), "", nil
	} else if !location.TemplateVsys.IsNull() {
		var innerLocation CertificateImportTemplateVsysLocation
		diags.Append(location.TemplateVsys.As(ctx, &innerLocation, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return "", "", diags
		}
		return innerLocation.Template.ValueString(), innerLocation.Vsys.ValueString(), nil
	} else if !location.TemplateStackVsys.IsNull() {
		var innerLocation CertificateImportTemplateStackVsysLocation
		diags.Append(location.TemplateVsys.As(ctx, &innerLocation, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return "", "", diags
		}
		return innerLocation.TemplateStack.ValueString(), innerLocation.Vsys.ValueString(), nil
	}

	return "", "", nil
}

func (o *CertificateImportResource) ReadCustom(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CertificateImportResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var terraformLocation CertificateImportLocation
	resp.Diagnostics.Append(state.Location.As(ctx, &terraformLocation, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdkLocation, diags := o.terraformToPangoLocation(ctx, terraformLocation)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	service := certificate.NewService(o.client)

	obj, err := service.Read(ctx, *sdkLocation, name, "get")
	if err != nil && !sdkerrors.IsObjectNotFound(err) {
		resp.Diagnostics.AddError("Failed to create resource", err.Error())
	}

	if obj == nil {
		return
	}

	if state.Local == nil {
		return
	}

	if state.Local.Pem != nil {
		state.Local.Pem.Certificate = types.StringPointerValue(obj.PublicKey)
		state.Local.Pem.PrivateKey = types.StringPointerValue(obj.PrivateKey)
	} else if state.Local.Pkcs12 != nil {

	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *CertificateImportResource) CreateCustom(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state CertificateImportResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var terraformLocation CertificateImportLocation
	resp.Diagnostics.Append(state.Location.As(ctx, &terraformLocation, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdkLocation, diags := o.terraformToPangoLocation(ctx, terraformLocation)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	service := certificate.NewService(o.client)
	obj, err := service.Read(ctx, *sdkLocation, name, "get")
	if err != nil && !sdkerrors.IsObjectNotFound(err) {
		resp.Diagnostics.AddError("Failed to create resource", err.Error())
	}

	if obj != nil {
		resp.Diagnostics.AddError("Failed to create resource", sdkmanager.ErrConflict.Error())
	}

	template, vsys, diags := o.getImportLocationExtras(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err = o.importCertificate(ctx, &state, template, vsys)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create resource", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *CertificateImportResource) UpdateCustom(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan CertificateImportResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var terraformLocation CertificateImportLocation
	resp.Diagnostics.Append(state.Location.As(ctx, &terraformLocation, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdkLocation, diags := o.terraformToPangoLocation(ctx, terraformLocation)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := certificate.NewService(o.client)

	certificateRenamed := state.Name.ValueString() != plan.Name.ValueString()
	if certificateRenamed {
		obj, err := service.Read(ctx, *sdkLocation, plan.Name.ValueString(), "get")
		if err != nil && !sdkerrors.IsObjectNotFound(err) {
			resp.Diagnostics.AddError("Failed to create resource", err.Error())
			return
		}

		if obj != nil {
			resp.Diagnostics.AddError("Failed to create resource", sdkmanager.ErrConflict.Error())
			return
		}
	}

	template, vsys, diags := o.getImportLocationExtras(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := o.importCertificate(ctx, &plan, template, vsys)
	if err != nil {
		resp.Diagnostics.AddError("Failed to import certificate", err.Error())
		return
	}

	if certificateRenamed {
		err := service.Delete(ctx, *sdkLocation, state.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to delete old certificate after rename", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *CertificateImportResource) DeleteCustom(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CertificateImportResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var terraformLocation CertificateImportLocation
	resp.Diagnostics.Append(state.Location.As(ctx, &terraformLocation, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdkLocation, diags := o.terraformToPangoLocation(ctx, terraformLocation)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	service := certificate.NewService(o.client)
	err := service.Delete(ctx, *sdkLocation, name)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete certificate from the device", err.Error())
	}
}

func certificateExportErrorType(err error) error {
	if err == nil {
		return nil
	}

	message := err.Error()
	if strings.HasSuffix(message, "private key may be blocked") {
		return certificateExportErrorPrivateKey
	}

	return certificateExportErrorOther
}
