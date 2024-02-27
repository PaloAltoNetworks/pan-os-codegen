package parsing

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

func ReadDataFromFile(filename string) ([]byte, error) {
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return yamlFile, err
}

func MarshallYaml(yamlData map[interface{}]interface{}) (string, error) {
	yamlDump, err := yaml.Marshal(&yamlData)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return string(yamlDump), err
}

func UnmarshallYaml(inputData []byte) (map[interface{}]interface{}, error) {
	yamlData := make(map[interface{}]interface{})

	err := yaml.Unmarshal(inputData, &yamlData)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return yamlData, err
}
