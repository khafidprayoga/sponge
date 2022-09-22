// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: api/types/types.proto

package types

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/anypb"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = anypb.Any{}
	_ = sort.Sort
)

// Validate checks the field values on Column with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *Column) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on Column with the rules defined in the
// proto definition for this message. If any rules are violated, the result is
// a list of violation errors wrapped in ColumnMultiError, or nil if none found.
func (m *Column) ValidateAll() error {
	return m.validate(true)
}

func (m *Column) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Name

	// no validation rules for Exp

	// no validation rules for Value

	// no validation rules for Logic

	if len(errors) > 0 {
		return ColumnMultiError(errors)
	}

	return nil
}

// ColumnMultiError is an error wrapping multiple validation errors returned by
// Column.ValidateAll() if the designated constraints aren't met.
type ColumnMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ColumnMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ColumnMultiError) AllErrors() []error { return m }

// ColumnValidationError is the validation error returned by Column.Validate if
// the designated constraints aren't met.
type ColumnValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ColumnValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ColumnValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ColumnValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ColumnValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ColumnValidationError) ErrorName() string { return "ColumnValidationError" }

// Error satisfies the builtin error interface
func (e ColumnValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sColumn.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ColumnValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ColumnValidationError{}

// Validate checks the field values on Params with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *Params) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on Params with the rules defined in the
// proto definition for this message. If any rules are violated, the result is
// a list of violation errors wrapped in ParamsMultiError, or nil if none found.
func (m *Params) ValidateAll() error {
	return m.validate(true)
}

func (m *Params) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Page

	// no validation rules for Limit

	// no validation rules for Sort

	for idx, item := range m.GetColumns() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, ParamsValidationError{
						field:  fmt.Sprintf("Columns[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, ParamsValidationError{
						field:  fmt.Sprintf("Columns[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return ParamsValidationError{
					field:  fmt.Sprintf("Columns[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if len(errors) > 0 {
		return ParamsMultiError(errors)
	}

	return nil
}

// ParamsMultiError is an error wrapping multiple validation errors returned by
// Params.ValidateAll() if the designated constraints aren't met.
type ParamsMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ParamsMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ParamsMultiError) AllErrors() []error { return m }

// ParamsValidationError is the validation error returned by Params.Validate if
// the designated constraints aren't met.
type ParamsValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ParamsValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ParamsValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ParamsValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ParamsValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ParamsValidationError) ErrorName() string { return "ParamsValidationError" }

// Error satisfies the builtin error interface
func (e ParamsValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sParams.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ParamsValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ParamsValidationError{}