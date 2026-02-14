package config

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	durationType = reflect.TypeFor[time.Duration]()
	timeType     = reflect.TypeFor[time.Time]()
)

func (c *Config) Unmarshal(key string, target any) error {
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("config: unmarshal target must be a non-nil pointer to struct")
	}

	elem := rv.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("config: unmarshal target must be a pointer to struct, got pointer to %s", elem.Kind())
	}

	var values map[string]any
	if key == "" {
		values = c.values
	} else {
		sub, ok := c.find(key)
		if !ok {
			return fmt.Errorf("config: key %q not found", key)
		}
		m, ok := sub.(map[string]any)
		if !ok {
			return fmt.Errorf("config: key %q is not a map", key)
		}
		values = m
	}

	return unmarshalStruct(values, elem)
}

func unmarshalStruct(values map[string]any, rv reflect.Value) error {
	rt := rv.Type()

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fieldVal := rv.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		tag := field.Tag.Get("cfg")
		if tag == "-" {
			continue
		}
		if tag == "" {
			tag = strings.ToLower(field.Name)
		}

		val, exists := values[tag]

		if !exists || val == nil {
			if err := applyDefault(field, fieldVal); err != nil {
				return err
			}
			continue
		}

		ft := field.Type
		if ft.Kind() == reflect.Pointer {
			ft = ft.Elem()
		}

		if ft.Kind() == reflect.Struct && ft != timeType && ft != durationType {
			if err := unmarshalNestedStruct(field, fieldVal, ft, val); err != nil {
				return err
			}
			continue
		}

		converted, err := convertToType(val, field.Type, field.Tag)
		if err != nil {
			return fmt.Errorf("field %s: %w", field.Name, err)
		}
		fieldVal.Set(converted)
	}

	return nil
}

func applyDefault(field reflect.StructField, fieldVal reflect.Value) error {
	defaultStr, ok := field.Tag.Lookup("default")
	if !ok {
		return nil
	}
	parsed, err := parseStringToType(defaultStr, field.Type, field.Tag)
	if err != nil {
		return fmt.Errorf("field %s: invalid default %q: %w", field.Name, defaultStr, err)
	}
	fieldVal.Set(parsed)
	return nil
}

func unmarshalNestedStruct(field reflect.StructField, fieldVal reflect.Value, ft reflect.Type, val any) error {
	subMap, ok := val.(map[string]any)
	if !ok {
		return fmt.Errorf("field %s: expected map, got %T", field.Name, val)
	}

	if field.Type.Kind() == reflect.Pointer {
		ptr := reflect.New(ft)
		if err := unmarshalStruct(subMap, ptr.Elem()); err != nil {
			return err
		}
		fieldVal.Set(ptr)
		return nil
	}

	return unmarshalStruct(subMap, fieldVal)
}

func convertToType(val any, t reflect.Type, tag reflect.StructTag) (reflect.Value, error) {
	if t.Kind() == reflect.Pointer {
		inner, err := convertToType(val, t.Elem(), tag)
		if err != nil {
			return reflect.Value{}, err
		}
		ptr := reflect.New(t.Elem())
		ptr.Elem().Set(inner)
		return ptr, nil
	}

	if t == durationType {
		return convertToDuration(val)
	}

	if t == timeType {
		layout := tag.Get("layout")
		if layout == "" {
			layout = time.RFC3339
		}
		return convertToTime(val, layout)
	}

	switch t.Kind() {
	case reflect.String:
		return reflect.ValueOf(fmt.Sprintf("%v", val)).Convert(t), nil

	case reflect.Bool:
		b, ok := toBool(val)
		if !ok {
			return reflect.Value{}, fmt.Errorf("cannot convert %T to bool", val)
		}
		return reflect.ValueOf(b).Convert(t), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return convertToInt(val, t)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return convertToUint(val, t)

	case reflect.Float32, reflect.Float64:
		return convertToFloat(val, t)

	case reflect.Slice:
		return convertToSlice(val, t, tag)

	case reflect.Map:
		return convertToMap(val, t)

	default:
		return reflect.Value{}, fmt.Errorf("unsupported type %s", t)
	}
}

func convertToDuration(val any) (reflect.Value, error) {
	switch v := val.(type) {
	case string:
		d, err := time.ParseDuration(v)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot parse duration %q: %w", v, err)
		}
		return reflect.ValueOf(d), nil
	case int:
		return reflect.ValueOf(time.Duration(v) * time.Millisecond), nil
	case int64:
		return reflect.ValueOf(time.Duration(v) * time.Millisecond), nil
	case float64:
		return reflect.ValueOf(time.Duration(v * float64(time.Millisecond))), nil
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %T to time.Duration", val)
	}
}

func convertToTime(val any, layout string) (reflect.Value, error) {
	s, ok := val.(string)
	if !ok {
		return reflect.Value{}, fmt.Errorf("cannot convert %T to time.Time", val)
	}
	t, err := time.Parse(layout, s)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("cannot parse time %q with layout %q: %w", s, layout, err)
	}
	return reflect.ValueOf(t), nil
}

func convertToInt(val any, t reflect.Type) (reflect.Value, error) {
	var i64 int64
	switch v := val.(type) {
	case int:
		i64 = int64(v)
	case int64:
		i64 = v
	case uint64:
		if v > uint64(math.MaxInt64) {
			return reflect.Value{}, fmt.Errorf("uint64 %d overflows %s", v, t)
		}
		i64 = int64(v)
	case float64:
		i64 = int64(v)
	case bool:
		if v {
			i64 = 1
		}
	case string:
		parsed, err := strconv.ParseInt(v, 10, t.Bits())
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot parse %q as %s: %w", v, t, err)
		}
		i64 = parsed
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %T to %s", val, t)
	}

	rv := reflect.New(t).Elem()
	if rv.OverflowInt(i64) {
		return reflect.Value{}, fmt.Errorf("value %d overflows %s", i64, t)
	}
	rv.SetInt(i64)
	return rv, nil
}

func convertToUint(val any, t reflect.Type) (reflect.Value, error) {
	var u64 uint64
	switch v := val.(type) {
	case uint64:
		u64 = v
	case int:
		if v < 0 {
			return reflect.Value{}, fmt.Errorf("negative value %d for %s", v, t)
		}
		u64 = uint64(v)
	case int64:
		if v < 0 {
			return reflect.Value{}, fmt.Errorf("negative value %d for %s", v, t)
		}
		u64 = uint64(v)
	case float64:
		if v < 0 {
			return reflect.Value{}, fmt.Errorf("negative value %v for %s", v, t)
		}
		u64 = uint64(v)
	case string:
		parsed, err := strconv.ParseUint(v, 10, t.Bits())
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot parse %q as %s: %w", v, t, err)
		}
		u64 = parsed
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %T to %s", val, t)
	}

	rv := reflect.New(t).Elem()
	if rv.OverflowUint(u64) {
		return reflect.Value{}, fmt.Errorf("value %d overflows %s", u64, t)
	}
	rv.SetUint(u64)
	return rv, nil
}

func convertToFloat(val any, t reflect.Type) (reflect.Value, error) {
	var f64 float64
	switch v := val.(type) {
	case float64:
		f64 = v
	case float32:
		f64 = float64(v)
	case int:
		f64 = float64(v)
	case int64:
		f64 = float64(v)
	case uint64:
		f64 = float64(v)
	case string:
		parsed, err := strconv.ParseFloat(v, t.Bits())
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot parse %q as %s: %w", v, t, err)
		}
		f64 = parsed
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %T to %s", val, t)
	}

	rv := reflect.New(t).Elem()
	if rv.OverflowFloat(f64) {
		return reflect.Value{}, fmt.Errorf("value %v overflows %s", f64, t)
	}
	rv.SetFloat(f64)
	return rv, nil
}

func convertToSlice(val any, t reflect.Type, tag reflect.StructTag) (reflect.Value, error) {
	elemType := t.Elem()

	switch items := val.(type) {
	case []any:
		slice := reflect.MakeSlice(t, 0, len(items))
		for i, item := range items {
			converted, err := convertToType(item, elemType, tag)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("index %d: %w", i, err)
			}
			slice = reflect.Append(slice, converted)
		}
		return slice, nil

	case []string:
		slice := reflect.MakeSlice(t, 0, len(items))
		for i, item := range items {
			converted, err := convertToType(item, elemType, tag)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("index %d: %w", i, err)
			}
			slice = reflect.Append(slice, converted)
		}
		return slice, nil

	case string:
		sep := ","
		if s := tag.Get("separator"); s != "" {
			sep = s
		}
		parts := strings.Split(items, sep)
		slice := reflect.MakeSlice(t, 0, len(parts))
		for i, part := range parts {
			converted, err := convertToType(strings.TrimSpace(part), elemType, tag)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("index %d: %w", i, err)
			}
			slice = reflect.Append(slice, converted)
		}
		return slice, nil

	default:
		converted, err := convertToType(val, elemType, tag)
		if err != nil {
			return reflect.Value{}, err
		}
		slice := reflect.MakeSlice(t, 0, 1)
		return reflect.Append(slice, converted), nil
	}
}

func convertToMap(val any, t reflect.Type) (reflect.Value, error) {
	m, ok := val.(map[string]any)
	if !ok {
		return reflect.Value{}, fmt.Errorf("cannot convert %T to %s", val, t)
	}

	keyType := t.Key()
	valType := t.Elem()

	if keyType.Kind() != reflect.String {
		return reflect.Value{}, fmt.Errorf("map key type %s is not supported, only string keys", keyType)
	}

	result := reflect.MakeMapWithSize(t, len(m))
	for k, v := range m {
		converted, err := convertToType(v, valType, "")
		if err != nil {
			return reflect.Value{}, fmt.Errorf("map key %q: %w", k, err)
		}
		result.SetMapIndex(reflect.ValueOf(k), converted)
	}

	return result, nil
}

func parseStringToType(s string, t reflect.Type, tag reflect.StructTag) (reflect.Value, error) {
	if t == durationType {
		d, err := time.ParseDuration(s)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(d), nil
	}
	if t == timeType {
		layout := tag.Get("layout")
		if layout == "" {
			layout = time.RFC3339
		}
		parsed, err := time.Parse(layout, s)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(parsed), nil
	}

	return convertToType(s, t, tag)
}

func toBool(v any) (bool, bool) {
	switch val := v.(type) {
	case bool:
		return val, true
	case string:
		switch strings.ToLower(val) {
		case "true", "1", "on", "yes", "y":
			return true, true
		case "false", "0", "off", "no", "n":
			return false, true
		}
	case int:
		return val != 0, true
	case int64:
		return val != 0, true
	case float64:
		return val != 0, true
	}
	return false, false
}
