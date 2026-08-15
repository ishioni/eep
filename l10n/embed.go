package l10n

import _ "embed"

//go:generate go run ./generate/localize.go -locales ./locales.json -out ./localize.js -out-min localize.min.js -out-go ./supported_locales.go

//go:embed localize.min.js
var content string

// ScriptContent returns the generated client-side localization runtime.
func ScriptContent() string { return content }
