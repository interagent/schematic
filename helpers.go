package schematic

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"
)

var helpers = template.FuncMap{
	"initialCap":       initialCap,
	"initialLow":       initialLow,
	"methodCap":        methodCap,
	"asComment":        asComment,
	"fieldTag":         fieldTag,
	"params":           params,
	"requestParams":    requestParams,
	"args":             args,
	"values":           values,
	"goType":           goType,
	"linkGoType":       linkGoType,
	"returnType":       returnType,
	"defineCustomType": defineCustomType,
	"paramType":        paramType,
}

var (
	newlines  = regexp.MustCompile(`(?m:\s*$)`)
	acronyms  = regexp.MustCompile(`(Url|Http|Id|Io|Uuid|Api|Uri|Ssl|Cname|Oauth|Otp|Cidr|Nat|Vpn)$`)
	camelcase = regexp.MustCompile(`(?m)[-.$/:_{}\s]+`)
)

func goType(p *Schema) string {
	return p.GoType()
}

func linkGoType(l *Link) string {
	t, _ := l.GoType()
	return t
}

func required(n string, def *Schema) bool {
	return contains(n, def.Required)
}

func fieldTag(n string, required bool) string {
	return fmt.Sprintf("`%s %s`", jsonTag(n, required), urlTag(n, required))
}

func jsonTag(n string, required bool) string {
	tags := []string{n}
	if !required {
		tags = append(tags, "omitempty")
	}
	return fmt.Sprintf("json:\"%s\"", strings.Join(tags, ","))
}

func urlTag(n string, required bool) string {
	tags := []string{n}
	if !required {
		tags = append(tags, "omitempty")
	}
	tags = append(tags, "key")
	return fmt.Sprintf("url:\"%s\"", strings.Join(tags, ","))
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
	if ident == "" {
		panic("blank identifier")
	}
	return depunct(ident, false)
}

func depunct(ident string, initialCap bool) string {
	matches := camelcase.Split(ident, -1)
	for i, m := range matches {
		if initialCap || i > 0 {
			m = capFirst(m)
		}
		matches[i] = acronyms.ReplaceAllStringFunc(m, func(c string) string {
			if len(c) > 4 {
				return strings.ToUpper(c[:2]) + c[2:]
			}
			return strings.ToUpper(c)
		})
	}
	return strings.Join(matches, "")
}

func capFirst(ident string) string {
	r, n := utf8.DecodeRuneInString(ident)
	return string(unicode.ToUpper(r)) + ident[n:]
}

func asComment(c string) string {
	var buf bytes.Buffer
	const maxLen = 70
	r := []rune(c)
	for len(r) > 0 {
		line := r
		if len(line) < maxLen {
			fmt.Fprintf(&buf, "// %s\n", removeNewlines(line))
			break
		}
		line = line[:maxLen]
		si := lastIndex(line, func(r rune) bool {
			return unicode.IsSpace(r)
		})
		if si != -1 {
			line = line[:si]
		}
		fmt.Fprintf(&buf, "// %s\n", removeNewlines(line))
		r = r[len(line):]
		if si != -1 {
			r = r[1:]
		}
	}
	return buf.String()
}

func values(n string, s *Schema, l *Link) string {
	v := s.Values(n, l)
	return strings.Join(v, ", ")
}

func params(name string, l *Link) string {
	var p []string
	order, params := l.Parameters(name)
	for _, n := range order {
		p = append(p, fmt.Sprintf("%s %s", initialLow(n), params[n]))
	}
	return strings.Join(p, ", ")
}

func requestParams(l *Link) string {
	_, params := l.Parameters("")
	if strings.ToUpper(l.Method) == "DELETE" {
		return ""
	}
	p := []string{""}
	if _, ok := params["o"]; ok {
		p = append(p, "o")
	} else {
		p = append(p, "nil")
	}
	if _, ok := params["lr"]; ok {
		p = append(p, "lr")
	} else if strings.ToUpper(l.Method) == "GET" {
		p = append(p, "nil")
	}
	return strings.Join(p, ", ")
}

func args(h *HRef) string {
	return strings.Join(h.Order, ", ")
}

func sortedKeys(m map[string]*Schema) (keys []string) {
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return
}

func returnType(name string, s *Schema, l *Link) string {
	if defineCustomType(s, l) {
		return initialCap(fmt.Sprintf("%s-%s-Result", name, l.Title))
	}
	return initialCap(name)
}

func paramType(name string, l *Link) string {
	if l.AcceptsCustomType() {
		return initialCap(fmt.Sprintf("%s-%s-Opts", name, l.Title))
	}
	return initialCap(name)
}

func defineCustomType(s *Schema, l *Link) bool {
	return l.TargetSchema != nil && l.TargetSchema != s
}

func removeNewlines(s []rune) string {
	return strings.Replace(string(s), "\n", "\n// ", -1)
}

func lastIndex(s []rune, f func(rune) bool) int {
	for i := len(s) - 1; i > 0; i-- {
		if f(s[i]) {
			return i
		}
	}
	return -1
}
