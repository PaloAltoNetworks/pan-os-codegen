package provider

import (
	"bytes"
	"context"
	"encoding/pem"
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

func (o *CertificateImportResource) importCertificate(ctx context.Context, state *CertificateImportResourceModel) error {
	command := &xmlapi.Import{}
	command.Extras = url.Values{}

	command.Extras.Set("certificate-name", state.Name.ValueString())

	if state.Local != nil {
		local := state.Local
		if local.Pem != nil {
			command.Extras.Add("format", "pem")

			command.Category = "certificate"
			certificate := local.Pem.Certificate.ValueString()

			_, _, err := o.client.ImportFile(ctx, command, certificate, "cert.pem", "file", false, nil)
			if err != nil {
				return fmt.Errorf("Failed to import PEM certificate into PAN-OS: %w", err)
			}

			command.Category = "private-key"
			privateKey := local.Pem.PrivateKey.ValueStringPointer()
			if privateKey != nil {
				command.Extras.Add("passphrase", local.Pem.Passphrase.ValueString())
				_, _, err := o.client.ImportFile(ctx, command, *privateKey, "cert-key.pem", "file", false, nil)
				if err != nil {
					return fmt.Errorf("Failed to import PEM certificate into PAN-OS: %w", err)
				}
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

		if !source.DeviceGroup.IsNull() {
			location.DeviceGroup = &certificate.DeviceGroupLocation{}
			var innerLocation CertificateImportDeviceGroupLocation
			diags.Append(source.DeviceGroup.As(ctx, &innerLocation, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return nil, diags
			}
			location.DeviceGroup.PanoramaDevice = innerLocation.PanoramaDevice.ValueString()
			location.DeviceGroup.DeviceGroup = innerLocation.Name.ValueString()
		}
	}

	return &location, diags
}

func (o *CertificateImportResource) ReadCustom(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CertificateImportResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	export := &xmlapi.Export{
		Category: "certificate",
	}
	export.Extras = url.Values{}

	export.Extras.Set("certificate-name", state.Name.ValueString())

	var statePrivateKeyData []byte
	var statePrivateKeyDataEncoded []byte
	if state.Local != nil {
		local := state.Local
		if local.Pem != nil {
			export.Extras.Add("format", "pem")

			privateKey := local.Pem.PrivateKey.ValueStringPointer()
			if privateKey != nil {
				for block, rest := pem.Decode([]byte(*privateKey)); block != nil; block, rest = pem.Decode(rest) {
					switch block.Type {
					case "PRIVATE KEY", "ENCRYPTED PRIVATE KEY":
						if statePrivateKeyData != nil {
							resp.Diagnostics.AddError("Failed to parse PEM input", "Multiple private keys found")
							return
						}
						statePrivateKeyData = block.Bytes
						statePrivateKeyDataEncoded = pem.EncodeToMemory(block)
					}
				}
				export.Extras.Add("include-key", "yes")
				export.Extras.Add("passphrase", local.Pem.Passphrase.ValueString())
			} else {
				export.Extras.Add("include-key", "no")
			}
		}
	}

	_, data, _, err := o.client.ExportFile(ctx, export, nil)
	if err != nil {
		switch certificateExportErrorType(err) {
		case certificateExportErrorOther:
			resp.Diagnostics.AddError("Failed to read certificate from server", err.Error())
			return
		}
	}

	var certData []byte
	var privateKeyData []byte

	for block, rest := pem.Decode(data); block != nil; block, rest = pem.Decode(rest) {
		switch block.Type {
		case "CERTIFICATE":
			certData = pem.EncodeToMemory(block)
		case "PRIVATE KEY", "ENCRYPTED PRIVATE KEY":
			privateKeyData = block.Bytes
		}
	}

	state.Local.Pem.Certificate = types.StringValue(string(certData))
	if !bytes.Equal(statePrivateKeyData, privateKeyData) {
		state.Local.Pem.PrivateKey = types.StringValue(string(statePrivateKeyDataEncoded))
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

	err = o.importCertificate(ctx, &state)
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

	certificateRenamed := state.Name.ValueString() != plan.Name.ValueString()
	if certificateRenamed {
		// check if rename would override another certificate
	}

	err := o.importCertificate(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Failed to import certificate", err.Error())
		return
	}

	if certificateRenamed {
		// delete certificate under the old name
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
