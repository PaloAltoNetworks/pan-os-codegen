package terraform_provider

const certificateImportImports = ``
const certificateImportCommon = `
type _ = diag.Diagnostics
var _ = tflog.Info
var _ = errors.ErrUnsupported
`

const certificateImportResourceRead = `
o.ReadCustom(ctx, req, resp)
`

const certificateImportCreate = `
r.CreateCustom(ctx, req, resp)
`

const certificateImportUpdate = `
r.UpdateCustom(ctx, req, resp)
`

const certificateImportDelete = `
r.DeleteCustom(ctx, req, resp)
`
