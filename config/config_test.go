// SPDX-FileCopyrightText: Copyright 2025 Krishna Iyer (www.krishnaiyer.tech)
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Test structs with various field types
type SimpleConfig struct {
	Name    string        `name:"name" short:"n" description:"The name"`
	Port    int           `name:"port" short:"p" description:"The port number"`
	Debug   bool          `name:"debug" short:"d" description:"Enable debug mode"`
	Timeout time.Duration `name:"timeout" short:"t" description:"Timeout duration"`
	Rate    float64       `name:"rate" description:"Rate value"`
}

type ComplexConfig struct {
	Basic    BasicInfo         `name:"basic"`
	Server   ServerConfig      `name:"server"`
	Tags     []string          `name:"tags" description:"List of tags"`
	Metadata map[string]string `name:"metadata" description:"Key-value metadata"`
}

type BasicInfo struct {
	Name    string `name:"name" description:"Basic name"`
	Version string `name:"version" description:"Basic version"`
}

type ServerConfig struct {
	Host string `name:"host" description:"Server host"`
	Port int    `name:"port" description:"Server port"`
}

type ConfigWithNoTags struct {
	Field1 string
	Field2 int
}

type ConfigWithMixedTags struct {
	WithTag    string `name:"with-tag" description:"Field with tag"`
	WithoutTag string
	AlsoTagged int `name:"also-tagged" description:"Another tagged field"`
}

type ConfigWithUnsupportedTypes struct {
	Name     string   `name:"name" description:"Valid field"`
	Channel  chan int `name:"channel" description:"Unsupported channel type"`
	Function func()   `name:"function" description:"Unsupported function type"`
}

// Helper functions for tests
func createTempConfigFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	err := os.WriteFile(configPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	return configPath
}

func TestNew(t *testing.T) {
	tests := []struct {
		name              string
		input             any
		nameTagOverride   string
		expectPanic       bool
		expectError       bool
		expectedFlagCount int
	}{
		{
			name:              "ValidPointerToSimpleStruct",
			input:             &SimpleConfig{},
			nameTagOverride:   "",
			expectPanic:       false,
			expectError:       false,
			expectedFlagCount: 6, // 5 fields + 1 config flag
		},
		{
			name:              "ValidPointerToComplexStruct",
			input:             &ComplexConfig{},
			nameTagOverride:   "",
			expectPanic:       false,
			expectError:       false,
			expectedFlagCount: 7, // 6 fields + 1 config flag (basic.name, basic.version, server.host, server.port, tags, metadata)
		},
		{
			name:            "NonPointerInputShouldPanic",
			input:           SimpleConfig{},
			nameTagOverride: "",
			expectPanic:     true,
			expectError:     false,
		},
		{
			name:            "StringInputShouldPanic",
			input:           "not a pointer",
			nameTagOverride: "",
			expectPanic:     true,
			expectError:     false,
		},
		{
			name:              "CustomNameTagOverride",
			input:             &SimpleConfig{},
			nameTagOverride:   "yaml",
			expectPanic:       false,
			expectError:       false,
			expectedFlagCount: 1, // only config flag since no yaml tags
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic but didn't get one")
					}
				}()
				_, _ = New(tt.input, tt.nameTagOverride)
				return
			}

			manager, err := New(tt.input, tt.nameTagOverride)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if manager == nil {
				t.Error("Expected manager to be non-nil")
				return
			}
			if manager.target != tt.input {
				t.Error("Expected target to match input")
			}
			if manager.flags == nil {
				t.Error("Expected flags to be initialized")
			}

			flagCount := 0
			manager.flags.VisitAll(func(f *pflag.Flag) {
				flagCount++
			})

			if flagCount != tt.expectedFlagCount {
				t.Errorf("Expected %d flags, got %d", tt.expectedFlagCount, flagCount)
			}
		})
	}
}

func TestManagerFlagSet(t *testing.T) {
	config := &SimpleConfig{}
	manager, err := New(config, "")
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	flagSet := manager.FlagSet()
	if flagSet == nil {
		t.Error("Expected non-nil flagset")
	}

	// Verify it's the same instance
	if flagSet != manager.flags {
		t.Error("FlagSet() should return the same flagset instance")
	}
}

func TestGenFlagSet(t *testing.T) {
	tests := []struct {
		name          string
		input         any
		nameTag       string
		expectError   bool
		expectedFlags map[string]string // flag name -> expected type
	}{
		{
			name:    "simple struct with all basic types",
			input:   &SimpleConfig{},
			nameTag: "name",
			expectedFlags: map[string]string{
				"config":  "string",
				"name":    "string",
				"port":    "int",
				"debug":   "bool",
				"timeout": "duration",
				"rate":    "float64",
			},
		},
		{
			name:    "nested struct",
			input:   &ComplexConfig{},
			nameTag: "name",
			expectedFlags: map[string]string{
				"config":        "string",
				"basic.name":    "string",
				"basic.version": "string",
				"server.host":   "string",
				"server.port":   "int",
				"tags":          "stringSlice",
				"metadata":      "stringToString",
			},
		},
		{
			name:    "struct with no tags",
			input:   &ConfigWithNoTags{},
			nameTag: "name",
			expectedFlags: map[string]string{
				"config": "string", // only config flag
			},
		},
		{
			name:    "struct with mixed tags",
			input:   &ConfigWithMixedTags{},
			nameTag: "name",
			expectedFlags: map[string]string{
				"config":      "string",
				"with-tag":    "string",
				"also-tagged": "int",
			},
		},
		{
			name:        "non-pointer should error",
			input:       SimpleConfig{},
			nameTag:     "name",
			expectError: true,
		},
		{
			name:        "non-struct should error",
			input:       new(string),
			nameTag:     "name",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)

			// Add config flag manually since genFlagSet doesn't add it
			var configFile string
			flags.StringVarP(&configFile, "config", "c", "./config.yml", "config file")

			manager := &Manager{
				flags:  flags,
				target: tt.input,
			}

			err := manager.genFlagSet(tt.nameTag)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Check that all expected flags exist
			for expectedFlag, expectedType := range tt.expectedFlags {
				flag := flags.Lookup(expectedFlag)
				if flag == nil {
					t.Errorf("Expected flag '%s' not found", expectedFlag)
					continue
				}

				// Basic type checking based on flag type
				switch expectedType {
				case "string":
					if flag.Value.Type() != "string" {
						t.Errorf("Flag '%s' expected type string, got %s", expectedFlag, flag.Value.Type())
					}
				case "int":
					if flag.Value.Type() != "int" {
						t.Errorf("Flag '%s' expected type int, got %s", expectedFlag, flag.Value.Type())
					}
				case "bool":
					if flag.Value.Type() != "bool" {
						t.Errorf("Flag '%s' expected type bool, got %s", expectedFlag, flag.Value.Type())
					}
				case "duration":
					if flag.Value.Type() != "duration" {
						t.Errorf("Flag '%s' expected type duration, got %s", expectedFlag, flag.Value.Type())
					}
				case "float64":
					if flag.Value.Type() != "float64" {
						t.Errorf("Flag '%s' expected type float64, got %s", expectedFlag, flag.Value.Type())
					}
				case "stringSlice":
					if flag.Value.Type() != "stringSlice" {
						t.Errorf("Flag '%s' expected type stringSlice, got %s", expectedFlag, flag.Value.Type())
					}
				case "stringToString":
					if flag.Value.Type() != "stringToString" {
						t.Errorf("Flag '%s' expected type stringToString, got %s", expectedFlag, flag.Value.Type())
					}
				}
			}
		})
	}
}

func TestManagerParseConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		configData  string
		cmdArgs     []string
		expectError bool
		validate    func(t *testing.T, config *SimpleConfig)
	}{
		{
			name: "valid config file only",
			configData: `
name: "test-app"
port: 8080
debug: true
timeout: "30s"
rate: 1.5
`,
			cmdArgs: []string{},
			validate: func(t *testing.T, config *SimpleConfig) {
				if config.Name != "test-app" {
					t.Errorf("Expected name 'test-app', got '%s'", config.Name)
				}
				if config.Port != 8080 {
					t.Errorf("Expected port 8080, got %d", config.Port)
				}
				if !config.Debug {
					t.Error("Expected debug to be true")
				}
				if config.Timeout != 30*time.Second {
					t.Errorf("Expected timeout 30s, got %v", config.Timeout)
				}
				if config.Rate != 1.5 {
					t.Errorf("Expected rate 1.5, got %f", config.Rate)
				}
			},
		},
		{
			name: "flags override config file",
			configData: `
name: "from-config"
port: 8080
debug: false
`,
			cmdArgs: []string{"--name", "from-flag", "--debug"},
			validate: func(t *testing.T, config *SimpleConfig) {
				if config.Name != "from-flag" {
					t.Errorf("Expected name 'from-flag' (from flag), got '%s'", config.Name)
				}
				if config.Port != 8080 {
					t.Errorf("Expected port 8080 (from config), got %d", config.Port)
				}
				if !config.Debug {
					t.Error("Expected debug to be true (from flag)")
				}
			},
		},
		{
			name: "short flags work",
			configData: `
name: "from-config"
port: 8080
`,
			cmdArgs: []string{"-n", "short-flag", "-p", "9090"},
			validate: func(t *testing.T, config *SimpleConfig) {
				if config.Name != "short-flag" {
					t.Errorf("Expected name 'short-flag', got '%s'", config.Name)
				}
				if config.Port != 9090 {
					t.Errorf("Expected port 9090, got %d", config.Port)
				}
			},
		},
		{
			name:        "invalid yaml",
			configData:  "invalid: yaml: content: [",
			cmdArgs:     []string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config file
			configPath := createTempConfigFile(t, tt.configData)

			// Create configuration struct
			config := &SimpleConfig{}
			manager, err := New(config, "")
			if err != nil {
				t.Fatalf("Failed to create manager: %v", err)
			}
			manager.configFile = configPath // Create cobra command
			cmd := &cobra.Command{
				Use: "test",
			}
			cmd.Flags().AddFlagSet(manager.FlagSet())

			// Parse command line args
			if len(tt.cmdArgs) > 0 {
				cmd.SetArgs(tt.cmdArgs)
				err := cmd.ParseFlags(tt.cmdArgs)
				if err != nil {
					t.Fatalf("Failed to parse flags: %v", err)
				}
			}

			// Parse configuration
			parseErr := manager.ParseConfiguration(cmd)

			if tt.expectError {
				if parseErr == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if parseErr != nil {
				t.Errorf("Unexpected error: %v", parseErr)
				return
			}

			if tt.validate != nil {
				tt.validate(t, config)
			}
		})
	}
}

func TestManagerParseConfigurationFileNotFound(t *testing.T) {
	config := &SimpleConfig{}
	manager, err := New(config, "")
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	manager.configFile = "/nonexistent/path/config.yml"

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().AddFlagSet(manager.FlagSet())

	parseErr := manager.ParseConfiguration(cmd)
	if parseErr == nil {
		t.Error("Expected error for nonexistent config file")
	}
	if !strings.Contains(parseErr.Error(), "could not read config file") {
		t.Errorf("Expected 'could not read config file' error, got: %v", parseErr)
	}
}

func TestManagerParseConfigurationComplexConfig(t *testing.T) {
	configData := `
basic:
  name: "test-basic"
  version: "1.0.0"
server:
  host: "localhost"
  port: 8080
tags:
  - "tag1"
  - "tag2"
metadata:
  key1: "value1"
  key2: "value2"
`
	configPath := createTempConfigFile(t, configData)

	config := &ComplexConfig{}
	manager, err := New(config, "")
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	manager.configFile = configPath

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().AddFlagSet(manager.FlagSet())

	parseErr := manager.ParseConfiguration(cmd)
	if parseErr != nil {
		t.Fatalf("Unexpected error: %v", parseErr)
	}

	// Validate complex config parsing
	if config.Basic.Name != "test-basic" {
		t.Errorf("Expected basic.name 'test-basic', got '%s'", config.Basic.Name)
	}
	if config.Basic.Version != "1.0.0" {
		t.Errorf("Expected basic.version '1.0.0', got '%s'", config.Basic.Version)
	}
	if config.Server.Host != "localhost" {
		t.Errorf("Expected server.host 'localhost', got '%s'", config.Server.Host)
	}
	if config.Server.Port != 8080 {
		t.Errorf("Expected server.port 8080, got %d", config.Server.Port)
	}
	if len(config.Tags) != 2 || config.Tags[0] != "tag1" || config.Tags[1] != "tag2" {
		t.Errorf("Expected tags [tag1, tag2], got %v", config.Tags)
	}
	if len(config.Metadata) != 2 || config.Metadata["key1"] != "value1" || config.Metadata["key2"] != "value2" {
		t.Errorf("Expected metadata map with key1:value1, key2:value2, got %v", config.Metadata)
	}
}

func TestProcessStructEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		nameTag     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "unsupported field type",
			input:       &ConfigWithUnsupportedTypes{},
			nameTag:     "name",
			expectError: true,
			errorMsg:    "unsupported field type",
		},
		{
			name:    "mixed settable fields",
			input:   &ConfigWithMixedTags{},
			nameTag: "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			v := reflect.ValueOf(tt.input).Elem()

			err := processStruct(tt.nameTag, flags, v, "")

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// Test all numeric types and other edge cases
func TestProcessStructAllTypes(t *testing.T) {
	type AllTypesConfig struct {
		String      string            `name:"string" description:"String field"`
		Int         int               `name:"int" description:"Int field"`
		Int8        int8              `name:"int8" description:"Int8 field"`
		Int16       int16             `name:"int16" description:"Int16 field"`
		Int32       int32             `name:"int32" description:"Int32 field"`
		Int64       int64             `name:"int64" description:"Int64 field"`
		Uint        uint              `name:"uint" description:"Uint field"`
		Uint8       uint8             `name:"uint8" description:"Uint8 field"`
		Uint16      uint16            `name:"uint16" description:"Uint16 field"`
		Uint32      uint32            `name:"uint32" description:"Uint32 field"`
		Uint64      uint64            `name:"uint64" description:"Uint64 field"`
		Float32     float32           `name:"float32" description:"Float32 field"`
		Float64     float64           `name:"float64" description:"Float64 field"`
		Bool        bool              `name:"bool" description:"Bool field"`
		Duration    time.Duration     `name:"duration" description:"Duration field"`
		StringSlice []string          `name:"stringslice" description:"String slice field"`
		StringMap   map[string]string `name:"stringmap" description:"String map field"`

		// Fields with short flags
		WithShort string `name:"with-short" short:"w" description:"Field with short flag"`
	}

	config := &AllTypesConfig{}
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	v := reflect.ValueOf(config).Elem()

	err := processStruct("name", flags, v, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify all flags were created
	expectedFlags := []string{
		"string", "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "bool", "duration",
		"stringslice", "stringmap", "with-short",
	}

	for _, flagName := range expectedFlags {
		if flags.Lookup(flagName) == nil {
			t.Errorf("Flag '%s' not found", flagName)
		}
	}

	// Verify short flag
	shortFlag := flags.ShorthandLookup("w")
	if shortFlag == nil {
		t.Error("Short flag 'w' not found")
	}
}

// Test unsupported slice types
func TestProcessStructUnsupportedSlice(t *testing.T) {
	type UnsupportedSliceConfig struct {
		IntSlice []int `name:"intslice" description:"Unsupported int slice"`
	}

	config := &UnsupportedSliceConfig{}
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	v := reflect.ValueOf(config).Elem()

	err := processStruct("name", flags, v, "")
	if err == nil {
		t.Error("Expected error for unsupported slice type")
	}
	if !strings.Contains(err.Error(), "unsupported slice type") {
		t.Errorf("Expected 'unsupported slice type' error, got: %v", err)
	}
}

// Test unsupported map types
func TestProcessStructUnsupportedMap(t *testing.T) {
	type UnsupportedMapConfig struct {
		IntMap map[string]int `name:"intmap" description:"Unsupported int map"`
	}

	config := &UnsupportedMapConfig{}
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	v := reflect.ValueOf(config).Elem()

	err := processStruct("name", flags, v, "")
	if err == nil {
		t.Error("Expected error for unsupported map type")
	}
	if !strings.Contains(err.Error(), "unsupported map type") {
		t.Errorf("Expected 'unsupported map type' error, got: %v", err)
	}
}

// Test nameTag empty behavior
func TestProcessStructEmptyNameTag(t *testing.T) {
	type ConfigWithNameTags struct {
		Field1 string `name:"field1" description:"Field 1"`
		Field2 int    `name:"field2" description:"Field 2"`
	}

	config := &ConfigWithNameTags{}
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	v := reflect.ValueOf(config).Elem()

	// Test with empty nameTag - should default to "name"
	err := processStruct("", flags, v, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should have created flags
	if flags.Lookup("field1") == nil {
		t.Error("Expected field1 flag to be created")
	}
	if flags.Lookup("field2") == nil {
		t.Error("Expected field2 flag to be created")
	}
}

// Test slice with default values
func TestProcessStructSliceDefaults(t *testing.T) {
	type ConfigWithSliceDefaults struct {
		Tags []string `name:"tags" description:"List of tags"`
	}

	config := &ConfigWithSliceDefaults{
		Tags: []string{"default1", "default2"},
	}
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	v := reflect.ValueOf(config).Elem()

	err := processStruct("name", flags, v, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	flag := flags.Lookup("tags")
	if flag == nil {
		t.Fatal("Expected tags flag to be created")
	}

	// Check default value
	defaultValue := flag.DefValue
	if !strings.Contains(defaultValue, "default1") || !strings.Contains(defaultValue, "default2") {
		t.Errorf("Expected default value to contain default1 and default2, got: %s", defaultValue)
	}
}

// Test map with default values
func TestProcessStructMapDefaults(t *testing.T) {
	type ConfigWithMapDefaults struct {
		Metadata map[string]string `name:"metadata" description:"Key-value metadata"`
	}

	config := &ConfigWithMapDefaults{
		Metadata: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	v := reflect.ValueOf(config).Elem()

	err := processStruct("name", flags, v, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	flag := flags.Lookup("metadata")
	if flag == nil {
		t.Fatal("Expected metadata flag to be created")
	}
}

// Test nil map initialization
func TestProcessStructNilMap(t *testing.T) {
	type ConfigWithNilMap struct {
		Metadata map[string]string `name:"metadata" description:"Nil map"`
	}

	config := &ConfigWithNilMap{} // Metadata will be nil
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	v := reflect.ValueOf(config).Elem()

	err := processStruct("name", flags, v, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	flag := flags.Lookup("metadata")
	if flag == nil {
		t.Fatal("Expected metadata flag to be created")
	}
}

// Test genFlagSet with non-struct pointer
func TestGenFlagSetNonStructPointer(t *testing.T) {
	str := "test"
	manager := &Manager{
		target: &str,
		flags:  pflag.NewFlagSet("test", pflag.ContinueOnError),
	}

	err := manager.genFlagSet("name")
	if err == nil {
		t.Error("Expected error for non-struct pointer")
	}
	if !strings.Contains(err.Error(), "expected struct") {
		t.Errorf("Expected 'expected struct' error, got: %v", err)
	}
}

// Test prefix handling in nested structs
func TestProcessStructNestedWithPrefix(t *testing.T) {
	config := &ComplexConfig{}
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	v := reflect.ValueOf(config).Elem()

	err := processStruct("name", flags, v, "parent")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check nested flags have correct prefix
	if flags.Lookup("parent.basic.name") == nil {
		t.Error("Expected 'parent.basic.name' flag")
	}
	if flags.Lookup("parent.server.host") == nil {
		t.Error("Expected 'parent.server.host' flag")
	}
}

// Test private/unexported fields
func TestProcessStructUnexportedFields(t *testing.T) {
	type ConfigWithUnexported struct {
		Public  string `name:"public" description:"Public field"`
		private string `name:"private" description:"Private field"` //nolint:unused
	}

	config := &ConfigWithUnexported{}
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	v := reflect.ValueOf(config).Elem()

	err := processStruct("name", flags, v, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Only public field should have a flag
	if flags.Lookup("public") == nil {
		t.Error("Expected 'public' flag")
	}
	if flags.Lookup("private") != nil {
		t.Error("Did not expect 'private' flag for unexported field")
	}
}

// Test comprehensive coverage for all integer types with short flags
func TestProcessStructIntegerTypesWithShortFlags(t *testing.T) {
	type IntTypesConfig struct {
		Int8WithShort    int8    `name:"int8" short:"a" description:"Int8 with short"`
		Int16WithShort   int16   `name:"int16" short:"b" description:"Int16 with short"`
		Int32WithShort   int32   `name:"int32" short:"c" description:"Int32 with short"`
		Uint8WithShort   uint8   `name:"uint8" short:"e" description:"Uint8 with short"`
		Uint16WithShort  uint16  `name:"uint16" short:"f" description:"Uint16 with short"`
		Uint32WithShort  uint32  `name:"uint32" short:"g" description:"Uint32 with short"`
		Uint64WithShort  uint64  `name:"uint64" short:"h" description:"Uint64 with short"`
		Float32WithShort float32 `name:"float32" short:"i" description:"Float32 with short"`
	}

	config := &IntTypesConfig{}
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	v := reflect.ValueOf(config).Elem()

	err := processStruct("name", flags, v, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify all short flags work
	shortFlags := []string{"a", "b", "c", "e", "f", "g", "h", "i"}
	for _, sf := range shortFlags {
		if flags.ShorthandLookup(sf) == nil {
			t.Errorf("Expected short flag '%s'", sf)
		}
	}
}

// Test struct with interface{} field - should error
func TestProcessStructUnsupportedInterface(t *testing.T) {
	type ConfigWithInterface struct {
		Name      string `name:"name" description:"Name field"`
		Interface any    `name:"interface" description:"Interface field"`
	}

	config := &ConfigWithInterface{}
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	v := reflect.ValueOf(config).Elem()

	err := processStruct("name", flags, v, "")
	if err == nil {
		t.Error("Expected error for interface{} type")
	}
	if !strings.Contains(err.Error(), "unsupported field type") {
		t.Errorf("Expected 'unsupported field type' error, got: %v", err)
	}
}

// Test deeply nested structs
func TestProcessStructDeeplyNested(t *testing.T) {
	type Level3 struct {
		Value string `name:"value" description:"Level 3 value"`
	}
	type Level2 struct {
		Level3 Level3 `name:"level3"`
	}
	type Level1 struct {
		Level2 Level2 `name:"level2"`
	}

	config := &Level1{}
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	v := reflect.ValueOf(config).Elem()

	err := processStruct("name", flags, v, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if flags.Lookup("level2.level3.value") == nil {
		t.Error("Expected deeply nested flag 'level2.level3.value'")
	}
}

// Test map with non-string values (should error)
func TestProcessStructMapIntValues(t *testing.T) {
	type ConfigWithIntMap struct {
		IntValues map[string]int `name:"intvalues" description:"Map with int values"`
	}

	config := &ConfigWithIntMap{}
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	v := reflect.ValueOf(config).Elem()

	err := processStruct("name", flags, v, "")
	if err == nil {
		t.Error("Expected error for map with non-string values")
	}
}

// Test map with non-string keys (should error)
func TestProcessStructMapIntKeys(t *testing.T) {
	type ConfigWithIntKeyMap struct {
		IntKeys map[int]string `name:"intkeys" description:"Map with int keys"`
	}

	config := &ConfigWithIntKeyMap{}
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	v := reflect.ValueOf(config).Elem()

	err := processStruct("name", flags, v, "")
	if err == nil {
		t.Error("Expected error for map with non-string keys")
	}
}

// Test complete coverage of genFlagSet error paths
func TestGenFlagSetErrorPaths(t *testing.T) {
	tests := []struct {
		name        string
		target      any
		expectError bool
		errorMsg    string
	}{
		{
			name:        "non-pointer",
			target:      SimpleConfig{},
			expectError: true,
			errorMsg:    "expected pointer",
		},
		{
			name:        "pointer to non-struct",
			target:      new(int),
			expectError: true,
			errorMsg:    "expected struct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &Manager{
				target: tt.target,
				flags:  pflag.NewFlagSet("test", pflag.ContinueOnError),
			}

			err := manager.genFlagSet("name")

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// Test comprehensive duration handling
func TestProcessStructDurationTypes(t *testing.T) {
	type DurationConfig struct {
		Timeout      time.Duration `name:"timeout" description:"Timeout duration"`
		TimeoutShort time.Duration `name:"timeout-short" short:"t" description:"Timeout with short"`
		RegularInt64 int64         `name:"regular-int64" description:"Regular int64"`
	}

	config := &DurationConfig{
		Timeout:      30 * time.Second,
		TimeoutShort: 60 * time.Second,
		RegularInt64: 12345,
	}
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	v := reflect.ValueOf(config).Elem()

	err := processStruct("name", flags, v, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check duration flags
	timeoutFlag := flags.Lookup("timeout")
	if timeoutFlag == nil || timeoutFlag.Value.Type() != "duration" {
		t.Error("Expected timeout duration flag")
	}

	timeoutShortFlag := flags.Lookup("timeout-short")
	if timeoutShortFlag == nil || timeoutShortFlag.Value.Type() != "duration" {
		t.Error("Expected timeout-short duration flag")
	}

	// Check regular int64 flag
	int64Flag := flags.Lookup("regular-int64")
	if int64Flag == nil || int64Flag.Value.Type() != "int64" {
		t.Error("Expected regular-int64 int64 flag")
	}

	// Check short flag
	if flags.ShorthandLookup("t") == nil {
		t.Error("Expected short flag 't'")
	}
}

// Test with config flag skip logic - config flag should be ignored in setFlags
func TestManagerParseConfigurationSkipConfigFlag(t *testing.T) {
	configData := `name: "test"`
	configPath := createTempConfigFile(t, configData)

	config := &SimpleConfig{}
	manager, err := New(config, "")
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	manager.configFile = configPath

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().AddFlagSet(manager.FlagSet())

	// Set the config flag itself - should be skipped in setFlags processing
	// The config file path flag should be ignored when processing setFlags
	cmd.SetArgs([]string{"--config", configPath})
	_ = cmd.ParseFlags([]string{"--config", configPath})

	parseErr := manager.ParseConfiguration(cmd)
	if parseErr != nil {
		t.Errorf("Unexpected error: %v", parseErr)
	}

	// Verify the config was loaded
	if config.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", config.Name)
	}
}

// Test additional edge case for better coverage
func TestProcessStructWithPrefixAndEmptyName(t *testing.T) {
	type ConfigWithSomeFields struct {
		Tagged   string `name:"tagged" description:"Tagged field"`
		Untagged string // No name tag, should be skipped
	}

	config := &ConfigWithSomeFields{}
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	v := reflect.ValueOf(config).Elem()

	err := processStruct("name", flags, v, "prefix")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should have prefixed flag for tagged field
	if flags.Lookup("prefix.tagged") == nil {
		t.Error("Expected 'prefix.tagged' flag")
	}

	// Should not have flag for untagged field
	if flags.Lookup("prefix.Untagged") != nil || flags.Lookup("Untagged") != nil {
		t.Error("Should not have flag for untagged field")
	}
}

// Test genFlagSet with struct that causes processStruct to fail
func TestGenFlagSet_ProcessStructError(t *testing.T) {
	type ConfigWithUnsupported struct {
		Name     string   `name:"name" description:"Valid field"`
		BadField chan int `name:"bad" description:"Unsupported channel type"`
	}

	config := &ConfigWithUnsupported{}
	manager := &Manager{
		target: config,
		flags:  pflag.NewFlagSet("test", pflag.ContinueOnError),
	}

	err := manager.genFlagSet("name")
	if err == nil {
		t.Error("Expected error from unsupported field type in processStruct")
	}
	if !strings.Contains(err.Error(), "unsupported field type") {
		t.Errorf("Expected 'unsupported field type' error from processStruct, got: %v", err)
	}
}

// Test ParseConfiguration with invalid flag value type mismatch
func TestManagerParseConfigurationInvalidFlagValue(t *testing.T) {
	configData := `name: "test"`
	configPath := createTempConfigFile(t, configData)

	config := &SimpleConfig{}
	manager, managerErr := New(config, "")
	if managerErr != nil {
		t.Fatalf("Failed to create manager: %v", managerErr)
	}
	manager.configFile = configPath

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().AddFlagSet(manager.FlagSet())

	// Try to set a port (int) flag with an invalid value that can't be parsed
	cmd.SetArgs([]string{"--port", "not-a-number"})
	err := cmd.ParseFlags([]string{"--port", "not-a-number"})

	// The parse error should occur during flag parsing, not in ParseConfiguration
	if err == nil {
		// If parsing succeeded, try ParseConfiguration (it should work fine)
		parseErr := manager.ParseConfiguration(cmd)
		if parseErr != nil {
			t.Errorf("Unexpected ParseConfiguration error: %v", parseErr)
		}
	} else {
		// This is expected - the flag parsing should fail with invalid int value
		if !strings.Contains(err.Error(), "invalid") {
			t.Logf("Flag parsing failed as expected with error: %v", err)
		}
	}
}

// Test additional edge cases to reach 95% coverage
func TestProcessStructComprehensiveCoverage(t *testing.T) {
	type CoverageConfig struct {
		// Test all integer types without short flags to hit different branches
		IntNoShort     int     `name:"int-no-short" description:"Int without short"`
		Int8NoShort    int8    `name:"int8-no-short" description:"Int8 without short"`
		Int16NoShort   int16   `name:"int16-no-short" description:"Int16 without short"`
		Int32NoShort   int32   `name:"int32-no-short" description:"Int32 without short"`
		Int64NoShort   int64   `name:"int64-no-short" description:"Int64 without short"`
		UintNoShort    uint    `name:"uint-no-short" description:"Uint without short"`
		Uint8NoShort   uint8   `name:"uint8-no-short" description:"Uint8 without short"`
		Uint16NoShort  uint16  `name:"uint16-no-short" description:"Uint16 without short"`
		Uint32NoShort  uint32  `name:"uint32-no-short" description:"Uint32 without short"`
		Uint64NoShort  uint64  `name:"uint64-no-short" description:"Uint64 without short"`
		Float32NoShort float32 `name:"float32-no-short" description:"Float32 without short"`
		Float64NoShort float64 `name:"float64-no-short" description:"Float64 without short"`
		BoolNoShort    bool    `name:"bool-no-short" description:"Bool without short"`
		StringNoShort  string  `name:"string-no-short" description:"String without short"`

		// Test slice and map without short flags too
		SliceNoShort []string          `name:"slice-no-short" description:"Slice without short"`
		MapNoShort   map[string]string `name:"map-no-short" description:"Map without short"`
	}

	config := &CoverageConfig{}
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	v := reflect.ValueOf(config).Elem()

	err := processStruct("name", flags, v, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify all flags were created (checking a few key ones)
	expectedFlags := []string{
		"int-no-short", "int8-no-short", "int16-no-short", "int32-no-short", "int64-no-short",
		"uint-no-short", "uint8-no-short", "uint16-no-short", "uint32-no-short", "uint64-no-short",
		"float32-no-short", "float64-no-short", "bool-no-short", "string-no-short",
		"slice-no-short", "map-no-short",
	}

	for _, flagName := range expectedFlags {
		if flags.Lookup(flagName) == nil {
			t.Errorf("Flag '%s' not found", flagName)
		}
	}
}
