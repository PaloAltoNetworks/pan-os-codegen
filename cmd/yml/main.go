package main

import (
	"fmt"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/parsing"
	"os"
)

func main() {
	argsWithoutProg := os.Args[1:]

	fmt.Printf("------------ YAML file ------------\n%v\n\n", argsWithoutProg[0])
	yamlFile, _ := parsing.ReadDataFromFile(argsWithoutProg[0])

	yamlParser, _ := parsing.NewYamlSpecParser(yamlFile)
	fmt.Printf("------------ YAML content ------------\n%v\n\n", yamlParser)

	yamlString, _ := yamlParser.Dump()
	fmt.Printf("------------ YAML dump ------------\n%s\n\n", yamlString)
}
