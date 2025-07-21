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
		index      int
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
	 * Read the type info for this struct
	 */
	t := reflect.TypeOf(config).Elem()
	containers = make([]any, t.NumField())

	/*
	 * First setup each field of the config struct. These are stored in "containers".
	 * Each container knows the field type, value, env name, flag name, and adds
	 * to the provided flag set.
	 */
	for index = 0; index < t.NumField(); index++ {
		containers[index], _ = container.New(config, index, envFile)
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
	for index = 0; index < t.NumField(); index++ {
		c := containers[index]

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
		}
	}
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
