package templates

import "text/template"

var templates = map[string]string{"astruct.tmpl": `{{$Root := .Root}} struct {
  {{range $Name, $Definition := .Definition.Properties}}
    {{initialCap $Name}} {{$Root.GoType $Definition}} {{jsonTag $Name $Definition}} {{asComment $Definition.Description}}
  {{end}}
}`,
	"funcs.tmpl": `{{$Name := .Name}}
{{$Root := .Root}}
{{range .Definition.Links}}
  {{asComment .Description}}
  func (c *Client) {{printf "%s-%s" $Name .Title | initialCap}}({{$Root.Parameters . | params}}) ({{$Root.Values $Name . | join}}) {
    {{if eq .Rel "destroy"}}
      return c.Delete(fmt.Sprintf("{{.HRef}}", {{.HRef.Resolve $Root | args}}))
    {{else if eq .Rel "self"}}
      {{$Var := initialLow $Name}}var {{$Var}} {{initialCap $Name}}
      return &{{$Var}}, c.Get(&{{$Var}}, fmt.Sprintf("{{.HRef}}", {{.HRef.Resolve $Root | args}}))
    {{else if eq .Rel "instances"}}
      {{$Var := printf "%s-%s" $Name "List" | initialLow}}
      var {{$Var}} []*{{initialCap $Name}}
      return {{$Var}}, c.{{methodCap .Method}}(&{{$Var}}, fmt.Sprintf("{{.HRef}}", {{.HRef.Resolve $Root | args}}))
    {{else}}
      {{$Var := initialLow $Name}}var {{$Var}} {{initialCap $Name}}
      return &{{$Var}}, c.{{methodCap .Method}}(&{{$Var}}, fmt.Sprintf("{{.HRef}}", {{.HRef.Resolve $Root | args}}), o)
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
	"struct.tmpl": `{{if .Definition.Properties}}
  {{asComment .Definition.Description}}
  type {{initialCap .Name}} {{template "astruct.tmpl" .}}
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
