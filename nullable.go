package civiltime

import (
	"database/sql/driver"
	"encoding/json"
)

// NullDate represents a Date that may be SQL NULL or JSON null.
type NullDate struct {
	Date  Date
	Valid bool
}

// Scan implements database/sql.Scanner.
func (n *NullDate) Scan(src any) error {
	if src == nil {
		*n = NullDate{}
		return nil
	}
	var d Date
	if err := d.Scan(src); err != nil {
		return err
	}
	*n = NullDate{Date: d, Valid: true}
	return nil
}

// Value implements database/sql/driver.Valuer.
func (n NullDate) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Date.Value()
}

// MarshalJSON encodes an invalid NullDate as JSON null.
func (n NullDate) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return n.Date.MarshalJSON()
}

// UnmarshalJSON decodes a date string or JSON null.
func (n *NullDate) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*n = NullDate{}
		return nil
	}
	var d Date
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}
	*n = NullDate{Date: d, Valid: true}
	return nil
}

// String returns the date or an empty string when invalid.
func (n NullDate) String() string {
	if !n.Valid {
		return ""
	}
	return n.Date.String()
}

// NullTime represents a Time that may be SQL NULL or JSON null.
type NullTime struct {
	Time  Time
	Valid bool
}

// Scan implements database/sql.Scanner.
func (n *NullTime) Scan(src any) error {
	if src == nil {
		*n = NullTime{}
		return nil
	}
	var tm Time
	if err := tm.Scan(src); err != nil {
		return err
	}
	*n = NullTime{Time: tm, Valid: true}
	return nil
}

// Value implements database/sql/driver.Valuer.
func (n NullTime) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Time.Value()
}

// MarshalJSON encodes an invalid NullTime as JSON null.
func (n NullTime) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return n.Time.MarshalJSON()
}

// UnmarshalJSON decodes a time string or JSON null.
func (n *NullTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*n = NullTime{}
		return nil
	}
	var tm Time
	if err := json.Unmarshal(data, &tm); err != nil {
		return err
	}
	*n = NullTime{Time: tm, Valid: true}
	return nil
}

// String returns the time or an empty string when invalid.
func (n NullTime) String() string {
	if !n.Valid {
		return ""
	}
	return n.Time.String()
}

// NullDateTime represents a DateTime that may be SQL NULL or JSON null.
type NullDateTime struct {
	DateTime DateTime
	Valid    bool
}

// Scan implements database/sql.Scanner.
func (n *NullDateTime) Scan(src any) error {
	if src == nil {
		*n = NullDateTime{}
		return nil
	}
	var dt DateTime
	if err := dt.Scan(src); err != nil {
		return err
	}
	*n = NullDateTime{DateTime: dt, Valid: true}
	return nil
}

// Value implements database/sql/driver.Valuer.
func (n NullDateTime) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.DateTime.Value()
}

// MarshalJSON encodes an invalid NullDateTime as JSON null.
func (n NullDateTime) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return n.DateTime.MarshalJSON()
}

// UnmarshalJSON decodes a datetime string or JSON null.
func (n *NullDateTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*n = NullDateTime{}
		return nil
	}
	var dt DateTime
	if err := json.Unmarshal(data, &dt); err != nil {
		return err
	}
	*n = NullDateTime{DateTime: dt, Valid: true}
	return nil
}

// String returns the datetime or an empty string when invalid.
func (n NullDateTime) String() string {
	if !n.Valid {
		return ""
	}
	return n.DateTime.String()
}
