package civiltime

import "testing"

func BenchmarkParseDate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := ParseDate("2026-08-29"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseDateTime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := ParseDateTime("2026-08-29T12:30:45.123456789"); err != nil {
			b.Fatal(err)
		}
	}
}
