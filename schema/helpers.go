package schema

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"
)

var helpers = template.FuncMap{
	"initialCap":     initialCap,
	"initialLow":     initialLow,
	"methodCap":      methodCap,
	"validIdentifer": validIdentifer,
	"asComment":      asComment,
	"jsonTag":        jsonTag,
	"params":         params,
	"args":           args,
	"join":           join,
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

func join(p []string) string {
	return strings.Join(p, ",")
}

func params(m map[string]string) string {
	var p []string
	for k, v := range m {
		p = append(p, fmt.Sprintf("%s %s", k, v))
	}
	return strings.Join(p, ",")
}

func args(m map[string]*Schema) string {
	var p []string
	for k, _ := range m {
		p = append(p, k)
	}
	return strings.Join(p, ", ")
}
