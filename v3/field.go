package env

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/simia-tech/env/v3/internal/parser"
)

var (
	ErrMissingValue = errors.New("missing value")
	ErrInvalidValue = errors.New("invalid value")
)

type FieldType interface {
	bool | int | string | []string
}

type FieldValue[T FieldType] struct {
	name         string
	location     string
	defaultValue T
	options      options
}

var nameRegexp = regexp.MustCompile("^[A-Z0-9_]+$")

func Field[T FieldType](name string, defaultValue T, opts ...Option) *FieldValue[T] {
	if !nameRegexp.MatchString(name) {
		panic(fmt.Sprintf("field name [%s] must only contain capital letters, numbers or underscores", name))
	}
	_, filename, line, _ := runtime.Caller(1)

	f := &FieldValue[T]{
		name:         name,
		location:     fmt.Sprintf("%s:%d", filename, line),
		defaultValue: defaultValue,
		options:      newOptions(opts),
	}
	registerField(f)
	return f
}

func (f *FieldValue[T]) Name() string {
	return f.name
}

func (f *FieldValue[T]) Description() string {
	if f.options.description != "" {
		return f.options.description
	}
	sentences := []string{f.label() + " field."}
	if f.options.required {
		sentences = append(sentences, "Required field.")
	}
	if f.options.allowedValues != nil {
		sentences = append(sentences, fmt.Sprintf("Allowed values are %s.", joinStringValues(f.options.allowedValues)))
	}
	sentences = append(sentences, "The default value is '"+formatValue(f.defaultValue)+"'.")
	sentences = append(sentences, "Defined at "+f.location+".")
	return strings.Join(sentences, " ")
}

func (f *FieldValue[T]) GetRaw() (string, error) {
	text, ok := os.LookupEnv(f.name)
	if !ok {
		if f.options.required {
			return formatValue(f.defaultValue), fmt.Errorf("field [%s]: %w", f.name, ErrMissingValue)
		}
		return formatValue(f.defaultValue), nil
	}
	text = strings.TrimSpace(text)

	if !f.options.isAllowedValue(text) {
		return formatValue(f.defaultValue), fmt.Errorf("field [%s]: value [%s]: %w", f.name, text, ErrInvalidValue)
	}

	return text, nil
}

func (f *FieldValue[T]) GetRawOrDefault() string {
	value, _ := f.GetRaw()
	return value
}

func (f *FieldValue[T]) Get() (T, error) {
	raw, err := f.GetRaw()
	if err != nil {
		return f.defaultValue, err
	}

	result, err := parseValue[T](raw)
	if err != nil {
		return f.defaultValue, fmt.Errorf("field [%s]: %w", f.name, err)
	}

	return result, nil
}

func (f *FieldValue[T]) label() string {
	switch any(f.defaultValue).(type) {
	case bool:
		return "Boolean"
	case int:
		return "Int"
	case string:
		return "String"
	case []string:
		return "Strings"
	default:
		return ""
	}
}

func parseValue[T FieldType](raw string) (T, error) {
	value := *new(T)

	result := any(nil)
	switch any(value).(type) {
	case bool:
		switch raw {
		case "1", "true", "yes":
			result = true
		case "0", "false", "no":
			result = false
		default:
			return value, fmt.Errorf("parse bool [%s]: %w", raw, ErrInvalidValue)
		}

	case int:
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return value, fmt.Errorf("parse int [%s]: %w", raw, ErrInvalidValue)
		}
		result = int(v)

	case string:
		result = raw

	case []string:
		values, err := parser.ParseStrings(raw)
		if err != nil {
			return value, fmt.Errorf("parse strings [%s]: %w", raw, err)
		}
		result = values

	}

	return result.(T), nil
}

func formatValue[T FieldType](value T) string {
	switch t := any(value).(type) {
	case bool:
		if t {
			return "true"
		}
		return "false"

	case int:
		return strconv.FormatInt(int64(t), 10)

	case string:
		return t

	case []string:
		return parser.FormatStrings(t)

	default:
		return ""
	}
}
