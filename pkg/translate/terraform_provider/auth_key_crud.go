package terraform_provider

const authKeyImports = `
import (
	"encoding/xml"
        "regexp"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/PaloAltoNetworks/pango/xmlapi"
)
`

const authKeyCommon = `
type authKeyRequest struct {
	XMLName xml.Name ` + "`" + `xml:"request"` + "`" + `
        Lifetime int64 ` + "`" + `xml:"bootstrap>vm-auth-key>generate>lifetime"` + "`" + `
}

type authKeyResponse struct {
	XMLName xml.Name ` + "`" + `xml:"response"` + "`" + `
	Result string ` + "`" + `xml:"result"` + "`" + `
}
`

const authKeyOpen = `
var data AuthKeyResourceModel
resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
if resp.Diagnostics.HasError() {
	return
}

lifetime := data.Lifetime.ValueInt64()

cmd := &xmlapi.Op{
	Command: authKeyRequest{Lifetime: lifetime},
}

var serverResponse authKeyResponse
if _, _, err := r.client.Communicate(ctx, cmd, false, &serverResponse); err != nil {
	resp.Diagnostics.AddError("Failed to generate Authenticaion Key", "Server returned an error: " + err.Error())
	return
}

authKeyRegexp := ` + "`" + `VM auth key (?P<authkey>.+) generated. Expires at: (?P<expiration>.+)` + "`" + `
expr := regexp.MustCompile(authKeyRegexp)
match := expr.FindStringSubmatch(serverResponse.Result)
if match == nil {
	resp.Diagnostics.AddError("Failed to parse server response", "Server response did not match regular expression")
	return
}

groups := make(map[string]string)
for i, name := range expr.SubexpNames() {
	if i != 0 && name != "" {
		groups[name] = match[i]
	}
}


if authKey, found := groups["authkey"]; found {
	data.Authkey = types.StringValue(authKey)
} else {
	resp.Diagnostics.AddError("Failed to parse server response", "Server response did not contain matching authentication key")
	return
}

if expiration, found := groups["expiration"]; found {
	data.ExpirationDate = types.StringValue(expiration)
} else {
	resp.Diagnostics.AddWarning("Incomplete server response", "Server response didn't contain a valid expiration date")
}

resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
`
