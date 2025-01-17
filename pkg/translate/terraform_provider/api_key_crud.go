package terraform_provider

const apiKeyImports = `
import (
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
)
`
const apiKeyOpen = `
var data ApiKeyResourceModel
resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
if resp.Diagnostics.HasError() {
	return
}

username := data.Username.ValueString()
password := data.Password.ValueString()

apiKey, err := r.client.GenerateApiKey(ctx, username, password)
if err != nil {
	resp.Diagnostics.AddError("failed to generate API key", err.Error())
	return
}

data.ApiKey = types.StringValue(apiKey)
resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
`
