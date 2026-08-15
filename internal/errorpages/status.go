package errorpages

// ParseErrorStatus parses a three-digit 4xx or 5xx HTTP status code.
func ParseErrorStatus(status string) (uint16, bool) {
	if len(status) != 3 || (status[0] != '4' && status[0] != '5') {
		return 0, false
	}

	code := uint16(0)
	for index := range status {
		if status[index] < '0' || status[index] > '9' {
			return 0, false
		}
		code = code*10 + uint16(status[index]-'0')
	}

	return code, true
}

// IsErrorStatus reports whether status is a valid three-digit 4xx or 5xx status code.
func IsErrorStatus(status string) bool {
	_, ok := ParseErrorStatus(status)
	return ok
}

func getStatusMessage(code uint16) string {
	messages := map[uint16]string{
		400: "Bad Request",
		401: "Unauthorized",
		402: "Payment Required",
		403: "Forbidden",
		404: "Not Found",
		405: "Method Not Allowed",
		406: "Not Acceptable",
		407: "Proxy Authentication Required",
		408: "Request Timeout",
		409: "Conflict",
		410: "Gone",
		411: "Length Required",
		412: "Precondition Failed",
		413: "Payload Too Large",
		414: "URI Too Long",
		415: "Unsupported Media Type",
		416: "Requested Range Not Satisfiable",
		417: "Expectation Failed",
		418: "I'm a teapot",
		421: "Misdirected Request",
		422: "Unprocessable Entity",
		423: "Locked",
		424: "Failed Dependency",
		425: "Too Early",
		426: "Upgrade Required",
		428: "Precondition Required",
		429: "Too Many Requests",
		431: "Request Header Fields Too Large",
		451: "Unavailable For Legal Reasons",
		500: "Internal Server Error",
		501: "Not Implemented",
		502: "Bad Gateway",
		503: "Service Unavailable",
		504: "Gateway Timeout",
		505: "HTTP Version Not Supported",
		506: "Variant Also Negotiates",
		507: "Insufficient Storage",
		508: "Loop Detected",
		510: "Not Extended",
		511: "Network Authentication Required",
	}

	if message, ok := messages[code]; ok {
		return message
	}
	if code >= 400 && code < 500 {
		return "Client Error"
	}

	return "Server Error"
}

func getStatusDescription(code uint16) string {
	descriptions := map[uint16]string{
		400: "The server did not understand the request",
		401: "The requested page needs a username and a password",
		403: "Access is forbidden to the requested page",
		404: "The server can not find the requested page",
		405: "The method specified in the request is not allowed",
		407: "You must authenticate with a proxy server before this request can be served",
		408: "The request took longer than the server was prepared to wait",
		409: "The request could not be completed because of a conflict",
		410: "The requested page is no longer available",
		411: "The \"Content-Length\" is not defined. The server will not accept the request without it",
		412: "The pre condition given in the request evaluated to false by the server",
		413: "The server will not accept the request, because the request entity is too large",
		416: "The requested byte range is not available and is out of bounds",
		418: "Attempt to brew coffee with a teapot is not supported",
		429: "Too many requests in a given amount of time",
		500: "The server met an unexpected condition",
		502: "The server received an invalid response from the upstream server",
		503: "The server is temporarily overloading or down",
		504: "The gateway has timed out",
		505: "The server does not support the \"http protocol\" version",
	}

	if description, ok := descriptions[code]; ok {
		return description
	}
	if code >= 400 && code < 500 {
		return "An error occurred while processing your request."
	}

	return "The server encountered an error while processing your request."
}
