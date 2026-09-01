package preview

import "testing"

func TestOpenEmpty(t *testing.T) {
	t.Parallel()
	if err := Open(); err != nil {
		t.Fatalf("Open() with no paths error = %v", err)
	}
}
