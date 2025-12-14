package utils

import (
	"encoding/json"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/invopop/jsonschema"
	"github.com/labstack/echo/v4"
)

/* =========================
   API / GROUP
========================= */

type API struct {
	routes []*Route
}

func New() *API {
	return &API{}
}

func (a *API) Group(prefix string, mw ...echo.MiddlewareFunc) *Group {
	return &Group{
		api:        a,
		prefix:     prefix,
		middleware: mw,
		responses:  make(map[int]Response),
	}
}

type Group struct {
	api        *API
	prefix     string
	middleware []echo.MiddlewareFunc
	parent     *Group
	responses  map[int]Response
}

func (g *Group) Group(prefix string, mw ...echo.MiddlewareFunc) *Group {
	return &Group{
		api:        g.api,
		parent:     g,
		prefix:     prefix,
		middleware: mw,
		responses:  make(map[int]Response),
	}
}

func (g *Group) fullPrefix() string {
	if g.parent == nil {
		return g.prefix
	}
	return g.parent.fullPrefix() + g.prefix
}

func (g *Group) fullMiddleware() []echo.MiddlewareFunc {
	if g.parent == nil {
		return g.middleware
	}
	return append(g.parent.fullMiddleware(), g.middleware...)
}

func (g *Group) GET(path string, h echo.HandlerFunc, mw ...echo.MiddlewareFunc) *Route {
	r := g.api.add(
		http.MethodGet,
		g.fullPrefix()+path,
		h,
		append(g.fullMiddleware(), mw...)...,
	)
	for code, resp := range g.responses {
		r.Responses[code] = resp
	}
	return r
}

func (g *Group) POST(path string, h echo.HandlerFunc, mw ...echo.MiddlewareFunc) *Route {
	r := g.api.add(
		http.MethodPost,
		g.fullPrefix()+path,
		h,
		append(g.fullMiddleware(), mw...)...,
	)
	for code, resp := range g.responses {
		r.Responses[code] = resp
	}
	return r
}

func (g *Group) Returns(code int, schema Schema, contentType string) *Group {
	if contentType == "" {
		contentType = "application/json"
	}
	g.responses[code] = Response{
		Schema:      schema,
		ContentType: contentType,
	}
	return g
}

func (g *Group) Description(desc string) *Group {
	return g
}

/* =========================
   ROUTE
========================= */

type Route struct {
	Method      string
	Path        string
	Handler     echo.HandlerFunc
	Middleware  []echo.MiddlewareFunc
	Params      []Param
	Responses   map[int]Response
	description string
}

type Response struct {
	Schema      Schema
	ContentType string
}

func (a *API) add(
	method string,
	path string,
	h echo.HandlerFunc,
	mw ...echo.MiddlewareFunc,
) *Route {
	r := &Route{
		Method:     method,
		Path:       path,
		Handler:    h,
		Middleware: mw,
		Responses:  make(map[int]Response),
	}
	a.routes = append(a.routes, r)
	return r
}

func (r *Route) Returns(code int, schema Schema, contentType string) *Route {
	if contentType == "" {
		contentType = "application/json"
	}
	r.Responses[code] = Response{
		Schema:      schema,
		ContentType: contentType,
	}
	return r
}

func (r *Route) Description(desc string) *Route {
	r.description = desc
	return r
}

/* =========================
   PARAMS
========================= */

type Param interface {
	Name() string
	In() string
	Schema() Schema
}

type pathParam struct {
	name   string
	schema Schema
}

func Path(name string, s Schema) Param {
	return &pathParam{name, s}
}

func (p *pathParam) Name() string   { return p.name }
func (p *pathParam) In() string     { return "path" }
func (p *pathParam) Schema() Schema { return p.schema }

/* =========================
   SCHEMA
========================= */

type Schema interface {
	OpenAPI() *openapi3.SchemaRef
}

type structSchema struct {
	typ any
}

func JSON(v any) Schema {
	return &structSchema{v}
}

func (s *structSchema) OpenAPI() *openapi3.SchemaRef {
	if s.typ == nil {
		return openapi3.NewSchemaRef("", &openapi3.Schema{})
	}

	// Use jsonschema to generate schema from the struct
	jsonSch := jsonschema.Reflect(s.typ)

	// Convert jsonschema to openapi3 schema
	schemaJSON, err := json.MarshalIndent(jsonSch, "", "  ")
	if err != nil {
		return openapi3.NewSchemaRef("", &openapi3.Schema{})
	}

	// Parse the JSON back into an openapi3 schema
	var schema openapi3.Schema
	err = json.Unmarshal(schemaJSON, &schema)
	if err != nil {
		return openapi3.NewSchemaRef("", &openapi3.Schema{})
	}

	return openapi3.NewSchemaRef("", &schema)
}

/* =========================
   ECHO MOUNT
========================= */

func (a *API) Mount(e *echo.Echo) {
	for _, r := range a.routes {
		e.Add(
			r.Method,
			r.Path,
			r.Handler,
			r.Middleware...,
		)
	}
}

/* =========================
   OPENAPI
========================= */

func (a *API) OpenAPI() *openapi3.T {
	doc := &openapi3.T{
		OpenAPI: "3.1.0",
		Info: &openapi3.Info{
			Title:   "API",
			Version: "1.0.0",
		},
		Paths: &openapi3.Paths{},
	}

	for _, r := range a.routes {
		op := openapi3.NewOperation()
		op.Description = r.description

		for _, p := range r.Params {
			op.Parameters = append(op.Parameters, &openapi3.ParameterRef{
				Value: &openapi3.Parameter{
					Name:     p.Name(),
					In:       p.In(),
					Required: true,
					Schema:   p.Schema().OpenAPI(),
				},
			})
		}

		for code, resp := range r.Responses {
			contentType := resp.ContentType
			if contentType == "" {
				contentType = "application/json"
			}

			var content openapi3.Content
			if resp.Schema != nil {
				content = openapi3.NewContent()
				content[contentType] = &openapi3.MediaType{
					Schema: resp.Schema.OpenAPI(),
				}
			}

			op.AddResponse(code, &openapi3.Response{
				Content: content,
			})
		}

		doc.AddOperation(r.Path, r.Method, op)
	}

	return doc
}
