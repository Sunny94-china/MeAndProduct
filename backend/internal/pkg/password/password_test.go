package password

import "testing"

func TestHashAndVerify(t *testing.T) {
	pwd := "Hello@123"
	h, err := Hash(pwd)
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}
	if len(h) < 40 {
		t.Errorf("hash looks too short: %q", h)
	}
	if err := Verify(pwd, h); err != nil {
		t.Errorf("Verify correct pwd failed: %v", err)
	}
	if err := Verify("wrong", h); err == nil {
		t.Error("expected error for wrong password")
	}
	if _, err := Hash(""); err != ErrEmptyPassword {
		t.Errorf("Hash('') should return ErrEmptyPassword, got %v", err)
	}
}

func BenchmarkHash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Hash("benchmark-pwd-xyz")
	}
}
