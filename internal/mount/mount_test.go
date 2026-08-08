package mount

import "testing"

func TestUnescape(t *testing.T) {
	got := unescape(`/srv/my\040disk\011data\134file`)
	if got != "/srv/my disk\tdata\\file" {
		t.Fatalf("unexpected mount path: %q", got)
	}
}
