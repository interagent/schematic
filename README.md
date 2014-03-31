# Schematic

Generate Go client code for HTTP APIs described by [JSON Schemas](http://json-schema.org/).

## Installation

Download and install:

```console
$ go get github.com/heroku/schematic
```

**Warning**: schematic requires Go >= 1.2.

## Usage

Run it against your schema:

```console
$ schematic api.json > heroku/heroku.go 
```

This will generate a Go package named after your schema:

```go
package heroku
...
```

You then would be able to use the package as follow:

```go
h := heroku.NewService(nil)
addons, err := h.AddonList("schematic", nil)
if err != nil {
  ...
}
for _, addon := range addons {
  fmt.Println(addon.Name)
}
```

You can also provide a custom client:

```go
h := heroku.NewService(&HerokuClient{})
```

## Development

Schematic bundles templated Go code into a Go source file via the
[templates package](https://github.com/cyberdelia/templates). To rebuild
the Go source file after changing .tmpl files:

```console
$ templates -source schema/templates/ > schema/templates/templates.go
```
