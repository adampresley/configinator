package container

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Supported struct tags
const (
	TagFlagName     string = "flag"
	TagEnvName      string = "env"
	TagDefaultValue string = "default"
	TagDescription  string = "description"
)

// Custom errors
var (
	ErrNoFlagName = fmt.Errorf("no flag name")
	ErrCantSet    = fmt.Errorf("can't set private fields")

	timeFormats = []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05 MST",
		"2006-01-02T15:04:05-0700",
	}
)

// Generic container interface for type-safe value resolution
type Container[T any] interface {
	GetEnvValue() (T, bool)
	GetEnvFileValue() (T, bool)
	GetFlagValue() (T, bool)
	SetConfigValue(T)
	GetDefaultValue() T
}

// Base container holds common fields and methods
type baseContainer struct {
	config       any
	configValue  reflect.Value
	defaultValue string
	description  string
	envFile      map[string]string
	envName      string
	field        reflect.StructField
	fieldName    string
	fieldValue   reflect.Value
	flagName     string
}

type BoolContainer struct {
	baseContainer
	flagValue *bool
}

type IntContainer struct {
	baseContainer
	flagValue *int
}

type Float64Container struct {
	baseContainer
	flagValue *float64
}

type StringContainer struct {
	baseContainer
	flagValue *string
}

type TimeContainer struct {
	baseContainer
	flagValue *string
}

func newBaseContainer(config any, index int, envFile map[string]string) (baseContainer, error) {
	var (
		hasFlag  bool
		flagName string
	)

	t := reflect.TypeOf(config).Elem()
	configValue := reflect.ValueOf(config).Elem()

	if !configValue.Field(index).CanSet() {
		return baseContainer{}, ErrCantSet
	}

	flagName, hasFlag = t.Field(index).Tag.Lookup(TagFlagName)

	if !hasFlag {
		return baseContainer{}, ErrNoFlagName
	}

	return baseContainer{
		config:       config,
		configValue:  configValue,
		defaultValue: t.Field(index).Tag.Get(TagDefaultValue),
		description:  t.Field(index).Tag.Get(TagDescription),
		envFile:      envFile,
		envName:      t.Field(index).Tag.Get(TagEnvName),
		field:        t.Field(index),
		fieldName:    t.Field(index).Name,
		fieldValue:   configValue.Field(index),
		flagName:     flagName,
	}, nil
}

func NewBool(config any, index int, envFile map[string]string) (Container[bool], error) {
	base, err := newBaseContainer(config, index, envFile)
	if err != nil {
		return nil, err
	}

	result := &BoolContainer{baseContainer: base}

	if !flag.Parsed() {
		result.flagValue = flag.Bool(result.flagName, result.GetDefaultValue(), result.description)
	}

	result.SetConfigValue(result.GetDefaultValue())
	return result, nil
}

func NewInt(config any, index int, envFile map[string]string) (Container[int], error) {
	base, err := newBaseContainer(config, index, envFile)
	if err != nil {
		return nil, err
	}

	result := &IntContainer{baseContainer: base}

	if !flag.Parsed() {
		result.flagValue = flag.Int(result.flagName, result.GetDefaultValue(), result.description)
	}

	result.SetConfigValue(result.GetDefaultValue())
	return result, nil
}

func NewFloat64(config any, index int, envFile map[string]string) (Container[float64], error) {
	base, err := newBaseContainer(config, index, envFile)
	if err != nil {
		return nil, err
	}

	result := &Float64Container{baseContainer: base}

	if !flag.Parsed() {
		result.flagValue = flag.Float64(result.flagName, result.GetDefaultValue(), result.description)
	}

	result.SetConfigValue(result.GetDefaultValue())
	return result, nil
}

func NewString(config any, index int, envFile map[string]string) (Container[string], error) {
	base, err := newBaseContainer(config, index, envFile)
	if err != nil {
		return nil, err
	}

	result := &StringContainer{baseContainer: base}

	if !flag.Parsed() {
		result.flagValue = flag.String(result.flagName, result.GetDefaultValue(), result.description)
	}

	result.SetConfigValue(result.GetDefaultValue())
	return result, nil
}

func NewTime(config any, index int, envFile map[string]string) (Container[time.Time], error) {
	base, err := newBaseContainer(config, index, envFile)
	if err != nil {
		return nil, err
	}

	result := &TimeContainer{baseContainer: base}

	if !flag.Parsed() {
		result.flagValue = flag.String(result.flagName, result.defaultValue, result.description)
	}

	result.SetConfigValue(result.GetDefaultValue())
	return result, nil
}

func New(config any, index int, envFile map[string]string) (any, error) {
	t := reflect.TypeOf(config).Elem()
	fieldType := strings.ToLower(t.Field(index).Type.String())

	switch fieldType {
	case "bool":
		return NewBool(config, index, envFile)
	case "int":
		return NewInt(config, index, envFile)
	case "float64":
		return NewFloat64(config, index, envFile)
	case "string":
		return NewString(config, index, envFile)
	case "time.time":
		return NewTime(config, index, envFile)
	default:
		return nil, fmt.Errorf("unsupported field type: %s", fieldType)
	}
}

// BoolContainer methods
func (c *BoolContainer) GetEnvValue() (bool, bool) {
	value := os.Getenv(c.envName)
	if value != "" {
		if result, err := strconv.ParseBool(value); err == nil {
			return result, true
		}
	}

	return false, false
}

func (c *BoolContainer) GetEnvFileValue() (bool, bool) {
	if value, ok := c.envFile[c.envName]; ok {
		if result, err := strconv.ParseBool(value); err == nil {
			return result, true
		}
	}

	return false, false
}

func (c *BoolContainer) GetFlagValue() (bool, bool) {
	if c.flagValue != nil && *c.flagValue != c.GetDefaultValue() {
		return *c.flagValue, true
	}

	return false, false
}

func (c *BoolContainer) SetConfigValue(value bool) {
	c.fieldValue.SetBool(value)
}

func (c *BoolContainer) GetDefaultValue() bool {
	if result, err := strconv.ParseBool(c.defaultValue); err == nil {
		return result
	}

	return false
}

// IntContainer methods
func (c *IntContainer) GetEnvValue() (int, bool) {
	value := os.Getenv(c.envName)
	if value != "" {
		if result, err := strconv.Atoi(value); err == nil {
			return result, true
		}
	}

	return 0, false
}

func (c *IntContainer) GetEnvFileValue() (int, bool) {
	if value, ok := c.envFile[c.envName]; ok {
		if result, err := strconv.Atoi(value); err == nil {
			return result, true
		}
	}

	return 0, false
}

func (c *IntContainer) GetFlagValue() (int, bool) {
	if c.flagValue != nil && *c.flagValue != c.GetDefaultValue() {
		return *c.flagValue, true
	}

	return 0, false
}

func (c *IntContainer) SetConfigValue(value int) {
	c.fieldValue.SetInt(int64(value))
}

func (c *IntContainer) GetDefaultValue() int {
	if result, err := strconv.Atoi(c.defaultValue); err == nil {
		return result
	}

	return 0
}

// Float64Container methods
func (c *Float64Container) GetEnvValue() (float64, bool) {
	value := os.Getenv(c.envName)
	if value != "" {
		if result, err := strconv.ParseFloat(value, 64); err == nil {
			return result, true
		}
	}

	return 0.0, false
}

func (c *Float64Container) GetEnvFileValue() (float64, bool) {
	if value, ok := c.envFile[c.envName]; ok {
		if result, err := strconv.ParseFloat(value, 64); err == nil {
			return result, true
		}
	}

	return 0.0, false
}

func (c *Float64Container) GetFlagValue() (float64, bool) {
	if c.flagValue != nil && *c.flagValue != c.GetDefaultValue() {
		return *c.flagValue, true
	}

	return 0.0, false
}

func (c *Float64Container) SetConfigValue(value float64) {
	c.fieldValue.SetFloat(value)
}

func (c *Float64Container) GetDefaultValue() float64 {
	if result, err := strconv.ParseFloat(c.defaultValue, 64); err == nil {
		return result
	}

	return 0.0
}

// StringContainer methods
func (c *StringContainer) GetEnvValue() (string, bool) {
	value := os.Getenv(c.envName)
	if value != "" {
		return value, true
	}

	return "", false
}

func (c *StringContainer) GetEnvFileValue() (string, bool) {
	if value, ok := c.envFile[c.envName]; ok {
		return value, true
	}

	return "", false
}

func (c *StringContainer) GetFlagValue() (string, bool) {
	if c.flagValue != nil && *c.flagValue != c.GetDefaultValue() {
		return *c.flagValue, true
	}

	return "", false
}

func (c *StringContainer) SetConfigValue(value string) {
	c.fieldValue.SetString(value)
}

func (c *StringContainer) GetDefaultValue() string {
	return c.defaultValue
}

// TimeContainer methods
func (c *TimeContainer) GetEnvValue() (time.Time, bool) {
	value := os.Getenv(c.envName)
	if value != "" {
		return c.parseTime(value), true
	}

	return time.Time{}, false
}

func (c *TimeContainer) GetEnvFileValue() (time.Time, bool) {
	if value, ok := c.envFile[c.envName]; ok {
		return c.parseTime(value), true
	}

	return time.Time{}, false
}

func (c *TimeContainer) GetFlagValue() (time.Time, bool) {
	if c.flagValue != nil && *c.flagValue != c.defaultValue {
		return c.parseTime(*c.flagValue), true
	}

	return time.Time{}, false
}

func (c *TimeContainer) SetConfigValue(value time.Time) {
	c.fieldValue.Set(reflect.ValueOf(value))
}

func (c *TimeContainer) GetDefaultValue() time.Time {
	if !c.isTime(c.defaultValue) {
		return time.Time{}
	}

	return c.parseTime(c.defaultValue)
}

func (c *TimeContainer) parseTime(value string) time.Time {
	for _, f := range timeFormats {
		if t, err := time.Parse(f, value); err == nil {
			return t
		}
	}

	return time.Time{}
}

func (c *TimeContainer) isTime(value string) bool {
	for _, f := range timeFormats {
		if _, err := time.Parse(f, value); err == nil {
			return true
		}
	}

	return false
}
