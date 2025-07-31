package provider

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/url"
	"strings"

	"software.sslmate.com/src/go-pkcs12"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/PaloAltoNetworks/pango/device/certificate"
	sdkerrors "github.com/PaloAltoNetworks/pango/errors"
	"github.com/PaloAltoNetworks/pango/locking"
	"github.com/PaloAltoNetworks/pango/util"
	"github.com/PaloAltoNetworks/pango/xmlapi"
	sdkmanager "github.com/PaloAltoNetworks/terraform-provider-panos/internal/manager"
)

func (o *CertificateImportResource) importCertificate(ctx context.Context, state *CertificateImportResourceModel, template string, vsys string) error {
	mutex := locking.GetMutex(locking.ImportFileLockCategory, "")
	mutex.Lock()
	defer mutex.Unlock()

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

			_, _, err := o.client.ImportFile(ctx, command, []byte(certificate), "cert.pem", "file", false, nil)
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

				command.Extras.Add("passphrase", passphrase)

				_, _, err := o.client.ImportFile(ctx, command, []byte(*privateKey), "key.pem", "file", false, nil)
				if err != nil {
					return fmt.Errorf("Failed to import PEM private key into PAN-OS: %w", err)
				}
			}
		} else if local.Pkcs12 != nil {
			command.Extras.Add("format", "pkcs12")

			command.Category = "certificate"
			encoded := []byte(local.Pkcs12.Certificate.ValueString())

			certificate := make([]byte, base64.StdEncoding.DecodedLen(len(encoded)))
			_, err := base64.StdEncoding.Decode(certificate, encoded)
			if err != nil {
				return err
			}

			passphrase := local.Pkcs12.Passphrase.ValueString()
			if passphrase == "" {
				passphrase = ""
			}
			command.Extras.Add("passphrase", passphrase)

			_, _, err = o.client.ImportFile(ctx, command, certificate, "cert.pkcs12", "file", false, nil)
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

func serverPkcs12CertificateDiffersFromState(pkcs12Bundle types.String, passwordValue types.String, serverCert *string, serverPrivateKey *string) (bool, error) {
	pkcs12BundleValue := pkcs12Bundle.ValueString()
	encoded, err := base64.StdEncoding.DecodeString(pkcs12BundleValue)
	if err != nil {
		return false, err
	}

	password := passwordValue.ValueString()
	_, stateCert, err := pkcs12.Decode(encoded, password)
	if err != nil {
		return false, err
	}

	var stateCertPem bytes.Buffer
	err = pem.Encode(&stateCertPem, &pem.Block{Type: "CERTIFICATE", Bytes: stateCert.Raw})
	if err != nil {
		return false, err
	}

	if serverCert == nil {
		return true, nil
	}

	if strings.TrimSpace(*serverCert) != strings.TrimSpace(stateCertPem.String()) {
		return true, nil
	}

	return false, nil
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

	xpath, err := sdkLocation.XpathWithComponents(o.client.Versioning(), util.AsEntryXpath(name))
	if err != nil {
		resp.Diagnostics.AddError("Failed to read certificate from the device", err.Error())
	}

	obj, err := service.ReadWithXpath(ctx, util.AsXpath(xpath), "get")
	if err != nil && !sdkerrors.IsObjectNotFound(err) {
		resp.Diagnostics.AddError("Failed to create resource", err.Error())
	}

	if obj == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	if state.Local == nil {
		return
	}

	if state.Local.Pem != nil {
		if obj.PublicKey == nil {
			state.Local.Pem.Certificate = types.StringNull()
		} else {
			state.Local.Pem.Certificate = types.StringValue(strings.TrimSpace(*obj.PublicKey))
		}
	} else if state.Local.Pkcs12 != nil {
		if obj.PublicKey == nil {
			state.Local.Pkcs12.Certificate = types.StringNull()
			return
		}
		changed, err := serverPkcs12CertificateDiffersFromState(state.Local.Pkcs12.Certificate, state.Local.Pkcs12.Passphrase, obj.PublicKey, obj.PrivateKey)
		if err != nil {
			resp.Diagnostics.AddError("Failed to read certificate from the server", err.Error())
			return
		}

		if changed {
			state.Local.Pkcs12.Certificate = types.StringValue("[outdated]")
		}
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

	xpath, err := sdkLocation.XpathWithComponents(o.client.Versioning(), util.AsEntryXpath(name))
	if err != nil {
		resp.Diagnostics.AddError("Failed to read certificate from the device", err.Error())
	}

	service := certificate.NewService(o.client)
	obj, err := service.ReadWithXpath(ctx, util.AsXpath(xpath), "get")
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

	xpath, err := sdkLocation.XpathWithComponents(o.client.Versioning(), util.AsEntryXpath(plan.Name.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to read certificate from the device", err.Error())
	}

	service := certificate.NewService(o.client)

	certificateRenamed := state.Name.ValueString() != plan.Name.ValueString()
	if certificateRenamed {
		obj, err := service.ReadWithXpath(ctx, util.AsXpath(xpath), "get")
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

	err = o.importCertificate(ctx, &plan, template, vsys)
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
