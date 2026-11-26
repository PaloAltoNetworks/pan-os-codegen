package provider

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/PaloAltoNetworks/pango"
	"github.com/PaloAltoNetworks/pango/device/administrator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func administratorTestConfig(t *testing.T, fields map[string]attr.Value) tfsdk.Config {
	t.Helper()
	ctx := context.Background()
	schema := AdministratorResourceSchema()
	attrTypes := schema.Type().(types.ObjectType).AttrTypes
	values := make(map[string]attr.Value)
	for name, typ := range attrTypes {
		value, err := typ.ValueFromTerraform(ctx, tftypes.NewValue(typ.TerraformType(ctx), nil))
		if err != nil {
			t.Fatal(err)
		}
		values[name] = value
	}
	for name, value := range fields {
		values[name] = value
	}
	raw, err := types.ObjectValueMust(attrTypes, values).ToTerraformValue(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return tfsdk.Config{Schema: schema, Raw: raw}
}

func TestAdministratorWriteOnlyValidation(t *testing.T) {
	for _, test := range []struct {
		name      string
		fields    map[string]attr.Value
		wantError bool
	}{
		{"paired", map[string]attr.Value{"password_wo": types.StringUnknown(), "password_wo_version": types.Int64Value(1)}, false},
		{"missing-version", map[string]attr.Value{"password_wo": types.StringValue("fixture")}, true},
		{"missing-password", map[string]attr.Value{"password_wo_version": types.Int64Value(1)}, true},
		{"zero-version", map[string]attr.Value{"password_wo": types.StringValue("fixture"), "password_wo_version": types.Int64Value(0)}, true},
		{"conflicting", map[string]attr.Value{"password": types.StringValue("legacy"), "password_wo": types.StringValue("fixture"), "password_wo_version": types.Int64Value(1)}, true},
		{"legacy", map[string]attr.Value{"password": types.StringValue("legacy")}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			resp := &resource.ValidateConfigResponse{}
			r := &AdministratorResource{}
			r.ValidateConfigCustom(context.Background(), resource.ValidateConfigRequest{Config: administratorTestConfig(t, test.fields)}, resp)
			if resp.Diagnostics.HasError() != test.wantError {
				t.Fatal(resp.Diagnostics)
			}
		})
	}
}

func TestAdministratorWriteOnlyCreateAndRotationUseConfig(t *testing.T) {
	password := "fixture<&password"
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if err := r.ParseForm(); err != nil {
			t.Error(err)
		}
		var cmd struct {
			Password string `xml:"password-hash>password"`
		}
		if err := xml.Unmarshal([]byte(r.Form.Get("cmd")), &cmd); err != nil {
			t.Error(err)
		}
		if cmd.Password != password {
			t.Error("hashing must use the apply config value, not the null plan")
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(`<response status="success"><result><phash>fixture-hash</phash></result></response>`))
	}))
	defer server.Close()
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	client := &pango.Client{Hostname: u.Host, Protocol: "http", ApiKey: "fixture", Transport: server.Client().Transport.(*http.Transport)}
	if err := client.Setup(); err != nil {
		t.Fatal(err)
	}
	r := &AdministratorResource{client: client}
	config := administratorTestConfig(t, map[string]attr.Value{"password_wo": types.StringValue(password), "password_wo_version": types.Int64Value(1)})
	model := AdministratorResourceModel{PasswordWo: types.StringNull(), PasswordWoVersion: types.Int64Value(1)}
	entry := &administrator.Entry{}
	createResp := &resource.CreateResponse{}
	r.PreCreate(context.Background(), resource.CreateRequest{Config: config}, createResp, &model, administrator.Location{}, entry, nil)
	if createResp.Diagnostics.HasError() {
		t.Fatal(createResp.Diagnostics)
	}
	if calls != 1 || entry.Phash == nil || *entry.Phash != "fixture-hash" {
		t.Fatal("create did not hash config password")
	}
	old := model
	old.PasswordWoVersion = types.Int64Value(0)
	updateResp := &resource.UpdateResponse{}
	r.PreUpdate(context.Background(), resource.UpdateRequest{Config: config}, updateResp, &model, &old, administrator.Location{}, entry, nil)
	if updateResp.Diagnostics.HasError() {
		t.Fatal(updateResp.Diagnostics)
	}
	if calls != 2 {
		t.Fatal("version change did not trigger rotation")
	}
}

func TestAdministratorWriteOnlyUnchangedVersionDoesNotHash(t *testing.T) {
	model := AdministratorResourceModel{PasswordWoVersion: types.Int64Value(4)}
	hash := "UNCHANGED-HASH-FIXTURE"
	entry := &administrator.Entry{Phash: &hash}
	// A nil client/config is deliberate: an unnecessary hash call would panic.
	r := &AdministratorResource{}
	resp := &resource.UpdateResponse{}
	r.PreUpdate(context.Background(), resource.UpdateRequest{}, resp, &model, &model, administrator.Location{}, entry, nil)
	if resp.Diagnostics.HasError() {
		t.Fatal(resp.Diagnostics)
	}
	if entry.Phash == nil || *entry.Phash != hash {
		t.Fatal("unexpected rotation")
	}
}

func TestAdministratorWriteOnlyDoesNotPersistCredentials(t *testing.T) {
	ctx := context.Background()
	model := AdministratorResourceModel{
		Name:              types.StringValue("review"),
		Password:          types.StringNull(),
		PasswordWo:        types.StringNull(),
		PasswordWoVersion: types.Int64Value(1),
	}
	hash := "SENSITIVE-HASH-FIXTURE"
	plaintext := "SENSITIVE-PASSWORD-FIXTURE"
	ev, err := NewEncryptedValuesManager(nil, false)
	if err != nil {
		t.Fatal(err)
	}
	// Simulate switching from the legacy password mode, including private state.
	if err := ev.StorePlaintextValue("password", HashingCustomType, plaintext); err != nil {
		t.Fatal(err)
	}
	if err := ev.StoreEncryptedValue("password", HashingCustomType, hash); err != nil {
		t.Fatal(err)
	}
	entry := &administrator.Entry{Name: "review", Phash: &hash}
	if d := model.CopyToPango(ctx, nil, nil, &entry, ev); d.HasError() {
		t.Fatal(d)
	}
	if entry.Phash == nil || *entry.Phash != hash {
		t.Fatal("unrelated updates must preserve the server hash")
	}
	if d := model.CopyFromPango(ctx, nil, nil, entry, ev); d.HasError() {
		t.Fatal(d)
	}
	if !model.Password.IsNull() || !model.PasswordWo.IsNull() {
		t.Fatal("password must remain null in state")
	}
	payload, err := json.Marshal(ev)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), hash) || strings.Contains(string(payload), plaintext) {
		t.Fatal("credential retained in private state")
	}
}
