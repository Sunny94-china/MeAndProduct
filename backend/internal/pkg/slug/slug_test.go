package slug

import "testing"

func TestGenerate(t *testing.T) {
	cases := []struct {
		d, fb, want string
	}{
		{"valid-slug_01", "anything", "valid-slug_01"},
		{"IN-valid!", "fallback", "fallback"},
		{"  ", "Hello World", "hello-world"},
		{"", "张小明", ""}, // 中文全被洗掉 → sanitize 为空 → 退回 "founder"
		{"", "  User@Name.com  ", "user-name-com"},
	}
	for _, c := range cases {
		got := Generate(c.d, c.fb)
		if got == "" {
			got = "founder"
		}
		if c.want == "" {
			// fallback 为中文时的结果预期就是 founder
			if !IsValid(got) {
				t.Errorf("Generate(%q,%q) = %q invalid", c.d, c.fb, got)
			}
			continue
		}
		if got != c.want {
			t.Errorf("Generate(%q,%q) = %q want %q", c.d, c.fb, got, c.want)
		}
	}
}

func TestIsUnique(t *testing.T) {
	existing := map[string]bool{"alice": true, "alice-2": true}
	check := func(s string) (bool, error) { return existing[s], nil }

	first, err := IsUnique("bob", check)
	if err != nil {
		t.Fatal(err)
	}
	if first != "bob" {
		t.Errorf("expected bob, got %s", first)
	}
	got, err := IsUnique("alice", check)
	if err != nil {
		t.Fatal(err)
	}
	if got != "alice-3" { // alice 和 alice-2 都占用了
		t.Errorf("expected alice-3, got %s", got)
	}
}

func BenchmarkGenerate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Generate("", "some fallback username With spaces@example.com")
	}
}
