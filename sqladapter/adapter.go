// Package sqladapter adapts domain values to database/sql.
//
// It has no dependency on a concrete database driver. A driver-specific
// package can provide a Codec when its wire representation needs special
// handling.
package sqladapter

import (
	"database/sql"
	"database/sql/driver"
	"errors"

	"github.com/MouXiaoJun/civiltime"
)

var (
	// ErrNilTarget reports an adapter without a destination value.
	ErrNilTarget = errors.New("sqladapter: nil target")
	// ErrMissingDecoder reports a codec without a Decode function.
	ErrMissingDecoder = errors.New("sqladapter: missing decoder")
	// ErrMissingEncoder reports a codec without an Encode function.
	ErrMissingEncoder = errors.New("sqladapter: missing encoder")
)

// Codec converts a database value to and from a domain value.
type Codec[T any] struct {
	Decode func(any) (T, error)
	Encode func(T) (driver.Value, error)
}

// Adapter implements database/sql.Scanner and driver.Valuer for a Codec.
type Adapter[T any] struct {
	target *T
	codec  Codec[T]
}

// New binds codec to target. The returned adapter can be passed to Scan and
// used as a query argument through database/sql.
func New[T any](target *T, codec Codec[T]) *Adapter[T] {
	return &Adapter[T]{target: target, codec: codec}
}

// Scan decodes src into the target value.
func (a *Adapter[T]) Scan(src any) error {
	if a == nil || a.target == nil {
		return ErrNilTarget
	}
	if a.codec.Decode == nil {
		return ErrMissingDecoder
	}
	value, err := a.codec.Decode(src)
	if err != nil {
		return err
	}
	*a.target = value
	return nil
}

// Value encodes the current target value for database/sql.
func (a *Adapter[T]) Value() (driver.Value, error) {
	if a == nil || a.target == nil {
		return nil, ErrNilTarget
	}
	if a.codec.Encode == nil {
		return nil, ErrMissingEncoder
	}
	return a.codec.Encode(*a.target)
}

// Date adapts a civiltime.Date using its standard SQL representation.
func Date(target *civiltime.Date) *Adapter[civiltime.Date] {
	return New(target, Codec[civiltime.Date]{
		Decode: func(src any) (civiltime.Date, error) {
			var value civiltime.Date
			if err := value.Scan(src); err != nil {
				return civiltime.Date{}, err
			}
			return value, nil
		},
		Encode: func(value civiltime.Date) (driver.Value, error) {
			return value.Value()
		},
	})
}

// Time adapts a civiltime.Time using its standard SQL representation.
func Time(target *civiltime.Time) *Adapter[civiltime.Time] {
	return New(target, Codec[civiltime.Time]{
		Decode: func(src any) (civiltime.Time, error) {
			var value civiltime.Time
			if err := value.Scan(src); err != nil {
				return civiltime.Time{}, err
			}
			return value, nil
		},
		Encode: func(value civiltime.Time) (driver.Value, error) {
			return value.Value()
		},
	})
}

// DateTime adapts a civiltime.DateTime using its standard SQL representation.
func DateTime(target *civiltime.DateTime) *Adapter[civiltime.DateTime] {
	return New(target, Codec[civiltime.DateTime]{
		Decode: func(src any) (civiltime.DateTime, error) {
			var value civiltime.DateTime
			if err := value.Scan(src); err != nil {
				return civiltime.DateTime{}, err
			}
			return value, nil
		},
		Encode: func(value civiltime.DateTime) (driver.Value, error) {
			return value.Value()
		},
	})
}

var _ sql.Scanner = (*Adapter[int])(nil)
var _ driver.Valuer = (*Adapter[int])(nil)
