package runner

import "testing"

func TestSplitArgs(t *testing.T) {
	got := splitArgs(`-server "72.56.68.200:9443" -echo "hello world" -flag`)
	want := []string{"-server", "72.56.68.200:9443", "-echo", "hello world", "-flag"}
	if len(got) != len(want) {
		t.Fatalf("len mismatch: got %#v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg %d: got %q want %q", i, got[i], want[i])
		}
	}
}
