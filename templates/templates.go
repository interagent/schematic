package templates

import "text/template"

var templates = map[string]string{"field.tmpl": `{{initialCap .Name}} {{.Type}} {{fieldTag .Name .Required}} {{asComment .Definition.Description}}
`,
	"funcs.tmpl": `{{$Name := .Name}}
{{$Def := .Definition}}
{{range .Definition.Links}}
  {{if .AcceptsCustomType}}
   type {{paramType $Name .}} {{linkGoType .}}
  {{end}}

  {{if (defineCustomType $Def .)}}
   type {{returnType $Name $Def .}} {{$Def.ReturnedGoType .}}
  {{else if (defineCustomArray $Def .)}}
   type {{returnType $Name $Def .}} {{.TargetSchema.Items.GoType}}
  {{end}}

  {{asComment .Description}}
  func (s *Service) {{printf "%s-%s" $Name .Title | initialCap}}({{params $Name .}}) ({{values $Name $Def .}}) {
    {{if ($Def.EmptyResult .)}}
      return s.{{methodCap .Method}}(nil, fmt.Sprintf("{{.HRef}}", {{args .HRef}}){{requestParams .}})
    {{else}}
      {{$Var := initialLow $Name}}var {{$Var}} {{if ($Def.ReturnsArray .)}}[]{{end}}{{returnType $Name $Def .}}
      return {{if ($Def.ReturnsCustomType .)}}&{{end}}{{$Var}}, s.{{methodCap .Method}}(&{{$Var}}, fmt.Sprintf("{{.HRef}}", {{args .HRef}}){{requestParams .}})
    {{end}}
  }
{{end}}

`,
	"imports.tmpl": `{{if .}}
  {{if len . | eq 1}}
    import {{range .}}"{{.}}"{{end}}
  {{else}}
    import (
      {{range .}}
      	"{{.}}"
      {{end}}
    )
  {{end}}
{{end}}`,
	"package.tmpl": `// Generated service client for {{.}} API.
//
// To be able to interact with this API, you have to
// create a new service:
//
//     s := {{.}}.NewService(nil)
//
// The Service struct has all the methods you need
// to interact with {{.}} API.
//
package {{.}}
`,
	"service.tmpl": `const (
	Version          = "{{.Version}}"
	DefaultUserAgent = "{{.Name}}/" + Version + " (" + runtime.GOOS + "; " + runtime.GOARCH + ")"
	DefaultURL       = "{{.URL}}"
)

// Service represents your API.
type Service struct {
	client *http.Client
	URL string
}

// NewService creates a Service using the given, if none is provided
// it uses http.DefaultClient.
func NewService(c *http.Client) *Service {
	if c == nil {
		c = http.DefaultClient
	}
	return &Service{
		client: c,
		URL: DefaultURL,
	}
}

// NewRequest generates an HTTP request, but does not perform the request.
func (s *Service) NewRequest(method, path string, body interface{}, q interface{}) (*http.Request, error) {
	var ctype string
	var rbody io.Reader

	switch t := body.(type) {
	case nil:
	case string:
		rbody = bytes.NewBufferString(t)
	case io.Reader:
		rbody = t
	default:
		v := reflect.ValueOf(body)
		if !v.IsValid() {
			break
		}
		if v.Type().Kind() == reflect.Ptr {
			v = reflect.Indirect(v)
			if !v.IsValid() {
				break
			}
		}

		j, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		rbody = bytes.NewReader(j)
		ctype = "application/json"
	}
	req, err := http.NewRequest(method, s.URL+path, rbody)
	if err != nil {
		return nil, err
	}

	if q != nil {
		v, err := query.Values(q)
		if err != nil {
			return nil, err
		}
		query := v.Encode()
		if req.URL.RawQuery != "" && query != "" {
			req.URL.RawQuery += "&"
		}
		req.URL.RawQuery += query
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", DefaultUserAgent)
	if ctype != "" {
		req.Header.Set("Content-Type", ctype)
	}

	return req, nil
}

// Do sends a request and decodes the response into v.
func (s *Service) Do(v interface{}, method, path string, body interface{}, q interface{}, lr *ListRange) error {
	req, err := s.NewRequest(method, path, body, q)
	if err != nil {
		return err
	}
	if lr != nil {
		lr.SetHeader(req)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch t := v.(type) {
	case nil:
	case io.Writer:
		_, err = io.Copy(t, resp.Body)
	default:
		err = json.NewDecoder(resp.Body).Decode(v)
	}
	return err
}

// Get sends a GET request and decodes the response into v.
func (s *Service) Get(v interface{}, path string, query interface{}, lr *ListRange) error {
	return s.Do(v, "GET", path, nil, query, lr)
}

// Patch sends a Path request and decodes the response into v.
func (s *Service) Patch(v interface{}, path string, body interface{}) error {
	return s.Do(v, "PATCH", path, body, nil, nil)
}

// Post sends a POST request and decodes the response into v.
func (s *Service) Post(v interface{}, path string, body interface{}) error {
	return s.Do(v, "POST", path, body, nil, nil)
}

// Put sends a PUT request and decodes the response into v.
func (s *Service) Put(v interface{}, path string, body interface{}) error {
	return s.Do(v, "PUT", path, body, nil, nil)
}

// Delete sends a DELETE request.
func (s *Service) Delete(v interface{}, path string) error {
	return s.Do(v, "DELETE", path, nil, nil, nil)
}

// ListRange describes a range.
type ListRange struct {
	Field      string
	Max        int
	Descending bool
	FirstID    string
	LastID     string
}

// SetHeader set headers on the given Request.
func (lr *ListRange) SetHeader(req *http.Request) {
	var hdrval string
	if lr.Field != "" {
		hdrval += lr.Field + " "
	}
	hdrval += lr.FirstID + ".." + lr.LastID
	if lr.Max != 0 {
		hdrval += fmt.Sprintf("; max=%d", lr.Max)
		if lr.Descending {
			hdrval += ", "
		}
	}

	if lr.Descending {
		hdrval += ", order=desc"
	}

	req.Header.Set("Range", hdrval)
	return
}

// Bool allocates a new int value returns a pointer to it.
func Bool(v bool) *bool {
	p := new(bool)
	*p = v
	return p
}

// Int allocates a new int value returns a pointer to it.
func Int(v int) *int {
	p := new(int)
	*p = v
	return p
}

// Float64 allocates a new float64 value returns a pointer to it.
func Float64(v float64) *float64 {
	p := new(float64)
	*p = v
	return p
}

// String allocates a new string value returns a pointer to it.
func String(v string) *string {
	p := new(string)
	*p = v
	return p
}
`,
	"struct.tmpl": `{{asComment .Definition.Description}}
{{if (.Definition.Items)}}
type {{initialCap .Name}} {{goType .Definition.Items}}
{{else}}
type {{initialCap .Name}} {{goType .Definition}}
{{end}}
`,
}

func Parse(t *template.Template) (*template.Template, error) {
	for name, s := range templates {
		var tmpl *template.Template
		if t == nil {
			t = template.New(name)
		}
		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}
		if _, err := tmpl.Parse(s); err != nil {
			return nil, err
		}
	}
	return t, nil
}

