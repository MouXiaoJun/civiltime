package civiltime

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	stdtime "time"
)

func TestDate(t *testing.T) {
	d, err := ParseDate("2024-02-29")
	if err != nil {
		t.Fatal(err)
	}
	if got := d.AddDays(1).String(); got != "2024-03-01" {
		t.Fatalf("AddDays() = %q", got)
	}
	if got := d.AddMonths(12).String(); got != "2025-02-28" {
		t.Fatalf("AddMonths() = %q", got)
	}
	if got := d.Weekday(); got != stdtime.Thursday {
		t.Fatalf("Weekday() = %v", got)
	}
	if _, err := ParseDate("2023-02-29"); !errors.Is(err, ErrInvalidDate) {
		t.Fatalf("ParseDate() error = %v", err)
	}
	for _, test := range []struct {
		date   Date
		months int
		want   string
	}{
		{Date{Year: 2024, Month: stdtime.January, Day: 31}, 1, "2024-02-29"},
		{Date{Year: 2024, Month: stdtime.March, Day: 31}, -1, "2024-02-29"},
	} {
		if got := test.date.AddMonths(test.months).String(); got != test.want {
			t.Errorf("AddMonths(%d) = %q, want %q", test.months, got, test.want)
		}
	}
}

func TestTime(t *testing.T) {
	tm, err := ParseTime("23:59:59.1200")
	if err != nil {
		t.Fatal(err)
	}
	if got := tm.String(); got != "23:59:59.12" {
		t.Fatalf("String() = %q", got)
	}
	if !tm.After(Time{Hour: 23, Minute: 59, Second: 59}) {
		t.Fatal("fractional time should compare after whole seconds")
	}
	for _, input := range []string{"24:00:00", "12:60:00", "12:00:60", "12:00:00.1234567890"} {
		if _, err := ParseTime(input); !errors.Is(err, ErrInvalidTime) {
			t.Errorf("ParseTime(%q) error = %v", input, err)
		}
	}
}

func TestDateTimeAndExplicitLocation(t *testing.T) {
	dt, err := ParseDateTime("2024-06-01 12:30:00")
	if err != nil {
		t.Fatal(err)
	}
	loc := stdtime.FixedZone("UTC+8", 8*60*60)
	got := dt.In(loc)
	if got.Format(stdtime.RFC3339) != "2024-06-01T12:30:00+08:00" {
		t.Fatalf("In() = %s", got.Format(stdtime.RFC3339))
	}
	if got := dt.AddDays(1).String(); got != "2024-06-02T12:30:00" {
		t.Fatalf("AddDays() = %q", got)
	}
	if got := dt.AddMonths(1).String(); got != "2024-07-01T12:30:00" {
		t.Fatalf("AddMonths() = %q", got)
	}
	late := DateTime{Date: Date{Year: 2024, Month: stdtime.January, Day: 1}, Time: Time{Hour: 23, Minute: 30}}
	if got := late.Add(90 * stdtime.Minute).String(); got != "2024-01-02T01:00:00" {
		t.Fatalf("Add() = %q", got)
	}
	if got := late.Add(-90 * stdtime.Minute).String(); got != "2024-01-01T22:00:00" {
		t.Fatalf("Add(negative) = %q", got)
	}
	if got := late.Add(90 * stdtime.Minute).Sub(late); got != 90*stdtime.Minute {
		t.Fatalf("Sub() = %v", got)
	}
}

func TestJSONAndText(t *testing.T) {
	want := DateTime{Date: Date{Year: 2025, Month: stdtime.January, Day: 2}, Time: Time{Hour: 3, Minute: 4, Second: 5, Nanosecond: 6000000}}
	b, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(b); got != `"2025-01-02T03:04:05.006"` {
		t.Fatalf("MarshalJSON() = %s", got)
	}
	var got DateTime
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("UnmarshalJSON() = %#v, want %#v", got, want)
	}
	if err := json.Unmarshal([]byte("null"), &got); !errors.Is(err, ErrNull) {
		t.Fatalf("null error = %v", err)
	}
}

func TestSQLValueAndScan(t *testing.T) {
	want := DateTime{Date: Date{Year: 2025, Month: stdtime.January, Day: 2}, Time: Time{Hour: 3, Minute: 4, Second: 5}}
	value, err := want.Value()
	if err != nil {
		t.Fatal(err)
	}
	if value != driver.Value("2025-01-02T03:04:05") {
		t.Fatalf("Value() = %#v", value)
	}
	var got DateTime
	if err := got.Scan([]byte("2025-01-02 03:04:05")); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("Scan() = %#v, want %#v", got, want)
	}
	if err := got.Scan(nil); !errors.Is(err, ErrNull) {
		t.Fatalf("Scan(nil) error = %v", err)
	}
	if _, err := (Date{}).Value(); !errors.Is(err, ErrInvalidDate) {
		t.Fatalf("zero Value() error = %v", err)
	}
}

func TestScanTimeIgnoresLocation(t *testing.T) {
	loc := stdtime.FixedZone("UTC+8", 8*60*60)
	var got Time
	if err := got.Scan(stdtime.Date(2024, 1, 2, 3, 4, 5, 0, loc)); err != nil {
		t.Fatal(err)
	}
	if got != (Time{Hour: 3, Minute: 4, Second: 5}) {
		t.Fatalf("Scan() = %#v", got)
	}
}

func FuzzParsers(f *testing.F) {
	for _, seed := range []string{
		"2024-02-29",
		"23:59:59.999999999",
		"2024-02-29T23:59:59",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		if d, err := ParseDate(input); err == nil && !d.IsValid() {
			t.Fatalf("ParseDate returned invalid value: %#v", d)
		}
		if tm, err := ParseTime(input); err == nil && !tm.IsValid() {
			t.Fatalf("ParseTime returned invalid value: %#v", tm)
		}
		if dt, err := ParseDateTime(input); err == nil && !dt.IsValid() {
			t.Fatalf("ParseDateTime returned invalid value: %#v", dt)
		}
	})
}
