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
func (r *Schema) Generate() ([]byte, error) {
	var buf bytes.Buffer

	name := strings.ToLower(strings.Split(r.Title, " ")[0])
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
		URL:     r.URL(),
		Version: r.Version,
	})

	for name, schema := range r.Properties {
		schema := r.Resolve(schema)
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
			Root:       r,
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
func (r *Schema) Resolve(s *Schema) *Schema {
	if s.Ref != nil {
		s = s.Ref.Resolve(r)
	}
	if len(s.OneOf) > 0 {
		s = s.OneOf[0].Ref.Resolve(r)
	}
	if len(s.AnyOf) > 0 {
		s = s.AnyOf[0].Ref.Resolve(r)
	}
	return s
}

func (r *Schema) Types() (types []string) {
	if arr, ok := r.Type.([]interface{}); ok {
		for _, v := range arr {
			types = append(types, v.(string))
		}
	} else {
		types = append(types, r.Type.(string))
	}
	return types
}

// Return Go type for the given schema as string.
func (r *Schema) GoType(s *Schema) string {
	return r.goType(s, true, true)
}

func (r *Schema) goType(s *Schema, required bool, force bool) (goType string) {
	// Resolve JSON reference/pointer
	def := r.Resolve(s)

	types := def.Types()
	for _, kind := range types {
		switch kind {
		case "boolean":
			goType = "bool"
		case "string":
			switch def.Format {
			case "date-time":
				goType = "time.Time"
			default:
				goType = "string"
			}
		case "number":
			goType = "float64"
		case "integer":
			goType = "int"
		case "any":
			goType = "interface{}"
		case "array":
			goType = "[]" + r.goType(def.Items, required, force)
		case "object":
			// Check if additionalProperties is false.
			if a, ok := def.AdditionalProperties.(bool); ok && !a {
				if def.PatternProperties != nil {
					goType = "map[string]string"
					continue
				}
			}
			buf := bytes.NewBufferString("struct {")
			for name, prop := range def.Properties {
				req := contains(name, def.Required) || force
				templates.ExecuteTemplate(buf, "field.tmpl", struct {
					Definition *Schema
					Name       string
					Required   bool
					Type       string
				}{
					Definition: prop,
					Name:       name,
					Required:   req,
					Type:       r.goType(prop, req, force),
				})
			}
			buf.WriteString("}")
			goType = buf.String()
		case "null":
			continue
		default:
			panic(fmt.Sprintf("unknown type %s", kind))
		}
	}
	if goType == "" {
		panic(fmt.Sprintf("type not found : %s", types))
	}
	// Types allow null
	if contains("null", types) || !(required || force) {
		return "*" + goType
	}
	return goType
}

// Return function parameters names and types.
func (r *Schema) Parameters(l *Link) map[string]string {
	params := make(map[string]string)
	if l.HRef == nil {
		// No HRef property
		panic(fmt.Errorf("no href property declared for %s", l.Title))
	}
	for name, def := range l.HRef.Resolve(r) {
		params[name] = r.GoType(def)
	}
	switch l.Rel {
	case "update", "create":
		params["o"] = l.GoType(r)
	case "instances":
		params["lr"] = "*ListRange"
	}
	return params
}

// Return function return values types.
func (r *Schema) Values(name string, l *Link) []string {
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

// Return base URL
func (r *Schema) URL() string {
	for _, l := range r.Links {
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
	return r.goType(l.Schema, true, false)
}
