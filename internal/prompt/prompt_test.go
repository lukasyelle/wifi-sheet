package prompt

import "testing"

func TestParseSelection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		count   int
		want    []int
		wantErr bool
	}{
		{name: "single", input: "2", count: 5, want: []int{1}},
		{name: "comma list", input: "1,3,5", count: 5, want: []int{0, 2, 4}},
		{name: "spaces", input: "1 3", count: 5, want: []int{0, 2}},
		{name: "mixed", input: " 2, 2 , 4 ", count: 5, want: []int{1, 3}},
		{name: "all", input: "all", count: 3, want: []int{0, 1, 2}},
		{name: "all caps", input: "ALL", count: 2, want: []int{0, 1}},
		{name: "empty", input: "  ", count: 3, wantErr: true},
		{name: "zero", input: "0", count: 3, wantErr: true},
		{name: "too high", input: "4", count: 3, wantErr: true},
		{name: "garbage", input: "home", count: 3, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseSelection(tt.input, tt.count)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseSelection(%q) expected error, got %v", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseSelection(%q) error = %v", tt.input, err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("ParseSelection(%q) = %v, want %v", tt.input, got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("ParseSelection(%q) = %v, want %v", tt.input, got, tt.want)
				}
			}
		})
	}
}
