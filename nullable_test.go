package civiltime

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"testing"
	stdtime "time"
)

func TestNullDate(t *testing.T) {
	want := NullDate{Date: Date{Year: 2024, Month: stdtime.January, Day: 2}, Valid: true}
	value, err := want.Value()
	if err != nil || value != driver.Value("2024-01-02") {
		t.Fatalf("Value() = %#v, %v", value, err)
	}
	var got NullDate
	if err := got.Scan(nil); err != nil || got.Valid {
		t.Fatalf("Scan(nil) = %#v, %v", got, err)
	}
	if err := json.Unmarshal([]byte(`"2024-01-02"`), &got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("UnmarshalJSON() = %#v, want %#v", got, want)
	}
	if err := json.Unmarshal([]byte("null"), &got); err != nil || got.Valid {
		t.Fatalf("null = %#v, %v", got, err)
	}
}

func TestNullTimeAndDateTime(t *testing.T) {
	var nt NullTime
	if err := nt.Scan("03:04:05"); err != nil {
		t.Fatal(err)
	}
	if !nt.Valid || nt.Time.String() != "03:04:05" {
		t.Fatalf("NullTime = %#v", nt)
	}

	var ndt NullDateTime
	if err := ndt.Scan("2024-01-02T03:04:05"); err != nil {
		t.Fatal(err)
	}
	if !ndt.Valid || ndt.DateTime.String() != "2024-01-02T03:04:05" {
		t.Fatalf("NullDateTime = %#v", ndt)
	}
	data, err := json.Marshal(NullDateTime{})
	if err != nil || string(data) != "null" {
		t.Fatalf("MarshalJSON() = %s, %v", data, err)
	}
}

func TestNullValuesRejectInvalidNonNullValues(t *testing.T) {
	var nd NullDate
	if err := nd.Scan("2024-02-30"); !errors.Is(err, ErrInvalidDate) {
		t.Fatalf("invalid date error = %v", err)
	}
	if nd.Valid {
		t.Fatal("invalid scan must not mark value valid")
	}
}
