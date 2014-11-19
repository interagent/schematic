# Schematic

Generate Go client code for HTTP APIs described by [JSON Hyper-Schemas](http://json-schema.org/latest/json-schema-hypermedia.html).

## Installation

Download and install:

```console
$ go get -u github.com/interagent/schematic/cmd/schematic
```

**Warning**: schematic requires Go >= 1.2.

## Client Generation

Run it against your schema:

```console
$ schematic platform-api.json > heroku/heroku.go
```

This will generate a Go package named after your schema:

```go
package heroku
...
```

## Client Usage

You then would be able to use the package as follow:

```go
package main

import "./heroku"

func main() {
  h := heroku.NewService(nil)
  addons, err := h.AddonServiceList(nil)
  if err != nil {
    panic(err)
  }
  for _, addon := range addons {
    fmt.Println(addon.Name)
  }
}
```

A Service takes an optional [`http.Client`](http://golang.org/pkg/net/http/#Client)
as argument. This Client will need to handle any HTTP-related details
unique to the target API service such as request authorization, request
header setup, error handling, response debugging, etc.

As an example, if your service use [OAuth2](http://code.google.com/p/goauth2/)
for authentication:

```go
var config = &oauth.Config{
  ...
}
transport := &oauth.Transport{
  Token:     token,
  Config:    config,
  Transport: http.DefaultTransport,
}
httpClient := transport.Client()
s := api.NewService(httpClient)
```

For an example of a service using HTTP Basic Auth, see the generated
[heroku-go client](https://github.com/cyberdelia/heroku-go/blob/master/v3/transport.go).

In general, creating a custom [`http.Transport`](http://golang.org/pkg/net/http/#Transport)
and creating a client from that is a common way to get an appropriate
`http.Client` for your service.

## Client Method Types

Methods on generated clients will follow one of a few common patterns
corresponding to reading a single resource, reading a list of resources,
and creating or updating a resource.

Reading a single resource looks like this, for example:

```go
app, err := h.AppInfo("my-app")
if err != nil {
   panic(err)
}
fmt.Println(app.Name)
fmt.Printf("%+v\n", app)
```

Where the struct `app` is of a type e.g. `heroku.App` defined in the
generated client. It will include generated struct members for the
various fields included in the API response.

Methods to read a list of resources look similar, for example:

```go
apps, err := h.AppList(nil)
if err != nil {
    panic(err)
}
for _, app := range apps {
    fmt.Println(app.Name)
}
```

The first return value of these methods is a slice of domain structs
like `heroku.App` described above.

List methods take a `*ListRange` argument that can be used to specify
ordering and pagination range options on the underlying list call.

Methods to create or update look like this, for example:

```go
newName := "my-app"
updateOpts := heroku.AppUpdateOpts{
    Name: &newName,
}
app, err := h.AppUpdate("my-renamed-app", updateOpts)
if err != nil {
    panic(err)
}
fmt.Println(app.Name)
```

Note the availability of generated types of with suffixes `UpdateOpts`
and `CreateOpts` - these make it easy to generate arguments for update
and create calls. These types have pointer members instead of value
members so that you can omit some of them without defaulting them to
the a zero-value.

See the generated godocs for your package for details on the generated
methods and types.

## Development

Schematic bundles templated Go code into a Go source file via the
[templates package](https://github.com/cyberdelia/templates). To rebuild
the Go source file after changing .tmpl files:

```console
$ templates -source schema/templates/ > schema/templates/templates.go
```
