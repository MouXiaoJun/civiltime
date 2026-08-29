package sqladapter

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/MouXiaoJun/civiltime"
)

func TestDateTimeAdapter(t *testing.T) {
	var got civiltime.DateTime
	adapter := DateTime(&got)

	var scanner sql.Scanner = adapter
	if err := scanner.Scan([]byte("2026-01-02 03:04:05")); err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if got.String() != "2026-01-02T03:04:05" {
		t.Fatalf("DateTime = %s", got.String())
	}

	var valuer driver.Valuer = adapter
	value, err := valuer.Value()
	if err != nil || value != driver.Value("2026-01-02T03:04:05") {
		t.Fatalf("Value() = %#v, %v", value, err)
	}
}

func TestCustomCodec(t *testing.T) {
	codec := Codec[int]{
		Decode: func(src any) (int, error) {
			value, ok := src.(int64)
			if !ok {
				return 0, errors.New("want int64")
			}
			return int(value), nil
		},
		Encode: func(value int) (driver.Value, error) {
			return int64(value), nil
		},
	}
	var got int
	adapter := New(&got, codec)
	if err := adapter.Scan(int64(42)); err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if got != 42 {
		t.Fatalf("target = %d", got)
	}
	value, err := adapter.Value()
	if err != nil || value != int64(42) {
		t.Fatalf("Value() = %#v, %v", value, err)
	}
}

func TestAdapterErrors(t *testing.T) {
	if err := New[int](nil, Codec[int]{}).Scan(nil); !errors.Is(err, ErrNilTarget) {
		t.Fatalf("nil target Scan() error = %v", err)
	}
	var value int
	if err := New(&value, Codec[int]{}).Scan(nil); !errors.Is(err, ErrMissingDecoder) {
		t.Fatalf("missing decoder Scan() error = %v", err)
	}
	if _, err := New(&value, Codec[int]{}).Value(); !errors.Is(err, ErrMissingEncoder) {
		t.Fatalf("missing encoder Value() error = %v", err)
	}
}
