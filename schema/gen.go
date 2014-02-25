package schema

import (
	"bytes"
	"fmt"
	"go/format"
	"regexp"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"
)

var templates = template.Must(template.New("package.tmpl").Funcs(helpers).ParseGlob("schema/templates/*.tmpl"))

var helpers = template.FuncMap{
	"initialCap":     initialCap,
	"initialLow":     initialLow,
	"methodCap":      methodCap,
	"validIdentifer": validIdentifer,
	"asComment":      asComment,
	"jsonTag":        jsonTag,
	"params":         params,
}

var (
	newlines = regexp.MustCompile(`(?m:\s*$)`)
	acronyms = regexp.MustCompile(`(?m)(Url|Http|Id|Uuid|Api|Uri|Ssl|Cname|Oauth)`)
)

func (s *Schema) Generate() ([]byte, error) {
	var buf bytes.Buffer

	// TODO: Allow to change name of package.
	templates.ExecuteTemplate(&buf, "package.tmpl", "schema")

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
			panic(fmt.Sprintf("unknow type %s", kind))
		}
	}
	panic("type not found")
}

func (s *Schema) Parameters(l *Link) map[string]string {
	params := make(map[string]string)
	for name, def := range l.HRef.Resolve(s) {
		params[name] = s.GoType(def)
	}
	return params
}

func jsonTag(n string, def *Schema) string {
	tags := []string{n}
	if !contains(n, def.Required) {
		tags = append(tags, "omitempty")
	}
	return fmt.Sprintf("`json:\"%s\"`", strings.Join(tags, ","))
}

func contains(n string, r []string) bool {
	for _, r := range r {
		if r == n {
			return true
		}
	}
	return false
}

func initialCap(ident string) string {
	if ident == "" {
		panic("blank identifier")
	}
	return depunct(ident, true)
}

func methodCap(ident string) string {
	return initialCap(strings.ToLower(ident))
}

func initialLow(ident string) string {
	capitalize := initialCap(ident)
	r, n := utf8.DecodeRuneInString(capitalize)
	return string(unicode.ToLower(r)) + capitalize[n:]
}

func validIdentifer(ident string) string {
	return depunct(ident, false)
}

func depunct(ident string, needCap bool) string {
	var buf bytes.Buffer
	for _, c := range ident {
		if c == '-' || c == '.' || c == '$' || c == '/' || c == ':' || c == '_' || c == ' ' || c == '{' || c == '}' {
			needCap = true
			continue
		}
		if needCap {
			c = unicode.ToUpper(c)
			needCap = false
		}
		buf.WriteByte(byte(c))
	}
	depuncted := acronyms.ReplaceAllFunc(buf.Bytes(), func(m []byte) []byte {
		if len(m) > 4 {
			return append(bytes.ToUpper(m[:2]), m[2:]...)
		}
		return bytes.ToUpper(m)
	})
	return string(depuncted)
}

func asComment(c string) string {
	var buf bytes.Buffer
	const maxLen = 70
	removeNewlines := func(s string) string {
		return strings.Replace(s, "\n", "\n// ", -1)
	}
	for len(c) > 0 {
		line := c
		if len(line) < maxLen {
			fmt.Fprintf(&buf, "// %s\n", removeNewlines(line))
			break
		}
		line = line[:maxLen]
		si := strings.LastIndex(line, " ")
		if si != -1 {
			line = line[:si]
		}
		fmt.Fprintf(&buf, "// %s\n", removeNewlines(line))
		c = c[len(line):]
		if si != -1 {
			c = c[1:]
		}
	}
	return buf.String()
}

func params(m map[string]string) string {
	var p []string
	for k, v := range m {
		p = append(p, fmt.Sprintf("%s %s", k, v))
	}
	return strings.Join(p, ",")
}
