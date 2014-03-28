package templates

import "text/template"

var templates = map[string]string{"astruct.tmpl": `{{$Root := .Root}} struct {
  {{range $Name, $Definition := .Definition.Properties}}
    {{initialCap $Name}} {{goType $Root $Definition}} {{jsonTag $Name $Definition}} {{asComment $Definition.Description}}
  {{end}}
}`,
	"funcs.tmpl": `{{$Name := .Name}}
{{$Root := .Root}}
{{range .Definition.Links}}
  {{asComment .Description}}
  func (s *Service) {{printf "%s-%s" $Name .Title | initialCap}}({{params $Root .}}) ({{values $Root $Name .}}) {
    {{if eq .Rel "destroy"}}
      return s.Delete(fmt.Sprintf("{{.HRef}}", {{args $Root .HRef}}))
    {{else if eq .Rel "self"}}
      {{$Var := initialLow $Name}}var {{$Var}} {{initialCap $Name}}
      return &{{$Var}}, s.Get(&{{$Var}}, fmt.Sprintf("{{.HRef}}", {{args $Root .HRef}}))
    {{else if eq .Rel "instances"}}
      req, err := s.NewRequest("GET", fmt.Sprintf("{{.HRef}}", {{args $Root .HRef}}), nil)
      if err != nil {
        return nil, err
      }

      if lr != nil {
        lr.SetHeader(req)
      }
      {{$Var := printf "%s-%s" $Name "List" | initialLow}}
      var {{$Var}} []*{{initialCap $Name}}
      return {{$Var}}, s.Client.Do(req, &{{$Var}})
    {{else if eq .Rel "empty"}}
      return s.{{methodCap .Method}}(fmt.Sprintf("{{.HRef}}", {{args $Root .HRef}}))
    {{else}}
      {{$Var := initialLow $Name}}var {{$Var}} {{initialCap $Name}}
      return &{{$Var}}, s.{{methodCap .Method}}(&{{$Var}}, fmt.Sprintf("{{.HRef}}", {{args $Root .HRef}}), o)
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
	"package.tmpl": `package {{.}}
`,
	"service.tmpl": `const (
	Version          = "{{.Version}}"
	DefaultAPIURL    = "{{.URL}}"
	DefaultUserAgent = "{{.Name}}/" + Version + " (" + runtime.GOOS + "; " + runtime.GOARCH + ")"
)

// Client is the interface that wraps HTTP request creation
// and execution.
type Client interface {
	Do(req *http.Request, v interface{}) error
	NewRequest(method, path string, body interface{}) (*http.Request, error)
}

// Service represents your API.
type Service struct {
	Client Client
}

// Create a Service using the given, if none is provided
// it uses DefaultClient.
func NewService(client Client) *Service {
	if client == nil {
		client = &DefaultClient{}
	}
	return &Service{
		Client: client,
	}
}

// Generates an HTTP request, but does not perform the request.
func (s *Service) NewRequest(method, path string, body interface{}) (*http.Request, error) {
	return s.Client.NewRequest(method, path, body)
}

// Sends a request and decodes the response into v.
func (s *Service) Do(v interface{}, method, path string, body interface{}) error {
	req, err := s.Client.NewRequest(method, path, body)
	if err != nil {
		return err
	}
	return s.Client.Do(req, v)
}

func (s *Service) Get(v interface{}, path string) error {
	return s.Do(v, "GET", path, nil)
}

func (s *Service) Patch(v interface{}, path string, body interface{}) error {
	return s.Do(v, "PATCH", path, body)
}

func (s *Service) Post(v interface{}, path string, body interface{}) error {
	return s.Do(v, "POST", path, body)
}

func (s *Service) Put(v interface{}, path string, body interface{}) error {
	return s.Do(v, "PUT", path, body)
}

func (s *Service) Delete(path string) error {
	return s.Do(nil, "DELETE", path, nil)
}

// DefaultClient implements Client interface.
// DefaultClient has an internal HTTP client (HTTP) which defaults to http.DefaultClient.
type DefaultClient struct {
	// HTTP is the Client's internal http.Client, handling HTTP requests.
	HTTP *http.Client

	// The URL of the API to communicate with.
	URL string

	// Username is the HTTP basic auth username for API calls made by this Client.
	Username string

	// Password is the HTTP basic auth password for API calls made by this Client.
	Password string

	// UserAgent to be provided in API requests. Set to DefaultUserAgent if not
	// specified.
	UserAgent string

	// Debug mode can be used to dump the full request and response to stdout.
	Debug bool

	// AdditionalHeaders are extra headers to add to each HTTP request sent by
	// this Client.
	AdditionalHeaders http.Header
}

// Generates an HTTP request, but does not perform the request.
// The request's Accept header field will be set to:
//
//   Accept: application/json
//
// The Request-Id header will be set to a random UUID. The User-Agent header
// will be set to the Client's UserAgent, or DefaultUserAgent if UserAgent is
// not set.
//
// The type of body determines how to encode the request:
//
//   nil         no body
//   io.Reader   body is sent verbatim
//   else        body is encoded as application/json
//
func (c *DefaultClient) NewRequest(method, path string, body interface{}) (*http.Request, error) {
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
			log.Fatal(err)
		}
		rbody = bytes.NewReader(j)
		ctype = "application/json"
	}
	apiURL := strings.TrimRight(c.URL, "/")
	if apiURL == "" {
		apiURL = DefaultAPIURL
	}
	req, err := http.NewRequest(method, apiURL+path, rbody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Request-Id", uuid.New())
	useragent := c.UserAgent
	if useragent == "" {
		useragent = DefaultUserAgent
	}
	req.Header.Set("User-Agent", useragent)
	if ctype != "" {
		req.Header.Set("Content-Type", ctype)
	}
	req.SetBasicAuth(c.Username, c.Password)
	for k, v := range c.AdditionalHeaders {
		req.Header[k] = v
	}
	return req, nil
}

// Submits an HTTP request, checks its response, and deserializes
// the response into v. The type of v determines how to handle
// the response body:
//
//   nil        body is discarded
//   io.Writer  body is copied directly into v
//   else       body is decoded into v as json
//
func (c *DefaultClient) Do(req *http.Request, v interface{}) error {
	if c.Debug {
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			log.Println(err)
		} else {
			log.Println(string(dump))
		}
	}

	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if c.Debug {
		dump, err := httputil.DumpResponse(res, true)
		if err != nil {
			log.Println(err)
		} else {
			log.Println(string(dump))
		}
	}
	if err = checkResponse(res); err != nil {
		return err
	}
	switch t := v.(type) {
	case nil:
	case io.Writer:
		_, err = io.Copy(t, res.Body)
	default:
		err = json.NewDecoder(res.Body).Decode(v)
	}
	return err
}

func checkResponse(res *http.Response) error {
	if res.StatusCode/100 != 2 { // 200, 201, 202, etc
		// FIXME: Try to handle error json in a generic way.
		return fmt.Errorf("Encountered an error : %s", res.Status)
	}
	return nil
}

type ListRange struct {
	Field      string
	Max        int
	Descending bool
	FirstId    string
	LastId     string
}

func (lr *ListRange) SetHeader(req *http.Request) {
	var hdrval string
	if lr.Field != "" {
		hdrval += lr.Field + " "
	}
	hdrval += lr.FirstId + ".." + lr.LastId
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
`,
	"struct.tmpl": `{{asComment .Definition.Description}}
type {{initialCap .Name}} {{goType .Root .Definition}}
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
