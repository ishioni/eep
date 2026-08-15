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
		416: "Range Not Satisfiable",
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
		400: "The request could not be understood by the server due to malformed syntax.",
		401: "The request requires user authentication.",
		403: "The server understood the request, but is refusing to fulfill it.",
		404: "The requested resource could not be found.",
		405: "The method specified in the request is not allowed for the resource.",
		408: "The server timed out waiting for the request.",
		429: "Too many requests have been sent in a given amount of time.",
		500: "The server encountered an unexpected condition that prevented it from fulfilling the request.",
		502: "The server received an invalid response from the upstream server.",
		503: "The server is currently unable to handle the request due to temporary overloading or maintenance.",
		504: "The server did not receive a timely response from the upstream server.",
	}

	if description, ok := descriptions[code]; ok {
		return description
	}
	if code >= 400 && code < 500 {
		return "An error occurred while processing your request."
	}

	return "The server encountered an error while processing your request."
}
