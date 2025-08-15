package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// SwaggerSpec represents the structure of a Swagger/OpenAPI specification
type SwaggerSpec struct {
	Swagger             string                 `json:"swagger"`
	Info                Info                   `json:"info"`
	Host                string                 `json:"host"`
	BasePath            string                 `json:"basePath"`
	Schemes             []string               `json:"schemes"`
	Consumes            []string               `json:"consumes"`
	Produces            []string               `json:"produces"`
	Paths               map[string]interface{} `json:"paths"`
	Definitions         map[string]interface{} `json:"definitions"`
	SecurityDefinitions map[string]interface{} `json:"securityDefinitions"`
	Security            []map[string][]string  `json:"security"`
	Tags                []Tag                  `json:"tags"`
}

type Info struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func main() {
	// Find all swagger files in proto directory
	swaggerFiles, err := findSwaggerFiles("proto")
	if err != nil {
		log.Fatalf("Error finding swagger files: %v", err)
	}

	if len(swaggerFiles) == 0 {
		log.Fatal("No swagger files found in proto directory")
	}

	fmt.Printf("Found %d swagger files: %v\n", len(swaggerFiles), swaggerFiles)

	// Initialize merged swagger
	merged := SwaggerSpec{
		Swagger: "2.0",
		Info: Info{
			Title:       "Go Clean DDD ES Template API",
			Description: "A comprehensive API for the Go Clean DDD ES Template with Event Sourcing",
			Version:     "1.0.0",
		},
		Host:                "localhost:8080",
		BasePath:            "",
		Schemes:             []string{"http", "https"},
		Consumes:            []string{"application/json"},
		Produces:            []string{"application/json"},
		Paths:               make(map[string]interface{}),
		Definitions:         make(map[string]interface{}),
		SecurityDefinitions: make(map[string]interface{}),
		Security: []map[string][]string{
			{"BearerAuth": {}},
		},
		Tags: []Tag{},
	}

	// Add security definitions
	merged.SecurityDefinitions["BearerAuth"] = map[string]interface{}{
		"type":        "apiKey",
		"name":        "Authorization",
		"in":          "header",
		"description": "JWT token in format: Bearer <token>",
	}

	// Process each swagger file
	for _, filePath := range swaggerFiles {
		if err := mergeSwaggerFile(&merged, filePath); err != nil {
			log.Printf("Error processing %s: %v", filePath, err)
			continue
		}
	}

	// Write merged swagger to file
	outputPath := "docs/swagger.json"
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		log.Fatalf("Error creating docs directory: %v", err)
	}

	outputData, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling merged swagger: %v", err)
	}

	if err := ioutil.WriteFile(outputPath, outputData, 0o644); err != nil {
		log.Fatalf("Error writing merged swagger: %v", err)
	}

	fmt.Printf("✅ Successfully merged swagger files into %s\n", outputPath)
}

func findSwaggerFiles(rootDir string) ([]string, error) {
	var swaggerFiles []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if file ends with .swagger.json
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			baseName := filepath.Base(path)
			if len(baseName) > 13 && baseName[len(baseName)-13:] == ".swagger.json" {
				swaggerFiles = append(swaggerFiles, path)
			}
		}

		return nil
	})

	return swaggerFiles, err
}

func mergeSwaggerFile(merged *SwaggerSpec, filePath string) error {
	// Read swagger file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	// Parse swagger JSON
	var swagger SwaggerSpec
	if err := json.Unmarshal(data, &swagger); err != nil {
		return fmt.Errorf("error parsing swagger JSON: %v", err)
	}

	// Merge paths and normalize them
	for path, pathItem := range swagger.Paths {
		// Normalize path: remove /api prefix and ensure consistent format
		normalizedPath := path
		if strings.HasPrefix(normalizedPath, "/api/v1/") {
			normalizedPath = strings.Replace(normalizedPath, "/api/v1/", "/v1/", 1)
		} else if strings.HasPrefix(normalizedPath, "/v1/") {
			// Already in correct format
		} else {
			// Add /v1 prefix if not present
			if !strings.HasPrefix(normalizedPath, "/") {
				normalizedPath = "/" + normalizedPath
			}
			if !strings.HasPrefix(normalizedPath, "/v1/") {
				normalizedPath = "/v1" + normalizedPath
			}
		}
		merged.Paths[normalizedPath] = pathItem
	}

	// Merge definitions
	for defName, def := range swagger.Definitions {
		merged.Definitions[defName] = def
	}

	// Merge security definitions (if any)
	for secName, secDef := range swagger.SecurityDefinitions {
		merged.SecurityDefinitions[secName] = secDef
	}

	// Add tag from directory name
	dirName := filepath.Base(filepath.Dir(filePath))
	tagName := capitalizeFirst(dirName)

	// Check if tag already exists
	tagExists := false
	for _, tag := range merged.Tags {
		if tag.Name == tagName {
			tagExists = true
			break
		}
	}

	if !tagExists {
		merged.Tags = append(merged.Tags, Tag{
			Name:        tagName,
			Description: fmt.Sprintf("%s operations", tagName),
		})
	}

	fmt.Printf("✅ Merged %s (added tag: %s)\n", filePath, tagName)
	return nil
}

func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]&^32) + s[1:]
}
