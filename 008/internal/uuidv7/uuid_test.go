package uuidv7

import (
	"encoding/hex"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		before := time.Now().UnixMilli()
		id, err := New()
		after := time.Now().UnixMilli()
		if err != nil {
			t.Fatal(err)
		}
		if len(id) != 36 || id[8] != '-' || id[13] != '-' || id[18] != '-' || id[23] != '-' {
			t.Fatalf("invalid UUID format: %q", id)
		}
		data, err := hex.DecodeString(strings.ReplaceAll(id, "-", ""))
		if err != nil {
			t.Fatal(err)
		}
		if data[6]>>4 != 7 || data[8]>>6 != 2 {
			t.Fatalf("invalid UUID version or variant: %q", id)
		}
		var millis int64
		for _, b := range data[:6] {
			millis = millis<<8 | int64(b)
		}
		if millis < before || millis > after {
			t.Fatalf("timestamp %d outside [%d, %d]", millis, before, after)
		}
		if seen[id] {
			t.Fatalf("duplicate UUID: %q", id)
		}
		seen[id] = true
	}
}
