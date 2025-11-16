package configinator

import (
	"flag"
	"os"
	"reflect"
	"time"

	"github.com/adampresley/configinator/container"
	"github.com/adampresley/configinator/env"
)

/*
Behold initializes a provided struct with values from defaults,
environment, .env file, and flags. It does this by adding tags to your
struct. For example:

	  type Config struct {
		  Host string `flag:"host" env:"HOST" default:"localhost:8080" description:"Host and port to bind to"`
	  }

The above example will accept a command line flag of "host",
or an environment variable named "HOST". If none of the above
are provided then the value from 'default' is used.

If an .env file is found that will be read and used.
*/
func Behold(config any) {
	var (
		err        error
		containers []any
	)

	envFile := make(map[string]string)

	/*
	 * If we have an environment file, load it
	 */
	if env.FileExists(".env") {
		if envFile, err = env.ReadFile(".env"); err != nil {
			panic(err)
		}
	}

	/*
	 * Recursively discover all fields to be configured
	 */
	if containers, err = collectContainers(reflect.ValueOf(config).Elem(), envFile); err != nil {
		panic(err)
	}

	/*
	 * Parse flags
	 */
	if len(os.Args) > 1 && !flag.Parsed() {
		flag.Parse()
	}

	/*
	 * Set the values in the config struct following precedence rules.
	 * They already have default values set (precedence 1).
	 */
	for _, c := range containers {
		switch typedContainer := c.(type) {
		case container.Container[bool]:
			applyValueWithPrecedence(typedContainer)
		case container.Container[int]:
			applyValueWithPrecedence(typedContainer)
		case container.Container[float64]:
			applyValueWithPrecedence(typedContainer)
		case container.Container[string]:
			applyValueWithPrecedence(typedContainer)
		case container.Container[time.Time]:
			applyValueWithPrecedence(typedContainer)
		case container.Container[time.Duration]:
			applyValueWithPrecedence(typedContainer)
		}
	}
}

func collectContainers(configValue reflect.Value, envFile map[string]string) ([]any, error) {
	var (
		err                error
		containers         []any
		embeddedContainers []any
		c                  any
	)

	t := configValue.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := configValue.Field(i)

		if field.Anonymous && fieldValue.Kind() == reflect.Struct {
			if embeddedContainers, err = collectContainers(fieldValue, envFile); err != nil {
				return nil, err
			}

			containers = append(containers, embeddedContainers...)

		} else {
			isConfigField := field.Tag.Get("flag") != "" || field.Tag.Get("env") != "" || field.Tag.Get("default") != ""

			if c, err = container.New(field, fieldValue, envFile); err != nil {
				if isConfigField {
					return nil, err
				} else {
					continue
				}
			}

			if c != nil {
				containers = append(containers, c)
			}
		}
	}

	return containers, nil
}

func applyValueWithPrecedence[T any](c container.Container[T]) {
	// Environment variable (precedence 2)
	if value, ok := c.GetEnvValue(); ok {
		c.SetConfigValue(value)
	}

	// Environment file (precedence 3)
	if value, ok := c.GetEnvFileValue(); ok {
		c.SetConfigValue(value)
	}

	// Command line flag (highest precedence)
	if value, ok := c.GetFlagValue(); ok {
		c.SetConfigValue(value)
	}
}
