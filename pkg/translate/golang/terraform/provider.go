package terraform

import (
	"github.com/paloaltonetworks/pan-os-codegen/pkg/load"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"strings"
)

const (
	Resource   = "resources"
	DataSource = "dataSources"
)

// CreateResourceList creates a string containing a list of resources or data sources based on the elementType parameter.
// The function returns the generated list as a string.
func CreateResourceList(elementType string) string {
	var builder strings.Builder

	elements, _ := properties.GetNormalizations()

	var specNames []string
	for _, specPath := range elements {
		content, err := load.File(specPath)
		if err != nil {
		}
		config, err := properties.ParseSpec(content)
		if err != nil {
		}
		specNames = append(specNames, config.Name)
	}

	switch elementType {
	case Resource:
		for _, resource := range specNames {
			builder.WriteString(naming.CamelCase("", resource, "", true) + "ObjectResource,\n")
		}
	case DataSource:
		for _, datasource := range specNames {
			builder.WriteString(naming.CamelCase("", datasource, "", true) + "ObjectListDataSource,\n")
		}
	}

	// Remove last newline
	result := builder.String()
	if len(result) > 1 { // Check to avoid out of bounds
		result = result[:len(result)-1]
	}

	return result
}
