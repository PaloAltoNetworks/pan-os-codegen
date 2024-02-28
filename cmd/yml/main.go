package main

import (
	"fmt"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/parsing"
	"os"
)

func main() {
	argsWithoutProg := os.Args[1:]
	yamlFile, _ := parsing.ReadDataFromFile(argsWithoutProg[0])
	//yamlFile, _ := parsing.ReadDataFromFile("/Users/sczech/Work/code/pan-os-codegen/specs/objects/address.yml")

	yamlData, _ := parsing.UnmarshallYaml(yamlFile)
	fmt.Printf("------------ YAML content ------------\n%v\n\n", yamlData)

	yamlString, _ := parsing.MarshallYaml(yamlData)
	fmt.Printf("------------ YAML dump ------------\n%s\n\n", yamlString)
}
