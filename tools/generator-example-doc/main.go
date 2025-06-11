package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"slices"
	"strings"
)

var template string

type ResourceType struct {
	ResourceType string `json:"resourceType"`
	FriendlyName string `json:"friendlyName"`
	UrlValue     string `json:"urlValue"`
}

var resourceTypes map[string]ResourceType

func init() {
	var resourceTypeItems []ResourceType
	resourceTypeJsonPath := path.Join("tools", "generator-example-doc", "resource_types.json")
	// #nosec G304
	data, err := os.ReadFile(resourceTypeJsonPath)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &resourceTypeItems)
	if err != nil {
		panic(err)
	}
	resourceTypes = make(map[string]ResourceType)
	for _, item := range resourceTypeItems {
		resourceTypes[strings.ToLower(item.ResourceType)] = item
	}

	data, err = os.ReadFile(path.Join("tools", "generator-example-doc", "template.md"))
	if err != nil {
		panic(err)
	}
	template = string(data)
}

func main() {
	inputDir := flag.String("input-dir", "./examples/quickstarts", "directory to scan for example files")
	outputDir := flag.String("output-dir", "./docs/resources", "directory to write documentation files")

	flag.Parse()
	if *inputDir == "" || *outputDir == "" {
		log.Fatal("input-dir and output-dir flags are required")
	}

	resourceTypeDirs, err := os.ReadDir(*inputDir)
	if err != nil {
		log.Fatalf("Error reading input directory: %s", err)
	}

	for _, resourceTypeDir := range resourceTypeDirs {
		if !resourceTypeDir.IsDir() {
			continue
		}

		content, err := generateDocumentation(path.Join(*inputDir, resourceTypeDir.Name()))
		if err != nil {
			log.Fatalf("Error generating documentation for %s: %s", resourceTypeDir.Name(), err)
		}

		resourceTypeName := strings.Split(resourceTypeDir.Name(), "@")[0]
		outputFile := path.Join(*outputDir, resourceTypeName+".md")
		// #nosec G306
		err = os.WriteFile(outputFile, []byte(content), 0o644)
		if err != nil {
			log.Printf("Error writing documentation for %s: %s", resourceTypeDir.Name(), err)
		}
	}
}

func generateDocumentation(inputDir string) (string, error) {
	resourceType := path.Base(inputDir)
	resourceType = strings.ReplaceAll(resourceType, "_", "/")

	resourceTypeFriendlyName := GetResourceTypeFriendlyName(resourceType)
	if resourceTypeFriendlyName == "" {
		return "", fmt.Errorf("resource type %s friendly name not found, please add it to resource_types.json", resourceType)
	}

	out := template
	out = strings.ReplaceAll(out, "{{.resource_type}}", resourceType)
	out = strings.ReplaceAll(out, "{{.resource_type_friendly_name}}", resourceTypeFriendlyName)
	out = strings.ReplaceAll(out, "{{.reference_link}}", fmt.Sprintf("https://learn.microsoft.com/en-us/azure/templates/%s?pivots=deployment-language-terraform", resourceType))
	out = strings.ReplaceAll(out, "{{.resource_id}}", GetImportResourceId(resourceType))
	out = strings.ReplaceAll(out, "{{.url}}", GetUrlValue(resourceType))

	// key is the scenario name, value is the example content
	exampleMap := make(map[string]string)
	scenarioDirs, err := os.ReadDir(inputDir)
	if err != nil {
		return "", fmt.Errorf("error reading directory: %w", err)
	}
	for _, scenarioDir := range scenarioDirs {
		if !scenarioDir.IsDir() || scenarioDir.Name() == "testdata" {
			continue
		}

		scenarioName := scenarioDir.Name()
		exampleFilePath := path.Join(inputDir, scenarioName, "main.tf")
		// #nosec G304
		exampleContent, err := os.ReadFile(exampleFilePath)
		if err != nil {
			log.Printf("Error reading example file for %s: %s", exampleFilePath, err)
			continue
		}

		exampleMap[scenarioName] = string(exampleContent)
	}
	// check if there's main.tf in the inputDir
	mainFilePath := path.Join(inputDir, "main.tf")
	if _, err := os.Stat(mainFilePath); err == nil {
		// #nosec G304
		exampleContent, err := os.ReadFile(mainFilePath)
		if err != nil {
			log.Printf("Error reading example file for %s: %s", mainFilePath, err)
			return "", err
		}
		exampleMap["default"] = string(exampleContent)
	}

	scenarioNames := make([]string, 0)
	for scenarioName := range exampleMap {
		scenarioNames = append(scenarioNames, scenarioName)
	}
	slices.Sort(scenarioNames)

	example := ""
	for _, scenarioName := range scenarioNames {
		exampleContent := exampleMap[scenarioName]
		example += fmt.Sprintf("### %s\n\n", scenarioName)
		example += fmt.Sprintf("```hcl\n%s\n```\n\n", exampleContent)
	}
	out = strings.ReplaceAll(out, "{{.example}}", example)

	return out, nil
}

func GetResourceTypeFriendlyName(resourceType string) string {
	if v, ok := resourceTypes[strings.ToLower(resourceType)]; ok {
		return v.FriendlyName
	}
	return ""
}

func GetUrlValue(resourceType string) string {
	if v, ok := resourceTypes[strings.ToLower(resourceType)]; ok {
		return v.UrlValue
	}
	return ""
}

func GetImportResourceId(resourceType string) string {
	v, ok := resourceTypes[strings.ToLower(resourceType)]
	if !ok {
		return ""
	}

	out := fmt.Sprintf("/%s", v.UrlValue)

	endsWithRef := strings.HasSuffix(out, "/$ref")
	out = strings.TrimSuffix(out, "/$ref")
	lastSegment := out[strings.LastIndex(out, "/")+1:]
	out += fmt.Sprintf("/{%s-id}", lastSegment)

	if endsWithRef {
		out += "/$ref"
	}
	return out
}
