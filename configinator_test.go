package configinator

import (
	"flag"
	"os"
	"testing"
	"time"
)

type TestConfig struct {
	Host      string    `flag:"host" env:"HOST" default:"localhost:8080" description:"Host and port to bind to"`
	Port      int       `flag:"port" env:"PORT" default:"9000" description:"Port number"`
	Debug     bool      `flag:"debug" env:"DEBUG" default:"false" description:"Enable debug mode"`
	Timeout   float64   `flag:"timeout" env:"TIMEOUT" default:"30.5" description:"Timeout in seconds"`
	StartTime time.Time `flag:"start" env:"START_TIME" default:"2023-01-01T00:00:00Z" description:"Start time"`
}

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func resetEnv() {
	os.Unsetenv("HOST")
	os.Unsetenv("PORT")
	os.Unsetenv("DEBUG")
	os.Unsetenv("TIMEOUT")
	os.Unsetenv("START_TIME")
}

func fixFlags() func() {
	oldArgs := os.Args
	os.Args = []string{"test"}

	return func() {
		os.Args = oldArgs
	}
}

func TestBeholdDefaultValues(t *testing.T) {
	resetFlags()
	resetEnv()

	// Temporarily remove os.Args to prevent flag conflicts
	putOldFlagsBack := fixFlags()
	defer putOldFlagsBack()

	config := &TestConfig{}
	Behold(config)

	if config.Host != "localhost:8080" {
		t.Errorf("Expected Host to be 'localhost:8080', got '%s'", config.Host)
	}

	if config.Port != 9000 {
		t.Errorf("Expected Port to be 9000, got %d", config.Port)
	}

	if config.Debug != false {
		t.Errorf("Expected Debug to be false, got %t", config.Debug)
	}

	if config.Timeout != 30.5 {
		t.Errorf("Expected Timeout to be 30.5, got %f", config.Timeout)
	}

	expectedTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")

	if !config.StartTime.Equal(expectedTime) {
		t.Errorf("Expected StartTime to be %v, got %v", expectedTime, config.StartTime)
	}
}

func TestBeholdEnvironmentVariables(t *testing.T) {
	resetFlags()
	resetEnv()

	os.Setenv("HOST", "example.com:3000")
	os.Setenv("PORT", "8080")
	os.Setenv("DEBUG", "true")
	os.Setenv("TIMEOUT", "45.7")
	os.Setenv("START_TIME", "2024-01-01T12:00:00Z")

	defer resetEnv()

	// Temporarily set os.Args to prevent flag conflicts
	putOldFlagsBack := fixFlags()
	defer putOldFlagsBack()

	config := &TestConfig{}
	Behold(config)

	if config.Host != "example.com:3000" {
		t.Errorf("Expected Host to be 'example.com:3000', got '%s'", config.Host)
	}

	if config.Port != 8080 {
		t.Errorf("Expected Port to be 8080, got %d", config.Port)
	}

	if config.Debug != true {
		t.Errorf("Expected Debug to be true, got %t", config.Debug)
	}

	if config.Timeout != 45.7 {
		t.Errorf("Expected Timeout to be 45.7, got %f", config.Timeout)
	}

	expectedTime, _ := time.Parse(time.RFC3339, "2024-01-01T12:00:00Z")

	if !config.StartTime.Equal(expectedTime) {
		t.Errorf("Expected StartTime to be %v, got %v", expectedTime, config.StartTime)
	}
}

func TestBeholdEnvFile(t *testing.T) {
	resetFlags()
	resetEnv()

	// Create a temporary .env file
	envContent := `HOST=envfile.com:5000
PORT=7777
DEBUG=true
TIMEOUT=60.0
START_TIME=2025-01-01T00:00:00Z`

	err := os.WriteFile(".env", []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}
	defer os.Remove(".env")

	// Temporarily set os.Args to prevent flag conflicts
	putOldFlagsBack := fixFlags()
	defer putOldFlagsBack()

	config := &TestConfig{}
	Behold(config)

	if config.Host != "envfile.com:5000" {
		t.Errorf("Expected Host to be 'envfile.com:5000', got '%s'", config.Host)
	}

	if config.Port != 7777 {
		t.Errorf("Expected Port to be 7777, got %d", config.Port)
	}

	if config.Debug != true {
		t.Errorf("Expected Debug to be true, got %t", config.Debug)
	}

	if config.Timeout != 60.0 {
		t.Errorf("Expected Timeout to be 60.0, got %f", config.Timeout)
	}

	expectedTime, _ := time.Parse(time.RFC3339, "2025-01-01T00:00:00Z")

	if !config.StartTime.Equal(expectedTime) {
		t.Errorf("Expected StartTime to be %v, got %v", expectedTime, config.StartTime)
	}
}

func TestBeholdCommandLineFlags(t *testing.T) {
	resetFlags()
	resetEnv()

	// Simulate command line arguments
	oldArgs := os.Args
	os.Args = []string{"test", "-host=flag.com:4000", "-port=6666", "-debug=true", "-timeout=75.5", "-start=2026-01-01T00:00:00Z"}
	defer func() { os.Args = oldArgs }()

	config := &TestConfig{}
	Behold(config)

	if config.Host != "flag.com:4000" {
		t.Errorf("Expected Host to be 'flag.com:4000', got '%s'", config.Host)
	}
	if config.Port != 6666 {
		t.Errorf("Expected Port to be 6666, got %d", config.Port)
	}
	if config.Debug != true {
		t.Errorf("Expected Debug to be true, got %t", config.Debug)
	}
	if config.Timeout != 75.5 {
		t.Errorf("Expected Timeout to be 75.5, got %f", config.Timeout)
	}
	expectedTime, _ := time.Parse(time.RFC3339, "2026-01-01T00:00:00Z")
	if !config.StartTime.Equal(expectedTime) {
		t.Errorf("Expected StartTime to be %v, got %v", expectedTime, config.StartTime)
	}
}

func TestBeholdPrecedenceOrder(t *testing.T) {
	resetFlags()
	resetEnv()

	// Set all sources with different values
	// Environment variable
	os.Setenv("HOST", "env.com:2000")
	defer resetEnv()

	// Create .env file
	envContent := `HOST=envfile.com:3000`
	err := os.WriteFile(".env", []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}
	defer os.Remove(".env")

	// Command line flag (highest precedence)
	oldArgs := os.Args
	os.Args = []string{"test", "-host=flag.com:4000"}
	defer func() { os.Args = oldArgs }()

	config := &TestConfig{}
	Behold(config)

	// Flag should win (highest precedence)
	if config.Host != "flag.com:4000" {
		t.Errorf("Expected Host to be 'flag.com:4000' (flag precedence), got '%s'", config.Host)
	}
}

func TestBeholdPrecedenceWithoutFlag(t *testing.T) {
	resetFlags()
	resetEnv()

	// Set environment variable and .env file
	os.Setenv("HOST", "env.com:2000")
	defer resetEnv()

	envContent := `HOST=envfile.com:3000`
	err := os.WriteFile(".env", []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}
	defer os.Remove(".env")

	// No command line flag
	putOldFlagsBack := fixFlags()
	defer putOldFlagsBack()

	config := &TestConfig{}
	Behold(config)

	// .env file should win over environment variable
	if config.Host != "envfile.com:3000" {
		t.Errorf("Expected Host to be 'envfile.com:3000' (.env precedence), got '%s'", config.Host)
	}
}

func TestBeholdPrecedenceEnvOnly(t *testing.T) {
	resetFlags()
	resetEnv()

	// Only set environment variable
	os.Setenv("HOST", "env.com:2000")
	defer resetEnv()

	// No .env file, no flags
	putOldFlagsBack := fixFlags()
	defer putOldFlagsBack()

	config := &TestConfig{}
	Behold(config)

	// Environment variable should win over default
	if config.Host != "env.com:2000" {
		t.Errorf("Expected Host to be 'env.com:2000' (env precedence), got '%s'", config.Host)
	}
}

func TestBeholdAllDataTypes(t *testing.T) {
	resetFlags()
	resetEnv()

	type AllTypesConfig struct {
		StringVal string    `flag:"str" default:"test"`
		IntVal    int       `flag:"int" default:"42"`
		BoolVal   bool      `flag:"bool" default:"true"`
		FloatVal  float64   `flag:"float" default:"3.14"`
		TimeVal   time.Time `flag:"time" default:"2023-01-01T00:00:00Z"`
	}

	// Temporarily set os.Args to prevent flag conflicts
	putOldFlagsBack := fixFlags()
	defer putOldFlagsBack()

	config := &AllTypesConfig{}
	Behold(config)

	if config.StringVal != "test" {
		t.Errorf("Expected StringVal to be 'test', got '%s'", config.StringVal)
	}

	if config.IntVal != 42 {
		t.Errorf("Expected IntVal to be 42, got %d", config.IntVal)
	}

	if config.BoolVal != true {
		t.Errorf("Expected BoolVal to be true, got %t", config.BoolVal)
	}

	if config.FloatVal != 3.14 {
		t.Errorf("Expected FloatVal to be 3.14, got %f", config.FloatVal)
	}

	expectedTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")

	if !config.TimeVal.Equal(expectedTime) {
		t.Errorf("Expected TimeVal to be %v, got %v", expectedTime, config.TimeVal)
	}
}
