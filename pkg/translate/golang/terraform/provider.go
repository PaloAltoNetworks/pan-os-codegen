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

	var results []string
	switch elementType {
	case Resource:
		for _, resource := range specNames {
			results = append(results, naming.CamelCase("", resource, "", true)+"ObjectResource,")
		}
	case DataSource:
		for _, datasource := range specNames {
			results = append(results, naming.CamelCase("", datasource, "", true)+"ObjectListDataSource,")
		}
	}

	return strings.Join(results, "\n")
}
