// SPDX-FileCopyrightText: Copyright 2025 Krishna Iyer (www.krishnaiyer.tech)
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

// Manager manages configuration.
type Manager struct {
	flags      *pflag.FlagSet
	target     any
	configFile string
}

// New returns a new Manager.
// Out must be a pointer, else this function panics.
func New(out any, nameTagOverride string) (*Manager, error) {
	v := reflect.TypeOf(out).Kind()
	if v != reflect.Pointer {
		panic("out is not a pointer")
	}

	m := &Manager{
		target: out,
		flags:  pflag.NewFlagSet("config", pflag.ExitOnError),
	}
	// Add the config file flag by default.
	m.flags.StringVarP(
		&m.configFile,
		"config",
		"c",
		"./config.yml",
		"location of the configuration file (default: ./config.yml)",
	)
	err := m.genFlagSet(nameTagOverride)
	return m, err
}

// ParseConfiguration parses the configuration.
// Order of precedence; config file < flag < environment.
// TODO: Support environment.
func (m Manager) ParseConfiguration(cmd *cobra.Command) (err error) {
	// Save explicitly set flag values before loading the yaml.
	setFlags := make(map[string]string)
	cmd.Flags().Visit(func(f *pflag.Flag) {
		if f.Name != "config" {
			setFlags[f.Name] = f.Value.String()
		}
	})

	// Get values from the config file.
	raw, err := os.ReadFile(m.configFile)
	if err != nil {
		return fmt.Errorf("could not read config file: %w", err)
	}
	if err := yaml.Unmarshal(raw, m.target); err != nil {
		return fmt.Errorf("could not parse config file: %w", err)
	}

	// Override explicitly set flags from the args.
	for name, value := range setFlags {
		if err := cmd.Flags().Set(name, value); err != nil {
			return fmt.Errorf("could not set flag %s: %w", name, err)
		}
	}
	return nil
}

// FlagSet returns the manager's flagset.
func (m Manager) FlagSet() *pflag.FlagSet {
	return m.flags
}

// genFlagSet reads the configuration and uses reflection to generate a corresponding flagset.
// Takes an input pointer to bind flags directly to the element.
func (m Manager) genFlagSet(nameTag string) error {
	v := reflect.ValueOf(m.target)

	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer to struct, got %s", v.Kind())
	}

	v = v.Elem()

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %s", v.Kind())
	}

	if err := processStruct(nameTag, m.flags, v, ""); err != nil {
		return err
	}

	return nil
}

// processStruct recursively processes struct fields and adds flags
func processStruct(nameTag string, fs *pflag.FlagSet, v reflect.Value, prefix string) error {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip un-settable fields
		if !fieldValue.CanSet() {
			continue
		}

		// Get the required tag values
		if nameTag == "" {
			nameTag = "name"
		}
		name := field.Tag.Get(nameTag)
		short := field.Tag.Get("short")
		description := field.Tag.Get("description")

		// Skip fields without name tag
		if name == "" {
			continue
		}

		// Add prefix if present
		fullName := name
		if prefix != "" {
			fullName = prefix + "." + name
		}

		// Handle nested structs
		if fieldValue.Kind() == reflect.Struct {
			if err := processStruct(nameTag, fs, fieldValue, fullName); err != nil {
				return err
			}
			continue
		}

		// Get pointer to the field for *Var methods
		fieldPtr := fieldValue.Addr().Interface()

		switch fieldValue.Kind() {
		case reflect.String:
			if short != "" {
				fs.StringVarP(fieldPtr.(*string), fullName, short, fieldValue.String(), description)
			} else {
				fs.StringVar(fieldPtr.(*string), fullName, fieldValue.String(), description)
			}
		case reflect.Int:
			if short != "" {
				fs.IntVarP(fieldPtr.(*int), fullName, short, int(fieldValue.Int()), description)
			} else {
				fs.IntVar(fieldPtr.(*int), fullName, int(fieldValue.Int()), description)
			}
		case reflect.Int8:
			if short != "" {
				fs.Int8VarP(fieldPtr.(*int8), fullName, short, int8(fieldValue.Int()), description)
			} else {
				fs.Int8Var(fieldPtr.(*int8), fullName, int8(fieldValue.Int()), description)
			}
		case reflect.Int16:
			if short != "" {
				fs.Int16VarP(fieldPtr.(*int16), fullName, short, int16(fieldValue.Int()), description)
			} else {
				fs.Int16Var(fieldPtr.(*int16), fullName, int16(fieldValue.Int()), description)
			}
		case reflect.Int32:
			if short != "" {
				fs.Int32VarP(fieldPtr.(*int32), fullName, short, int32(fieldValue.Int()), description)
			} else {
				fs.Int32Var(fieldPtr.(*int32), fullName, int32(fieldValue.Int()), description)
			}
		case reflect.Int64:
			// Check if this is a time.Duration (which is an int64 alias)
			if fieldValue.Type().String() == "time.Duration" {
				if short != "" {
					fs.DurationVarP(fieldPtr.(*time.Duration), fullName, short, time.Duration(fieldValue.Int()), description)
				} else {
					fs.DurationVar(fieldPtr.(*time.Duration), fullName, time.Duration(fieldValue.Int()), description)
				}
			} else {
				if short != "" {
					fs.Int64VarP(fieldPtr.(*int64), fullName, short, fieldValue.Int(), description)
				} else {
					fs.Int64Var(fieldPtr.(*int64), fullName, fieldValue.Int(), description)
				}
			}
		case reflect.Uint:
			if short != "" {
				fs.UintVarP(fieldPtr.(*uint), fullName, short, uint(fieldValue.Uint()), description)
			} else {
				fs.UintVar(fieldPtr.(*uint), fullName, uint(fieldValue.Uint()), description)
			}
		case reflect.Uint8:
			if short != "" {
				fs.Uint8VarP(fieldPtr.(*uint8), fullName, short, uint8(fieldValue.Uint()), description)
			} else {
				fs.Uint8Var(fieldPtr.(*uint8), fullName, uint8(fieldValue.Uint()), description)
			}
		case reflect.Uint16:
			if short != "" {
				fs.Uint16VarP(fieldPtr.(*uint16), fullName, short, uint16(fieldValue.Uint()), description)
			} else {
				fs.Uint16Var(fieldPtr.(*uint16), fullName, uint16(fieldValue.Uint()), description)
			}
		case reflect.Uint32:
			if short != "" {
				fs.Uint32VarP(fieldPtr.(*uint32), fullName, short, uint32(fieldValue.Uint()), description)
			} else {
				fs.Uint32Var(fieldPtr.(*uint32), fullName, uint32(fieldValue.Uint()), description)
			}
		case reflect.Uint64:
			if short != "" {
				fs.Uint64VarP(fieldPtr.(*uint64), fullName, short, fieldValue.Uint(), description)
			} else {
				fs.Uint64Var(fieldPtr.(*uint64), fullName, fieldValue.Uint(), description)
			}
		case reflect.Bool:
			if short != "" {
				fs.BoolVarP(fieldPtr.(*bool), fullName, short, fieldValue.Bool(), description)
			} else {
				fs.BoolVar(fieldPtr.(*bool), fullName, fieldValue.Bool(), description)
			}
		case reflect.Float32:
			if short != "" {
				fs.Float32VarP(fieldPtr.(*float32), fullName, short, float32(fieldValue.Float()), description)
			} else {
				fs.Float32Var(fieldPtr.(*float32), fullName, float32(fieldValue.Float()), description)
			}
		case reflect.Float64:
			if short != "" {
				fs.Float64VarP(fieldPtr.(*float64), fullName, short, fieldValue.Float(), description)
			} else {
				fs.Float64Var(fieldPtr.(*float64), fullName, fieldValue.Float(), description)
			}
		case reflect.Slice:
			switch fieldValue.Type().Elem().Kind() {
			case reflect.String:
				defaultValue := make([]string, fieldValue.Len())
				for j := 0; j < fieldValue.Len(); j++ {
					defaultValue[j] = fieldValue.Index(j).String()
				}
				if short != "" {
					fs.StringSliceVarP(fieldPtr.(*[]string), fullName, short, defaultValue, description)
				} else {
					fs.StringSliceVar(fieldPtr.(*[]string), fullName, defaultValue, description)
				}
			case reflect.Int:
				defaultValue := make([]int, fieldValue.Len())
				for j := 0; j < fieldValue.Len(); j++ {
					defaultValue[j] = int(fieldValue.Index(j).Int())
				}
				if short != "" {
					fs.IntSliceVarP(fieldPtr.(*[]int), fullName, short, defaultValue, description)
				} else {
					fs.IntSliceVar(fieldPtr.(*[]int), fullName, defaultValue, description)
				}
			default:
				return fmt.Errorf("unsupported slice type %s for field %s", fieldValue.Type(), field.Name)
			}
		case reflect.Map:
			if fieldValue.Type().Key().Kind() == reflect.String && fieldValue.Type().Elem().Kind() == reflect.String {
				defaultValue := make(map[string]string)
				if !fieldValue.IsNil() {
					for _, key := range fieldValue.MapKeys() {
						defaultValue[key.String()] = fieldValue.MapIndex(key).String()
					}
				}
				if short != "" {
					fs.StringToStringVarP(fieldPtr.(*map[string]string), fullName, short, defaultValue, description)
				} else {
					fs.StringToStringVar(fieldPtr.(*map[string]string), fullName, defaultValue, description)
				}
			} else {
				return fmt.Errorf("unsupported map type %s for field %s", fieldValue.Type(), field.Name)
			}
		default:
			return fmt.Errorf("unsupported field type %s for field %s", fieldValue.Kind(), field.Name)
		}
	}

	return nil
}
