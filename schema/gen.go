package schema

import (
	"bytes"
	"fmt"
	bundle "github.com/heroku/schematic/schema/templates"
	"go/format"
	"regexp"
	"strings"
	"text/template"
)

var templates *template.Template

var (
	newlines = regexp.MustCompile(`(?m:\s*$)`)
	acronyms = regexp.MustCompile(`(?m)(Url|Http|Id|Uuid|Api|Uri|Ssl|Cname|Oauth)`)
)

func init() {
	templates = template.New("package.tmpl").Funcs(helpers)
	templates.ParseGlob("templates/*.tmpl")
	bundle.Parse(templates)
}

// Generate code according to the schema.
func (s *Schema) Generate() ([]byte, error) {
	var buf bytes.Buffer

	name := strings.ToLower(strings.Split(s.Title, " ")[0])
	templates.ExecuteTemplate(&buf, "package.tmpl", name)

	// TODO: Check if we need time.
	templates.ExecuteTemplate(&buf, "imports.tmpl", []string{"time", "fmt"})

	for name, def := range s.Definitions {
		// Skipping definitions because there is no links, nor properties.
		if def.Links == nil && def.Properties == nil {
			continue
		}

		context := struct {
			Name       string
			Definition *Schema
			Root       *Schema
		}{
			Name:       name,
			Definition: def,
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

	for _, kind := range types {
		switch kind {
		case "boolean":
			return "bool"
		case "string":
			switch prop.Format {
			case "date-time":
				return "time.Time"
			default:
				return "string"
			}
		case "number":
			return "float64"
		case "integer":
			return "int64"
		case "any":
			return "interface{}"
		case "array":
			return "[]" + s.GoType(prop.Items)
		case "object":
			// Check if additionalProperties is false.
			if m, ok := prop.AdditionalProperties.(bool); ok && !m {
				return "map[string]string"
			}
			var buf bytes.Buffer
			templates.ExecuteTemplate(&buf, "astruct.tmpl", struct {
				Definition *Schema
				Root       *Schema
			}{
				Definition: prop,
				Root:       s,
			})
			return buf.String()
		case "null":
			continue
		default:
			panic(fmt.Sprintf("unknown type %s", kind))
		}
	}
	panic(fmt.Sprintf("type not found : %s", types))
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
		params["r"] = "*ListRange"
	}
	return params
}

// Return function return values types.
func (s *Schema) Values(name string, l *Link) []string {
	var values []string
	name = initialCap(name)
	switch l.Rel {
	case "destroy":
		values = append(values, "error")
	case "instances":
		values = append(values, fmt.Sprintf("[]*%s", name), "error")
	default:
		values = append(values, fmt.Sprintf("*%s", name), "error")
	}
	return values
}

func (l *Link) GoType(r *Schema) string {
	// FIXME: Arguments are reverse from Schema.GoType()
	return r.GoType(l.Schema)
}
