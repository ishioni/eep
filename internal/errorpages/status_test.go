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

func TestLocalizedStatusDefaultsMatchUpstream(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		code        uint16
		message     string
		description string
	}{
		{code: 400, message: "Bad Request", description: "The server did not understand the request"},
		{code: 401, message: "Unauthorized", description: "The requested page needs a username and a password"},
		{code: 403, message: "Forbidden", description: "Access is forbidden to the requested page"},
		{code: 404, message: "Not Found", description: "The server can not find the requested page"},
		{code: 405, message: "Method Not Allowed", description: "The method specified in the request is not allowed"},
		{
			code:        407,
			message:     "Proxy Authentication Required",
			description: "You must authenticate with a proxy server before this request can be served",
		},
		{
			code:        408,
			message:     "Request Timeout",
			description: "The request took longer than the server was prepared to wait",
		},
		{code: 409, message: "Conflict", description: "The request could not be completed because of a conflict"},
		{code: 410, message: "Gone", description: "The requested page is no longer available"},
		{
			code:        411,
			message:     "Length Required",
			description: "The \"Content-Length\" is not defined. The server will not accept the request without it",
		},
		{
			code:        412,
			message:     "Precondition Failed",
			description: "The pre condition given in the request evaluated to false by the server",
		},
		{
			code:        413,
			message:     "Payload Too Large",
			description: "The server will not accept the request, because the request entity is too large",
		},
		{
			code:        416,
			message:     "Requested Range Not Satisfiable",
			description: "The requested byte range is not available and is out of bounds",
		},
		{code: 418, message: "I'm a teapot", description: "Attempt to brew coffee with a teapot is not supported"},
		{code: 429, message: "Too Many Requests", description: "Too many requests in a given amount of time"},
		{code: 500, message: "Internal Server Error", description: "The server met an unexpected condition"},
		{
			code:        502,
			message:     "Bad Gateway",
			description: "The server received an invalid response from the upstream server",
		},
		{code: 503, message: "Service Unavailable", description: "The server is temporarily overloading or down"},
		{code: 504, message: "Gateway Timeout", description: "The gateway has timed out"},
		{
			code:        505,
			message:     "HTTP Version Not Supported",
			description: "The server does not support the \"http protocol\" version",
		},
	} {
		got := (Data{StatusCode: test.code}).withDefaults()
		if got.Message != test.message || got.Description != test.description {
			t.Errorf(
				"defaults for %d = (%q, %q), want (%q, %q)",
				test.code,
				got.Message,
				got.Description,
				test.message,
				test.description,
			)
		}
	}
}
