// Package civiltime represents calendar values that do not identify a unique
// instant. Use DateTime.In only at a boundary where a location is known.
package civiltime

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	stdtime "time"
)

const (
	dateLayout = "2006-01-02"
	timeLayout = "15:04:05"
)

var (
	// ErrInvalidDate reports a date with an impossible calendar value.
	ErrInvalidDate = errors.New("civiltime: invalid date")
	// ErrInvalidTime reports a time with an impossible clock value.
	ErrInvalidTime = errors.New("civiltime: invalid time")
	// ErrInvalidDateTime reports a DateTime containing an invalid Date or Time.
	ErrInvalidDateTime = errors.New("civiltime: invalid datetime")
	// ErrNull reports an attempt to scan or unmarshal null into a non-nullable value.
	ErrNull = errors.New("civiltime: null is not allowed")
)

// Date is a calendar date without location or clock-time information.
type Date struct {
	Year  int
	Month stdtime.Month
	Day   int
}

// DateOf extracts the calendar date in t's location.
func DateOf(t stdtime.Time) Date {
	return Date{Year: t.Year(), Month: t.Month(), Day: t.Day()}
}

// ParseDate parses YYYY-MM-DD.
func ParseDate(s string) (Date, error) {
	t, err := stdtime.Parse(dateLayout, s)
	if err != nil {
		return Date{}, fmt.Errorf("%w: %v", ErrInvalidDate, err)
	}
	return DateOf(t), nil
}

// IsValid reports whether d is a real date representable by String.
func (d Date) IsValid() bool {
	if d.Year < 0 || d.Year > 9999 || d.Month < stdtime.January || d.Month > stdtime.December || d.Day < 1 {
		return false
	}
	next := stdtime.Date(d.Year, d.Month+1, 1, 0, 0, 0, 0, stdtime.UTC)
	last := next.AddDate(0, 0, -1).Day()
	return d.Day <= last
}

// IsZero reports whether d is its zero value. The zero value is not a valid date.
func (d Date) IsZero() bool { return d == Date{} }

// Compare returns -1, 0, or 1 according to d's order relative to other.
func (d Date) Compare(other Date) int {
	if d.Year != other.Year {
		if d.Year < other.Year {
			return -1
		}
		return 1
	}
	if d.Month != other.Month {
		if d.Month < other.Month {
			return -1
		}
		return 1
	}
	if d.Day < other.Day {
		return -1
	}
	if d.Day > other.Day {
		return 1
	}
	return 0
}

// Before reports whether d comes before other.
func (d Date) Before(other Date) bool { return d.Compare(other) < 0 }

// After reports whether d comes after other.
func (d Date) After(other Date) bool { return d.Compare(other) > 0 }

// AddDays returns d shifted by n calendar days. Invalid dates are returned unchanged.
func (d Date) AddDays(n int) Date {
	if !d.IsValid() {
		return d
	}
	return DateOf(stdtime.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, stdtime.UTC).AddDate(0, 0, n))
}

// AddMonths returns d shifted by n calendar months. If the target month has
// fewer days, the result is clamped to that month's last day.
func (d Date) AddMonths(n int) Date {
	if !d.IsValid() {
		return d
	}
	first := stdtime.Date(d.Year, d.Month, 1, 0, 0, 0, 0, stdtime.UTC).AddDate(0, n, 0)
	result := Date{Year: first.Year(), Month: first.Month(), Day: d.Day}
	if !result.IsValid() {
		result.Day = stdtime.Date(result.Year, result.Month+1, 0, 0, 0, 0, 0, stdtime.UTC).Day()
	}
	return result
}

// Weekday returns the day of the week for d. It returns the zero value for an invalid date.
func (d Date) Weekday() stdtime.Weekday {
	if !d.IsValid() {
		return stdtime.Sunday
	}
	return stdtime.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, stdtime.UTC).Weekday()
}

// In returns midnight on d in loc. It panics when d or loc is invalid.
func (d Date) In(loc *stdtime.Location) stdtime.Time {
	if !d.IsValid() {
		panic(ErrInvalidDate)
	}
	if loc == nil {
		panic("civiltime: nil location")
	}
	return stdtime.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, loc)
}

// String returns d in canonical YYYY-MM-DD form, or <invalid-date>.
func (d Date) String() string {
	if !d.IsValid() {
		return "<invalid-date>"
	}
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

// MarshalText implements encoding.TextMarshaler.
func (d Date) MarshalText() ([]byte, error) {
	if err := d.validate(); err != nil {
		return nil, err
	}
	return []byte(d.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (d *Date) UnmarshalText(data []byte) error {
	if d == nil {
		return errors.New("civiltime: nil Date receiver")
	}
	parsed, err := ParseDate(string(data))
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

// MarshalJSON encodes d as a JSON string.
func (d Date) MarshalJSON() ([]byte, error) {
	text, err := d.MarshalText()
	if err != nil {
		return nil, err
	}
	return json.Marshal(string(text))
}

// UnmarshalJSON decodes a JSON date string. null is rejected.
func (d *Date) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return ErrNull
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	return d.UnmarshalText([]byte(s))
}

// Scan implements database/sql.Scanner. SQL NULL is rejected; use a nullable
// database field type when NULL is part of the schema.
func (d *Date) Scan(src any) error {
	if d == nil {
		return errors.New("civiltime: nil Date receiver")
	}
	parsed, err := scanDate(src)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

// Value implements database/sql/driver.Valuer.
func (d Date) Value() (driver.Value, error) {
	if err := d.validate(); err != nil {
		return nil, err
	}
	return d.String(), nil
}

func (d Date) validate() error {
	if !d.IsValid() {
		return ErrInvalidDate
	}
	return nil
}

// Time is a wall-clock time without location information.
type Time struct {
	Hour       int
	Minute     int
	Second     int
	Nanosecond int
}

// TimeOf extracts the wall-clock time in t's location.
func TimeOf(t stdtime.Time) Time {
	return Time{Hour: t.Hour(), Minute: t.Minute(), Second: t.Second(), Nanosecond: t.Nanosecond()}
}

// ParseTime parses HH:MM:SS with an optional fractional part of one to nine digits.
func ParseTime(s string) (Time, error) {
	if len(s) < 8 || s[2] != ':' || s[5] != ':' || (len(s) > 8 && s[8] != '.') {
		return Time{}, fmt.Errorf("%w: invalid format", ErrInvalidTime)
	}
	if len(s) > 8 {
		fraction := s[9:]
		if len(fraction) == 0 || len(fraction) > 9 {
			return Time{}, fmt.Errorf("%w: invalid fraction", ErrInvalidTime)
		}
		for i := range fraction {
			if fraction[i] < '0' || fraction[i] > '9' {
				return Time{}, fmt.Errorf("%w: invalid fraction", ErrInvalidTime)
			}
		}
	}
	t, err := stdtime.Parse(timeLayout, s)
	if err != nil {
		return Time{}, fmt.Errorf("%w: %v", ErrInvalidTime, err)
	}
	return TimeOf(t), nil
}

// IsValid reports whether t is a valid 24-hour clock time.
func (t Time) IsValid() bool {
	return t.Hour >= 0 && t.Hour <= 23 &&
		t.Minute >= 0 && t.Minute <= 59 &&
		t.Second >= 0 && t.Second <= 59 &&
		t.Nanosecond >= 0 && t.Nanosecond <= 999999999
}

// IsZero reports whether t is its zero value. Midnight is also the zero value.
func (t Time) IsZero() bool { return t == Time{} }

// Compare returns -1, 0, or 1 according to t's order relative to other.
func (t Time) Compare(other Time) int {
	left := t.nanoseconds()
	right := other.nanoseconds()
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}

// Before reports whether t comes before other.
func (t Time) Before(other Time) bool { return t.Compare(other) < 0 }

// After reports whether t comes after other.
func (t Time) After(other Time) bool { return t.Compare(other) > 0 }

// String returns t in canonical HH:MM:SS[.fraction] form, or <invalid-time>.
func (t Time) String() string {
	if !t.IsValid() {
		return "<invalid-time>"
	}
	return stdtime.Date(0, stdtime.January, 1, t.Hour, t.Minute, t.Second, t.Nanosecond, stdtime.UTC).Format(timeLayout + ".999999999")
}

// MarshalText implements encoding.TextMarshaler.
func (t Time) MarshalText() ([]byte, error) {
	if err := t.validate(); err != nil {
		return nil, err
	}
	return []byte(t.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (t *Time) UnmarshalText(data []byte) error {
	if t == nil {
		return errors.New("civiltime: nil Time receiver")
	}
	parsed, err := ParseTime(string(data))
	if err != nil {
		return err
	}
	*t = parsed
	return nil
}

// MarshalJSON encodes t as a JSON string.
func (t Time) MarshalJSON() ([]byte, error) {
	text, err := t.MarshalText()
	if err != nil {
		return nil, err
	}
	return json.Marshal(string(text))
}

// UnmarshalJSON decodes a JSON time string. null is rejected.
func (t *Time) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return ErrNull
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	return t.UnmarshalText([]byte(s))
}

// Scan implements database/sql.Scanner. SQL NULL is rejected.
func (t *Time) Scan(src any) error {
	if t == nil {
		return errors.New("civiltime: nil Time receiver")
	}
	parsed, err := scanTime(src)
	if err != nil {
		return err
	}
	*t = parsed
	return nil
}

// Value implements database/sql/driver.Valuer.
func (t Time) Value() (driver.Value, error) {
	if err := t.validate(); err != nil {
		return nil, err
	}
	return t.String(), nil
}

func (t Time) nanoseconds() int64 {
	return (((int64(t.Hour)*60+int64(t.Minute))*60+int64(t.Second))*1e9 + int64(t.Nanosecond))
}

func (t Time) validate() error {
	if !t.IsValid() {
		return ErrInvalidTime
	}
	return nil
}

// DateTime combines a Date and a Time without location information. It does
// not identify a unique instant until converted with In.
type DateTime struct {
	Date Date
	Time Time
}

// DateTimeOf extracts the civil date and time in t's location.
func DateTimeOf(t stdtime.Time) DateTime {
	return DateTime{Date: DateOf(t), Time: TimeOf(t)}
}

// ParseDateTime parses YYYY-MM-DDTHH:MM:SS[.fraction]. A lower-case t or a
// space is also accepted for database text formats.
func ParseDateTime(s string) (DateTime, error) {
	if len(s) < len("2006-01-02T15:04:05") {
		return DateTime{}, fmt.Errorf("%w: %q", ErrInvalidDateTime, s)
	}
	sep := s[10]
	if sep != 'T' && sep != 't' && sep != ' ' {
		return DateTime{}, fmt.Errorf("%w: invalid separator", ErrInvalidDateTime)
	}
	d, err := ParseDate(s[:10])
	if err != nil {
		return DateTime{}, fmt.Errorf("%w: %v", ErrInvalidDateTime, err)
	}
	t, err := ParseTime(s[11:])
	if err != nil {
		return DateTime{}, fmt.Errorf("%w: %v", ErrInvalidDateTime, err)
	}
	return DateTime{Date: d, Time: t}, nil
}

// IsValid reports whether both components are valid.
func (dt DateTime) IsValid() bool { return dt.Date.IsValid() && dt.Time.IsValid() }

// IsZero reports whether both components are zero values.
func (dt DateTime) IsZero() bool { return dt.Date.IsZero() && dt.Time.IsZero() }

// Compare returns -1, 0, or 1 according to dt's order relative to other.
func (dt DateTime) Compare(other DateTime) int {
	if c := dt.Date.Compare(other.Date); c != 0 {
		return c
	}
	return dt.Time.Compare(other.Time)
}

// Before reports whether dt comes before other.
func (dt DateTime) Before(other DateTime) bool { return dt.Compare(other) < 0 }

// After reports whether dt comes after other.
func (dt DateTime) After(other DateTime) bool { return dt.Compare(other) > 0 }

// AddDays returns dt shifted by n calendar days. Invalid values are returned unchanged.
func (dt DateTime) AddDays(n int) DateTime {
	if !dt.IsValid() {
		return dt
	}
	return DateTime{Date: dt.Date.AddDays(n), Time: dt.Time}
}

// AddMonths returns dt shifted by n calendar months while preserving its time.
func (dt DateTime) AddMonths(n int) DateTime {
	if !dt.IsValid() {
		return dt
	}
	return DateTime{Date: dt.Date.AddMonths(n), Time: dt.Time}
}

// Add returns dt shifted by d. The arithmetic uses fixed 24-hour civil days;
// no timezone or daylight-saving rule is involved.
func (dt DateTime) Add(d stdtime.Duration) DateTime {
	if !dt.IsValid() {
		return dt
	}
	t := stdtime.Date(dt.Date.Year, dt.Date.Month, dt.Date.Day, dt.Time.Hour, dt.Time.Minute, dt.Time.Second, dt.Time.Nanosecond, stdtime.UTC)
	return DateTimeOf(t.Add(d))
}

// Sub returns the civil-time difference dt-other. Invalid values return zero.
// Like time.Time.Sub, an unrepresentable result is saturated to the limits of
// time.Duration.
func (dt DateTime) Sub(other DateTime) stdtime.Duration {
	if !dt.IsValid() || !other.IsValid() {
		return 0
	}
	left := stdtime.Date(dt.Date.Year, dt.Date.Month, dt.Date.Day, dt.Time.Hour, dt.Time.Minute, dt.Time.Second, dt.Time.Nanosecond, stdtime.UTC)
	right := stdtime.Date(other.Date.Year, other.Date.Month, other.Date.Day, other.Time.Hour, other.Time.Minute, other.Time.Second, other.Time.Nanosecond, stdtime.UTC)
	return left.Sub(right)
}

// In converts dt to a time in loc. The conversion is explicit because dt has
// no timezone of its own; DST rules are therefore applied by time.Date.
func (dt DateTime) In(loc *stdtime.Location) stdtime.Time {
	if !dt.IsValid() {
		panic(ErrInvalidDateTime)
	}
	if loc == nil {
		panic("civiltime: nil location")
	}
	return stdtime.Date(dt.Date.Year, dt.Date.Month, dt.Date.Day, dt.Time.Hour, dt.Time.Minute, dt.Time.Second, dt.Time.Nanosecond, loc)
}

// String returns dt in canonical YYYY-MM-DDTHH:MM:SS[.fraction] form, or <invalid-datetime>.
func (dt DateTime) String() string {
	if !dt.IsValid() {
		return "<invalid-datetime>"
	}
	return dt.Date.String() + "T" + dt.Time.String()
}

// MarshalText implements encoding.TextMarshaler.
func (dt DateTime) MarshalText() ([]byte, error) {
	if err := dt.validate(); err != nil {
		return nil, err
	}
	return []byte(dt.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (dt *DateTime) UnmarshalText(data []byte) error {
	if dt == nil {
		return errors.New("civiltime: nil DateTime receiver")
	}
	parsed, err := ParseDateTime(string(data))
	if err != nil {
		return err
	}
	*dt = parsed
	return nil
}

// MarshalJSON encodes dt as a JSON string.
func (dt DateTime) MarshalJSON() ([]byte, error) {
	text, err := dt.MarshalText()
	if err != nil {
		return nil, err
	}
	return json.Marshal(string(text))
}

// UnmarshalJSON decodes a JSON datetime string. null is rejected.
func (dt *DateTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return ErrNull
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	return dt.UnmarshalText([]byte(s))
}

// Scan implements database/sql.Scanner. A time.Time is read by its calendar
// fields; its location and offset are deliberately ignored.
func (dt *DateTime) Scan(src any) error {
	if dt == nil {
		return errors.New("civiltime: nil DateTime receiver")
	}
	parsed, err := scanDateTime(src)
	if err != nil {
		return err
	}
	*dt = parsed
	return nil
}

// Value implements database/sql/driver.Valuer.
func (dt DateTime) Value() (driver.Value, error) {
	if err := dt.validate(); err != nil {
		return nil, err
	}
	return dt.String(), nil
}

func (dt DateTime) validate() error {
	if !dt.IsValid() {
		return ErrInvalidDateTime
	}
	return nil
}

func scanDate(src any) (Date, error) {
	switch v := src.(type) {
	case nil:
		return Date{}, ErrNull
	case stdtime.Time:
		d := DateOf(v)
		return d, d.validate()
	case string:
		return ParseDate(v)
	case []byte:
		return ParseDate(string(v))
	default:
		return Date{}, fmt.Errorf("civiltime: cannot scan %T into Date", src)
	}
}

func scanTime(src any) (Time, error) {
	switch v := src.(type) {
	case nil:
		return Time{}, ErrNull
	case stdtime.Time:
		t := TimeOf(v)
		return t, t.validate()
	case string:
		return ParseTime(v)
	case []byte:
		return ParseTime(string(v))
	default:
		return Time{}, fmt.Errorf("civiltime: cannot scan %T into Time", src)
	}
}

func scanDateTime(src any) (DateTime, error) {
	switch v := src.(type) {
	case nil:
		return DateTime{}, ErrNull
	case stdtime.Time:
		dt := DateTimeOf(v)
		return dt, dt.validate()
	case string:
		return ParseDateTime(v)
	case []byte:
		return ParseDateTime(string(v))
	default:
		return DateTime{}, fmt.Errorf("civiltime: cannot scan %T into DateTime", src)
	}
}

var (
	_ encoding.TextMarshaler   = Date{}
	_ encoding.TextUnmarshaler = (*Date)(nil)
	_ json.Marshaler           = Date{}
	_ json.Unmarshaler         = (*Date)(nil)
	_ sql.Scanner              = (*Date)(nil)
	_ driver.Valuer            = Date{}
	_ encoding.TextMarshaler   = Time{}
	_ encoding.TextUnmarshaler = (*Time)(nil)
	_ json.Marshaler           = Time{}
	_ json.Unmarshaler         = (*Time)(nil)
	_ sql.Scanner              = (*Time)(nil)
	_ driver.Valuer            = Time{}
	_ encoding.TextMarshaler   = DateTime{}
	_ encoding.TextUnmarshaler = (*DateTime)(nil)
	_ json.Marshaler           = DateTime{}
	_ json.Unmarshaler         = (*DateTime)(nil)
	_ sql.Scanner              = (*DateTime)(nil)
	_ driver.Valuer            = DateTime{}
)
