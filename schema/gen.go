package schema

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"
	"text/template"

	bundle "github.com/heroku/schematic/schema/templates"
)

var templates *template.Template

func init() {
	templates = template.New("package.tmpl").Funcs(helpers)
	templates = template.Must(bundle.Parse(templates))
}

// Generate code according to the schema.
func (s *Schema) Generate() ([]byte, error) {
	var buf bytes.Buffer

	name := strings.ToLower(strings.Split(s.Title, " ")[0])
	templates.ExecuteTemplate(&buf, "package.tmpl", name)

	// TODO: Check if we need time.
	templates.ExecuteTemplate(&buf, "imports.tmpl", []string{
		"encoding/json", "fmt", "io", "reflect",
		"net/http", "runtime", "time", "bytes",
	})
	templates.ExecuteTemplate(&buf, "service.tmpl", struct {
		Name    string
		URL     string
		Version string
	}{
		Name:    name,
		URL:     s.URL(),
		Version: s.Version,
	})

	for name, schema := range s.Properties {
		schema := s.Resolve(schema)
		// Skipping definitions because there is no links, nor properties.
		if schema.Links == nil && schema.Properties == nil {
			continue
		}

		context := struct {
			Name       string
			Definition *Schema
			Root       *Schema
		}{
			Name:       name,
			Definition: schema,
			Root:       s,
		}

		templates.ExecuteTemplate(&buf, "struct.tmpl", context)
		templates.ExecuteTemplate(&buf, "funcs.tmpl", context)
	}

	// Remove blank lines added by text/template
	bytes := newlines.ReplaceAll(buf.Bytes(), []byte(""))

	// Format sources
	clean, err := format.Source(bytes)
	if err != nil {
		return buf.Bytes(), err
	}
	return clean, nil
}

// Resolve reference inside the schema.
func (s *Schema) Resolve(p *Schema) *Schema {
	if p.Ref != nil {
		p = p.Ref.Resolve(s)
	}
	if len(p.OneOf) > 0 {
		p = p.OneOf[0].Ref.Resolve(s)
	}
	if len(p.AnyOf) > 0 {
		p = p.AnyOf[0].Ref.Resolve(s)
	}
	return p
}

// Return Go type for the given schema as string.
func (s *Schema) GoType(p *Schema) string {
	prop := s.Resolve(p)

	var types []string
	if arr, ok := prop.Type.([]interface{}); ok {
		for _, v := range arr {
			types = append(types, v.(string))
		}
	} else {
		types = append(types, prop.Type.(string))
	}

	var goType string
	var nullable bool
	for _, kind := range types {
		switch kind {
		case "boolean":
			goType = "bool"
		case "string":
			switch prop.Format {
			case "date-time":
				goType = "time.Time"
			default:
				goType = "string"
			}
		case "number":
			goType = "float64"
		case "integer":
			goType = "int64"
		case "any":
			goType = "interface{}"
		case "array":
			goType = "[]" + s.GoType(prop.Items)
		case "object":
			// Check if additionalProperties is false.
			if m, ok := prop.AdditionalProperties.(bool); ok && !m {
				goType = "map[string]string"
				continue
			}
			var buf bytes.Buffer
			templates.ExecuteTemplate(&buf, "astruct.tmpl", struct {
				Definition *Schema
				Root       *Schema
			}{
				Definition: prop,
				Root:       s,
			})
			goType = buf.String()
		case "null":
			nullable = true
		default:
			panic(fmt.Sprintf("unknown type %s", kind))
		}
	}
	if goType == "" {
		panic(fmt.Sprintf("type not found : %s", types))
	}
	if nullable {
		return "*" + goType
	}
	return goType
}

// Return function parameters names and types.
func (s *Schema) Parameters(l *Link) map[string]string {
	params := make(map[string]string)
	if l.HRef == nil {
		// No HRef property
		goto Rel
	}
	for name, def := range l.HRef.Resolve(s) {
		params[name] = s.GoType(def)
	}
Rel:
	switch l.Rel {
	case "update", "create":
		params["o"] = l.GoType(s)
	case "instances":
		params["lr"] = "*ListRange"
	}
	return params
}

// Return function return values types.
func (s *Schema) Values(name string, l *Link) []string {
	var values []string
	name = initialCap(name)
	switch l.Rel {
	case "destroy", "empty":
		values = append(values, "error")
	case "instances":
		values = append(values, fmt.Sprintf("[]*%s", name), "error")
	default:
		values = append(values, fmt.Sprintf("*%s", name), "error")
	}
	return values
}

func (s *Schema) URL() string {
	for _, l := range s.Links {
		if l.Rel == "self" {
			return l.HRef.String()
		}
	}
	return ""
}

// Return Go type for the given schema as string.
func (l *Link) GoType(r *Schema) string {
	if l.Schema.Type == nil {
		l.Schema.Type = "object"
	}
	return r.GoType(l.Schema)
}
