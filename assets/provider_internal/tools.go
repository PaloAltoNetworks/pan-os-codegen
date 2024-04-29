package provider

import (
	"fmt"
	"strings"
)

type RuleInfo struct {
	Name string `json:"name"`
	Uuid string `json:"uuid"`
}

func ProviderParamDescription(desc, defaultValue, envName, jsonName string) string {
	var b strings.Builder

	b.WriteString(desc)

	if defaultValue != "" {
		b.WriteString(fmt.Sprintf(" Default: `%s`.", defaultValue))
	}

	if envName != "" {
		b.WriteString(fmt.Sprintf(" Environment variable: `%s`.", envName))
	}

	if jsonName != "" {
		b.WriteString(fmt.Sprintf(" JSON config file variable: `%s`.", jsonName))
	}

	return b.String()
}
