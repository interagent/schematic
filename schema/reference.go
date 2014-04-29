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

type Reference string

// Resolve reference.
func (rf Reference) Resolve(r *Schema) *Schema {
	if !strings.HasPrefix(string(rf), fragment) {
		panic(fmt.Sprintf("non-fragment reference are not supported : %s", rf))
	}
	var node interface{}
	node = r
	for _, t := range strings.Split(string(rf), separator)[1:] {
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
				panic(fmt.Sprintf("can't find '%s' field in %s", t, rf))
			}
			node = f.Interface()
		case reflect.Map:
			kv := v.MapIndex(reflect.ValueOf(t))
			if !kv.IsValid() {
				panic(fmt.Sprintf("can't find '%s' key in %s", t, rf))
			}
			node = kv.Interface()
		default:
			panic(fmt.Sprintf("can't follow pointer : %s", rf))
		}
	}
	return node.(*Schema)
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

type HRef string

func (h HRef) Resolve(r *Schema) map[string]*Schema {
	schemas := make(map[string]*Schema)
	for _, v := range href.FindAllString(string(h), -1) {
		u, err := url.QueryUnescape(v[2 : len(v)-2])
		if err != nil {
			panic(err)
		}
		parts := strings.Split(u, "/")
		name := initialLow(fmt.Sprintf("%s-%s", parts[len(parts)-3], parts[len(parts)-1]))
		schemas[name] = Reference(u).Resolve(r)
	}
	return schemas
}

func (h HRef) URL() (*url.URL, error) {
	return url.Parse(string(h))
}

func (h HRef) String() string {
	return href.ReplaceAllStringFunc(string(h), func(v string) string {
		return "%v"
	})
}
