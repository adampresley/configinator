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

type Container[T any] interface {
	GetEnvValue() (T, bool)
	GetEnvFileValue() (T, bool)
	GetFlagValue() (T, bool)
	SetConfigValue(T)
	GetDefaultValue() T
}

type baseContainer struct {
	defaultValue string
	description  string
	envFile      map[string]string
	envName      string
	field        reflect.StructField
	fieldName    string
	fieldValue   reflect.Value
	flagName     string
	hasFlag      bool
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

type StringSliceContainer struct {
	baseContainer
	flagValue *string
}

type TimeContainer struct {
	baseContainer
	flagValue *string
}

// DurationContainer implements Container[time.Duration]
type DurationContainer struct {
	baseContainer
	flagValue *string
}

func newBaseContainer(field reflect.StructField, fieldValue reflect.Value, envFile map[string]string) (baseContainer, error) {
	if !fieldValue.CanSet() {
		return baseContainer{}, ErrCantSet
	}

	flagName, hasFlag := field.Tag.Lookup(TagFlagName)

	return baseContainer{
		defaultValue: field.Tag.Get(TagDefaultValue),
		description:  field.Tag.Get(TagDescription),
		envFile:      envFile,
		envName:      field.Tag.Get(TagEnvName),
		field:        field,
		fieldName:    field.Name,
		fieldValue:   fieldValue,
		flagName:     flagName,
		hasFlag:      hasFlag,
	}, nil
}

func (c baseContainer) flagWasSet() bool {
	var result bool

	if !c.hasFlag {
		return false
	}

	flag.Visit(func(f *flag.Flag) {
		if f.Name == c.flagName {
			result = true
		}
	})

	return result
}

func NewBool(field reflect.StructField, fieldValue reflect.Value, envFile map[string]string) (Container[bool], error) {
	base, err := newBaseContainer(field, fieldValue, envFile)
	if err != nil {
		return nil, err
	}

	result := &BoolContainer{baseContainer: base}

	if !flag.Parsed() && result.hasFlag {
		result.flagValue = flag.Bool(result.flagName, result.GetDefaultValue(), result.description)
	}

	result.SetConfigValue(result.GetDefaultValue())
	return result, nil
}

func NewInt(field reflect.StructField, fieldValue reflect.Value, envFile map[string]string) (Container[int], error) {
	base, err := newBaseContainer(field, fieldValue, envFile)
	if err != nil {
		return nil, err
	}

	result := &IntContainer{baseContainer: base}

	if !flag.Parsed() && result.hasFlag {
		result.flagValue = flag.Int(result.flagName, result.GetDefaultValue(), result.description)
	}

	result.SetConfigValue(result.GetDefaultValue())
	return result, nil
}

func NewFloat64(field reflect.StructField, fieldValue reflect.Value, envFile map[string]string) (Container[float64], error) {
	base, err := newBaseContainer(field, fieldValue, envFile)
	if err != nil {
		return nil, err
	}

	result := &Float64Container{baseContainer: base}

	if !flag.Parsed() && result.hasFlag {
		result.flagValue = flag.Float64(result.flagName, result.GetDefaultValue(), result.description)
	}

	result.SetConfigValue(result.GetDefaultValue())
	return result, nil
}

func NewString(field reflect.StructField, fieldValue reflect.Value, envFile map[string]string) (Container[string], error) {
	base, err := newBaseContainer(field, fieldValue, envFile)
	if err != nil {
		return nil, err
	}

	result := &StringContainer{baseContainer: base}

	if !flag.Parsed() && result.hasFlag {
		result.flagValue = flag.String(result.flagName, result.GetDefaultValue(), result.description)
	}

	result.SetConfigValue(result.GetDefaultValue())
	return result, nil
}

func NewStringSlice(field reflect.StructField, fieldValue reflect.Value, envFile map[string]string) (Container[[]string], error) {
	var (
		err    error
		base   baseContainer
		result *StringSliceContainer
	)

	base, err = newBaseContainer(field, fieldValue, envFile)
	if err != nil {
		return nil, err
	}

	result = &StringSliceContainer{baseContainer: base}

	if !flag.Parsed() && result.hasFlag {
		result.flagValue = flag.String(result.flagName, result.defaultValue, result.description)
	}

	result.SetConfigValue(result.GetDefaultValue())
	return result, nil
}

func NewTime(field reflect.StructField, fieldValue reflect.Value, envFile map[string]string) (Container[time.Time], error) {
	base, err := newBaseContainer(field, fieldValue, envFile)
	if err != nil {
		return nil, err
	}

	result := &TimeContainer{baseContainer: base}

	if !flag.Parsed() && result.hasFlag {
		result.flagValue = flag.String(result.flagName, result.defaultValue, result.description)
	}

	result.SetConfigValue(result.GetDefaultValue())
	return result, nil
}

func NewDuration(field reflect.StructField, fieldValue reflect.Value, envFile map[string]string) (Container[time.Duration], error) {
	base, err := newBaseContainer(field, fieldValue, envFile)
	if err != nil {
		return nil, err
	}

	result := &DurationContainer{baseContainer: base}

	if !flag.Parsed() && result.hasFlag {
		result.flagValue = flag.String(result.flagName, result.defaultValue, result.description)
	}

	result.SetConfigValue(result.GetDefaultValue())
	return result, nil
}

func New(field reflect.StructField, fieldValue reflect.Value, envFile map[string]string) (any, error) {
	if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.String {
		return NewStringSlice(field, fieldValue, envFile)
	}

	fieldType := strings.ToLower(field.Type.String())

	switch fieldType {
	case "bool":
		return NewBool(field, fieldValue, envFile)
	case "int":
		return NewInt(field, fieldValue, envFile)
	case "float64":
		return NewFloat64(field, fieldValue, envFile)
	case "string":
		return NewString(field, fieldValue, envFile)
	case "time.time":
		return NewTime(field, fieldValue, envFile)
	case "time.duration":
		return NewDuration(field, fieldValue, envFile)
	default:
		return nil, nil
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
	if c.flagValue != nil && c.flagWasSet() {
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
	if c.flagValue != nil && c.flagWasSet() {
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
	if c.flagValue != nil && c.flagWasSet() {
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
	if c.flagValue != nil && c.flagWasSet() {
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

// StringSliceContainer methods
func (c *StringSliceContainer) GetEnvValue() ([]string, bool) {
	value, ok := os.LookupEnv(c.envName)
	if ok {
		return c.parseStringSlice(value), true
	}

	return []string{}, false
}

func (c *StringSliceContainer) GetEnvFileValue() ([]string, bool) {
	if value, ok := c.envFile[c.envName]; ok {
		return c.parseStringSlice(value), true
	}

	return []string{}, false
}

func (c *StringSliceContainer) GetFlagValue() ([]string, bool) {
	if c.flagValue != nil && c.flagWasSet() {
		return c.parseStringSlice(*c.flagValue), true
	}

	return []string{}, false
}

func (c *StringSliceContainer) SetConfigValue(value []string) {
	elementType := c.fieldValue.Type().Elem()
	result := reflect.MakeSlice(c.fieldValue.Type(), 0, len(value))

	for _, item := range value {
		result = reflect.Append(result, reflect.ValueOf(item).Convert(elementType))
	}

	c.fieldValue.Set(result)
}

func (c *StringSliceContainer) GetDefaultValue() []string {
	return c.parseStringSlice(c.defaultValue)
}

func (c *StringSliceContainer) parseStringSlice(value string) []string {
	var result []string

	result = make([]string, 0)

	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)

		if item != "" {
			result = append(result, item)
		}
	}

	return result
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
	if c.flagValue != nil && c.flagWasSet() {
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

// DurationContainer methods
func (c *DurationContainer) GetEnvValue() (time.Duration, bool) {
	value := os.Getenv(c.envName)
	if value != "" {
		if result, err := time.ParseDuration(value); err == nil {
			return result, true
		}
	}
	return 0, false
}

func (c *DurationContainer) GetEnvFileValue() (time.Duration, bool) {
	if value, ok := c.envFile[c.envName]; ok {
		if result, err := time.ParseDuration(value); err == nil {
			return result, true
		}
	}
	return 0, false
}

func (c *DurationContainer) GetFlagValue() (time.Duration, bool) {
	if c.flagValue != nil && c.flagWasSet() {
		if result, err := time.ParseDuration(*c.flagValue); err == nil {
			return result, true
		}
	}
	return 0, false
}

func (c *DurationContainer) SetConfigValue(value time.Duration) {
	c.fieldValue.Set(reflect.ValueOf(value))
}

func (c *DurationContainer) GetDefaultValue() time.Duration {
	if result, err := time.ParseDuration(c.defaultValue); err == nil {
		return result
	}
	return 0
}
