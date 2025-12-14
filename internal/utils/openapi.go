package utils

import (
	"encoding/json"
	"net/http"
	"strings"

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

func (g *Group) Returns(code int, schema Schema, contentType string, desc ...string) *Group {
	if contentType == "" {
		contentType = "application/json"
	}
	description := ""
	if len(desc) > 0 {
		description = desc[0]
	}
	g.responses[code] = Response{
		Schema:      schema,
		ContentType: contentType,
		Description: description,
	}
	return g
}

func (g *Group) Description(desc string) *Group {
	return g
}

/* =========================
   INPUT
========================= */

type Input struct {
	Schema      Schema
	ContentType string
	Required    bool
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
	input       *Input
}

type Response struct {
	Schema      Schema
	ContentType string
	Description string
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

func (r *Route) Returns(code int, schema Schema, contentType string, desc ...string) *Route {
	if contentType == "" {
		contentType = "application/json"
	}
	description := ""
	if len(desc) > 0 {
		description = desc[0]
	}
	r.Responses[code] = Response{
		Schema:      schema,
		ContentType: contentType,
		Description: description,
	}
	return r
}

func (r *Route) Description(desc string) *Route {
	r.description = desc
	return r
}

func (r *Route) Input(schema Schema, contentType string, required bool) *Route {
	if contentType == "" {
		contentType = "application/json"
	}
	r.input = &Input{
		Schema:      schema,
		ContentType: contentType,
		Required:    required,
	}
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

	// Convert to map to clean it up
	schemaJSON, err := json.Marshal(jsonSch)
	if err != nil {
		return openapi3.NewSchemaRef("", &openapi3.Schema{})
	}

	// Parse into map to remove problematic fields
	var schemaMap map[string]interface{}
	err = json.Unmarshal(schemaJSON, &schemaMap)
	if err != nil {
		return openapi3.NewSchemaRef("", &openapi3.Schema{})
	}

	// Extract and collect all definitions first
	defs := make(map[string]interface{})
	if defsVal, hasDefs := schemaMap["$defs"]; hasDefs {
		if defsMap, ok := defsVal.(map[string]interface{}); ok {
			for k, v := range defsMap {
				defs[k] = v
			}
		}
	}

	// Recursively inline all $refs
	inlineRefs(schemaMap, defs)

	// Remove $schema and $defs
	delete(schemaMap, "$schema")
	delete(schemaMap, "$defs")

	// Re-marshal and unmarshal into openapi3.Schema
	cleanJSON, _ := json.Marshal(schemaMap)
	var schema openapi3.Schema
	json.Unmarshal(cleanJSON, &schema)

	return openapi3.NewSchemaRef("", &schema)
}

// inlineRefs recursively replaces all $ref with their actual definitions
func inlineRefs(obj interface{}, defs map[string]interface{}) {
	switch v := obj.(type) {
	case map[string]interface{}:
		// If this object has a $ref, try to inline it
		if ref, hasRef := v["$ref"]; hasRef {
			if refStr, ok := ref.(string); ok {
				// Extract definition name from #/$defs/Name
				parts := strings.Split(refStr, "/")
				if len(parts) > 0 {
					defName := parts[len(parts)-1]
					if def, hasDef := defs[defName]; hasDef {
						// Copy definition fields into this object
						if defMap, ok := def.(map[string]interface{}); ok {
							// First recursively process the definition itself
							inlineRefs(defMap, defs)
							// Then copy all fields except metadata
							for k, val := range defMap {
								if k != "$schema" && k != "$defs" && k != "$ref" {
									v[k] = val
								}
							}
						}
						// Remove the $ref
						delete(v, "$ref")
					}
				}
			}
		}

		// Recursively process all nested objects
		for _, val := range v {
			inlineRefs(val, defs)
		}

	case []interface{}:
		for _, item := range v {
			inlineRefs(item, defs)
		}
	}
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

func getStatusDescription(code int) string {
	descriptions := map[int]string{
		200: "Success",
		201: "Created",
		204: "No Content",
		400: "Bad Request",
		401: "Unauthorized",
		403: "Forbidden",
		404: "Not Found",
		409: "Conflict",
		500: "Internal Server Error",
		502: "Bad Gateway",
		503: "Service Unavailable",
	}
	if desc, ok := descriptions[code]; ok {
		return desc
	}
	return "Response"
}

func (a *API) OpenAPI() *openapi3.T {
	doc := &openapi3.T{
		OpenAPI: "3.0.0",
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

		// Add request body if provided
		if r.input != nil {
			contentType := r.input.ContentType
			if contentType == "" {
				contentType = "application/json"
			}
			content := openapi3.NewContent()
			schemaRef := r.input.Schema.OpenAPI()
			if schemaRef.Value != nil {
				schemaRef.Value.Extensions = nil
			}
			content[contentType] = &openapi3.MediaType{
				Schema: schemaRef,
			}
			op.RequestBody = &openapi3.RequestBodyRef{
				Value: &openapi3.RequestBody{
					Content:  content,
					Required: r.input.Required,
				},
			}
		}

		for code, resp := range r.Responses {
			contentType := resp.ContentType
			if contentType == "" {
				contentType = "application/json"
			}

			var content openapi3.Content
			if resp.Schema != nil {
				content = openapi3.NewContent()
				schemaRef := resp.Schema.OpenAPI()
				// Clean up the schema - remove $schema, $defs that cause issues
				if schemaRef.Value != nil {
					schemaRef.Value.Extensions = nil
				}
				content[contentType] = &openapi3.MediaType{
					Schema: schemaRef,
				}
			}

			// Use custom description if provided, otherwise use status text
			statusDesc := resp.Description
			if statusDesc == "" {
				statusDesc = getStatusDescription(code)
			}

			op.AddResponse(code, &openapi3.Response{
				Description: &statusDesc,
				Content:     content,
			})
		}

		doc.AddOperation(r.Path, r.Method, op)
	}

	return doc
}
