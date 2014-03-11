package schema

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"
)

const (
	fragment  = "#"
	separator = "/"
)

var href = regexp.MustCompile(`({\([^\)]+)\)}`)

type Reference struct {
	ref string
}

// Create a new Reference.
func NewReference(ref string) *Reference {
	return &Reference{ref}
}

// Resolve reference.
func (r *Reference) Resolve(s *Schema) *Schema {
	if !strings.HasPrefix(r.ref, fragment) {
		panic("non-fragment reference are not supported")
	}
	var node interface{}
	node = s
	for _, t := range strings.Split(r.ref, separator)[1:] {
		t = decode(t)
		v := reflect.Indirect(reflect.ValueOf(node))
		switch v.Kind() {
		case reflect.Struct:
			var f reflect.Value
			for i := 0; i < v.NumField(); i++ {
				f = v.Field(i)
				ft := v.Type().Field(i)
				tag := ft.Tag.Get("json")
				if tag == "-" {
					continue
				}
				name := parseTag(tag)
				if name == "" {
					name = ft.Name
				}
				if name == t {
					break
				}
			}
			if !f.IsValid() {
				panic(fmt.Sprintf("can't find '%s' field in %s", t, r.ref))
			}
			node = f.Interface()
		case reflect.Map:
			kv := v.MapIndex(reflect.ValueOf(t))
			if !kv.IsValid() {
				panic(fmt.Sprintf("can't find '%s' key in %s", t, r.ref))
			}
			node = kv.Interface()
		default:
			panic("can't follow pointer")
		}
	}
	return node.(*Schema)
}

func (r *Reference) UnmarshalJSON(data []byte) error {
	r.ref = string(data[1 : len(data)-1])
	return nil
}

func (r *Reference) MarshalJSON() ([]byte, error) {
	return []byte(r.ref), nil
}

func (r *Reference) String() string {
	return r.ref
}

func encode(t string) (encoded string) {
	encoded = strings.Replace(t, "/", "~1", -1)
	return strings.Replace(encoded, "~", "~0", -1)
}

func decode(t string) (decoded string) {
	decoded = strings.Replace(t, "~1", "/", -1)
	return strings.Replace(decoded, "~0", "~", -1)
}

func parseTag(tag string) string {
	if i := strings.Index(tag, ","); i != -1 {
		return tag[:i]
	}
	return tag
}

type HRef struct {
	href string
}

func NewHRef(href string) *HRef {
	return &HRef{href}
}

func (h *HRef) UnmarshalJSON(data []byte) error {
	h.href = string(data[1 : len(data)-1])
	return nil
}

func (h *HRef) MarshalJSON() ([]byte, error) {
	return []byte(h.href), nil
}

func (h *HRef) Resolve(s *Schema) map[string]*Schema {
	schemas := make(map[string]*Schema)
	for _, v := range href.FindAllString(h.href, -1) {
		u, err := url.QueryUnescape(v[2 : len(v)-2])
		if err != nil {
			panic(err)
		}
		parts := strings.Split(u, "/")
		schemas[parts[len(parts)-1]] = NewReference(u).Resolve(s)
	}
	return schemas
}

func (h *HRef) String() string {
	return href.ReplaceAllStringFunc(h.href, func(v string) string {
		return "%v"
	})
}
