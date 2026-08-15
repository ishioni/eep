package errorpages

import (
	"testing"
	"time"
)

func TestCurrentTemplateFunctionCompatibility(t *testing.T) {
	t.Setenv("TEST_ENV_VAR", "unit-test")
	t.Setenv("SECRET_ENV_VAR", "supersecret")

	dependencies := templateDependencies{
		now:      func() time.Time { return time.Unix(1710000000, 0) },
		hostname: func() string { return "test-host" },
	}

	for name, test := range map[string]struct {
		source string
		want   string
	}{
		"now":              {source: `{{ now.Unix }}`, want: "1710000000"},
		"hostname":         {source: `{{ hostname }}`, want: "test-host"},
		"toJson":           {source: `{{ "test" | toJson }}`, want: `"test"`},
		"toJSON":           {source: `{{ 42 | toJSON }}`, want: "42"},
		"int boolean":      {source: `{{ true | int }}`, want: "1"},
		"int float string": {source: `{{ "3.14" | toInt }}`, want: "3"},
		"version":          {source: `{{ version }}`, want: "test-version"},
		"env":              {source: `{{ env "TEST_ENV_VAR" }}`, want: "unit-test"},
		"env secret":       {source: `{{ env "SECRET_ENV_VAR" }}`, want: "***********"},
		"escape":           {source: `{{ escape "<tag>" }}`, want: "&lt;tag&gt;"},
		"trim":             {source: `{{ "  test  " | trim }}`, want: "test"},
		"trimPrefix":       {source: `{{ "test" | trimPrefix "te" }}`, want: "st"},
		"trimSuffix":       {source: `{{ "test" | trimSuffix "st" }}`, want: "te"},
		"trimPostfix":      {source: `{{ "test" | trimPostfix "st" }}`, want: "te"},
		"replace":          {source: `{{ "test" | replace "t" "z" }}`, want: "zesz"},
		"contains":         {source: `{{ "test" | contains "es" }}`, want: "true"},
		"count":            {source: `{{ "test" | count "t" }}`, want: "2"},
		"fields":           {source: `{{ "foo bar" | fields }}`, want: "[foo bar]"},
		"lower":            {source: `{{ "TEST" | lower }}`, want: "test"},
		"upper":            {source: `{{ "test" | upper }}`, want: "TEST"},
		"default":          {source: `{{ "" | default "fallback" }}`, want: "fallback"},
		"hasPrefix":        {source: `{{ "test" | hasPrefix "te" }}`, want: "true"},
		"hasSuffix":        {source: `{{ "test" | hasSuffix "st" }}`, want: "true"},
		"hasPostfix":       {source: `{{ "test" | hasPostfix "st" }}`, want: "true"},
		"join and split":   {source: `{{ "a,b,c" | split "," | join "-" }}`, want: "a-b-c"},
		"quote":            {source: `{{ "test" | quote }}`, want: `"test"`},
		"squote":           {source: `{{ "test" | squote }}`, want: "'test'"},
		"repeat":           {source: `{{ "Ha" | repeat 3 }}`, want: "HaHaHa"},
		"substr":           {source: `{{ "Привет" | substr 1 4 }}`, want: "риве"},
		"toString":         {source: `{{ 3.14 | toString }}`, want: "3.14"},
		"str":              {source: `{{ true | str }}`, want: "true"},
		"ternary":          {source: `{{ true | ternary "yes" "no" }}`, want: "yes"},
		"coalesce":         {source: `{{ coalesce "" "value" }}`, want: "value"},
		"urlEncode":        {source: `{{ "/api/v1" | urlEncode }}`, want: "%2Fapi%2Fv1"},
		"isEmpty":          {source: `{{ isEmpty 0 }}`, want: "true"},
		"isNotEmpty":       {source: `{{ isNotEmpty "value" }}`, want: "true"},
		"truncate":         {source: `{{ "Hello, World!" | truncate 8 }}`, want: "Hello..."},
		"trimAll":          {source: `{{ ".....test....." | trimAll "." }}`, want: "test"},
		"l10nScript":       {source: `{{ l10nScript }}`, want: ""},
	} {
		t.Run(name, func(t *testing.T) {
			parsed, err := newTemplate([]byte(test.source), "test-version", dependencies)
			if err != nil {
				t.Fatalf("newTemplate() error = %v", err)
			}

			rendered, err := parsed.Render(Data{})
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			if got := string(rendered); got != test.want {
				t.Fatalf("Render() = %q, want %q", got, test.want)
			}
		})
	}
}
