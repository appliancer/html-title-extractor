package extractor

import "testing"

func TestClampInt(t *testing.T) {
	if clampInt(6, 8, 14) != 8 {
		t.Error("should not be clamped")
	}
	if clampInt(10, 8, 14) != 10 {
		t.Error("should be clamped to min")
	}
	if clampInt(16, 8, 14) != 14 {
		t.Error("should be clamped to max")
	}

	func() {
		defer func() {
			if recover() == nil {
				t.Error("should panic")
			}
		}()
		clampInt(10, 14, 8)
	}()
}
