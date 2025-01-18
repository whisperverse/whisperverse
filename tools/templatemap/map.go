package templatemap

import (
	"bytes"
	"encoding/json"
	"text/template"

	"github.com/benpate/derp"
)

type Map map[string]*template.Template

// Execute executes the named template with the specified value, and returns the result as a string.
func (m Map) Execute(name string, value any) string {

	if template, exists := m[name]; exists {
		var buffer bytes.Buffer
		if err := template.Execute(&buffer, value); err == nil {
			return buffer.String()
		}
	}

	return ""
}

func (m *Map) UnmarshalJSON(data []byte) error {

	const location = "tools.templatemap.UnmarshalJSON"

	temp := make(map[string]string)

	if err := json.Unmarshal(data, &temp); err != nil {
		return derp.Wrap(err, location, "Error unmarshalling JSON")
	}

	for key, value := range temp {
		tmpl, err := template.New(key).Parse(value)

		if err != nil {
			return derp.Wrap(err, location, "Error parsing template", key)
		}

		(*m)[key] = tmpl
	}

	return nil
}