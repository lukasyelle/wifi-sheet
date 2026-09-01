package networks

import "testing"

func TestParseWifiDevice(t *testing.T) {
	t.Parallel()

	output := `
Hardware Port: Ethernet Adapter (en4)
Device: en4
Ethernet Address: 7e:7f:ec:da:78:3c

Hardware Port: Wi-Fi
Device: en0
Ethernet Address: 5c:e9:1e:bc:a7:00

Hardware Port: Thunderbolt 1
Device: en1
Ethernet Address: 36:20:7e:d8:9b:80
`
	got, err := parseWifiDevice(output)
	if err != nil {
		t.Fatalf("parseWifiDevice() error = %v", err)
	}
	if got != "en0" {
		t.Fatalf("parseWifiDevice() = %q, want en0", got)
	}
}

func TestParseWifiDeviceMissing(t *testing.T) {
	t.Parallel()
	_, err := parseWifiDevice("Hardware Port: Ethernet\nDevice: en4\n")
	if err == nil {
		t.Fatal("parseWifiDevice() expected error")
	}
}

func TestParsePreferredNetworks(t *testing.T) {
	t.Parallel()

	output := "Preferred networks on en0:\n\tHome\n\tGuest\n\tLukas’ iPhone\n"
	got := parsePreferredNetworks(output)
	want := []string{"Home", "Guest", "Lukas’ iPhone"}
	if len(got) != len(want) {
		t.Fatalf("parsePreferredNetworks() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("parsePreferredNetworks()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestParseCurrentSSID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		output string
		want   string
	}{
		{
			name:   "present",
			output: "  BSSID : aa:bb\n  SSID : Home\n  IPv4 : ...\n",
			want:   "Home",
		},
		{
			name:   "redacted",
			output: "  SSID : <redacted>\n",
			want:   "",
		},
		{
			name:   "missing",
			output: "  BSSID : aa:bb\n",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := parseCurrentSSID(tt.output); got != tt.want {
				t.Fatalf("parseCurrentSSID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMergeCurrent(t *testing.T) {
	t.Parallel()

	got := mergeCurrent([]string{"Guest", "Home", "Cafe"}, "Home")
	want := []string{"Home", "Guest", "Cafe"}
	if len(got) != len(want) {
		t.Fatalf("mergeCurrent() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("mergeCurrent()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
