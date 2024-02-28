package parsing

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type YamlSpecParser struct {
	Name                    string                      `yaml:"name"`
	TerraformProviderSuffix string                      `yaml:"terraform_provider_suffix"`
	GoSdkPath               []string                    `yaml:"go_sdk_path"`
	XpathSuffix             []string                    `yaml:"xpath_suffix"`
	Locations               map[interface{}]interface{} `yaml:"locations"`
	Entry                   map[interface{}]interface{} `yaml:"entry"`
	Version                 string                      `yaml:"version"`
	Spec                    map[interface{}]interface{} `yaml:"spec"`
}

func ReadDataFromFile(filename string) ([]byte, error) {
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return yamlFile, err
}

func MarshallYaml(yamlData *YamlSpecParser) (string, error) {
	yamlDump, err := yaml.Marshal(&yamlData)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return string(yamlDump), err
}

func UnmarshallYaml(inputData []byte) (*YamlSpecParser, error) {
	yamlData := YamlSpecParser{}

	err := yaml.Unmarshal(inputData, &yamlData)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return &yamlData, err
}
