package translate

import (
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

func isParamListAndProfileTypeIsMember(param *properties.SpecParam) bool {
	return param.Type == "list" && param.Profiles != nil && len(param.Profiles) > 0 && param.Profiles[0].Type == "member"
}

func isParamListAndProfileTypeIsSingleEntry(param *properties.SpecParam) bool {
	return param.Type == "list" && param.Profiles != nil && len(param.Profiles) > 0 && param.Profiles[0].Type == "entry" && param.Items != nil && param.Items.Type == "string"
}

func isParamListAndProfileTypeIsExtendedEntry(param *properties.SpecParam) bool {
	return param != nil && param.Type == "list" && param.Profiles != nil && len(param.Profiles) > 0 && param.Profiles[0].Type == "entry" && param.Items != nil && param.Items.Type != "string"
}
