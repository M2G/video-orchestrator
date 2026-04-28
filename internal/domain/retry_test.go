package domain

import "testing"

func TestNextDelay(t *testing.T) {

	d1 := NextDelay(1)
	d2 := NextDelay(2)

	if d2 <= d1 {
		t.Fatal("expected increasing delay")
	}
}
