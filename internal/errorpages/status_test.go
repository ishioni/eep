package errorpages

import "testing"

func TestParseErrorStatus(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		status string
		want   uint16
		ok     bool
	}{
		{name: "client error", status: "404", want: 404, ok: true},
		{name: "server error", status: "503", want: 503, ok: true},
		{name: "success", status: "200"},
		{name: "non-digit", status: "4xx"},
		{name: "too short", status: "50"},
		{name: "too long", status: "5000"},
		{name: "empty"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, ok := ParseErrorStatus(test.status)
			if got != test.want || ok != test.ok {
				t.Fatalf("ParseErrorStatus(%q) = (%d, %v), want (%d, %v)", test.status, got, ok, test.want, test.ok)
			}
			if IsErrorStatus(test.status) != test.ok {
				t.Fatalf("IsErrorStatus(%q) = %v, want %v", test.status, !test.ok, test.ok)
			}
		})
	}
}
