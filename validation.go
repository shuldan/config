package config

import (
	"fmt"
	"regexp"
)

type Rule func(c *Config) error

func Required(key string) Rule {
	return func(c *Config) error {
		if !c.Has(key) {
			return fmt.Errorf("%q: required key is missing", key)
		}
		return nil
	}
}

func InRange(key string, min, max float64) Rule {
	return func(c *Config) error {
		if !c.Has(key) {
			return nil
		}

		v, ok := toFloat64(c.Get(key))
		if !ok {
			return fmt.Errorf("%q: value is not a number", key)
		}

		if v < min || v > max {
			return fmt.Errorf("%q: value %v is out of range [%v, %v]", key, v, min, max)
		}

		return nil
	}
}

func OneOf(key string, allowed ...string) Rule {
	return func(c *Config) error {
		if !c.Has(key) {
			return nil
		}

		v := c.GetString(key)
		for _, a := range allowed {
			if v == a {
				return nil
			}
		}

		return fmt.Errorf("%q: value %q is not one of %v", key, v, allowed)
	}
}

func MatchRegex(key string, pattern string) Rule {
	return func(c *Config) error {
		if !c.Has(key) {
			return nil
		}

		v := c.GetString(key)
		matched, err := regexp.MatchString(pattern, v)
		if err != nil {
			return fmt.Errorf("%q: invalid regex %q: %w", key, pattern, err)
		}
		if !matched {
			return fmt.Errorf("%q: value %q does not match pattern %q", key, v, pattern)
		}

		return nil
	}
}

func Custom(key string, fn func(v any) error) Rule {
	return func(c *Config) error {
		v := c.Get(key)
		if err := fn(v); err != nil {
			return fmt.Errorf("%q: %w", key, err)
		}
		return nil
	}
}

func (c *Config) Validate(rules ...Rule) error {
	var violations []string

	for _, rule := range rules {
		if err := rule(c); err != nil {
			violations = append(violations, err.Error())
		}
	}

	if len(violations) > 0 {
		return &ValidationError{Violations: violations}
	}

	return nil
}
