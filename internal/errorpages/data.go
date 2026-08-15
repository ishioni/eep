package errorpages

// Data is the upstream-compatible data model exposed to error page templates.
// Keep existing fields and their types aligned with error-pages' template data model.
type Data struct {
	StatusCode   uint16
	Message      string
	Description  string
	OriginalURI  string
	Namespace    string
	IngressName  string
	ServiceName  string
	ServicePort  string
	RequestID    string
	ForwardedFor string
	Host         string
	HomepageURL  string
	Links        []Link
	Config       TemplateConfig
}

// Link is a labeled hyperlink exposed to templates.
type Link struct {
	Label string
	URL   string
}

// TemplateConfig contains configuration values exposed to templates.
type TemplateConfig struct {
	ShowRequestDetails bool
	L10nDisabled       bool
}

func (d Data) withDefaults() Data {
	if d.Message == "" {
		d.Message = getStatusMessage(d.StatusCode)
	}
	if d.Description == "" {
		d.Description = getStatusDescription(d.StatusCode)
	}

	return d
}
