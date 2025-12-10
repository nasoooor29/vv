package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"reflect"

	"github.com/invopop/jsonschema"
	"github.com/labstack/echo/v4"
)

/* =======================
   RESPONSE CORE
======================= */

type SerializerType int

const (
	JSON SerializerType = iota
	XML
	HTML
	BLOB
)

type Response[T any] struct {
	Status int
	Body   T
	Type   SerializerType
}

func JSONResponse[T any](status int, body T) Response[T] {
	return Response[T]{Status: status, Body: body, Type: JSON}
}

/* =======================
   TYPED ECHO / GROUP
======================= */

type TypedEcho struct {
	*echo.Echo
	Schemas map[string]reflect.Type
}

type TypedGroup struct {
	Group   *echo.Group
	Prefix  string
	Schemas map[string]reflect.Type
}

func NewTypedEcho(e *echo.Echo) *TypedEcho {
	return &TypedEcho{
		Echo:    e,
		Schemas: make(map[string]reflect.Type),
	}
}

func (t *TypedEcho) Group(prefix string, m ...echo.MiddlewareFunc) *TypedGroup {
	return &TypedGroup{
		Group:   t.Echo.Group(prefix, m...),
		Prefix:  prefix,
		Schemas: t.Schemas,
	}
}

/* =======================
   GENERIC ROUTES
======================= */

func GET[T any](
	root *TypedEcho,
	pathname string,
	handler func(echo.Context) (Response[T], error),
) {
	registerGET[T](root.Echo, root.Schemas, pathname, handler)
}

func GroupGET[T any](
	g *TypedGroup,
	pathname string,
	handler func(echo.Context) (Response[T], error),
) {
	full := path.Join(g.Prefix, pathname)
	registerGET[T](g.Group, g.Schemas, full, handler)
}

func registerGET[T any](
	r interface {
		GET(string, echo.HandlerFunc, ...echo.MiddlewareFunc) *echo.Route
	},
	schemas map[string]reflect.Type,
	fullPath string,
	handler func(echo.Context) (Response[T], error),
) {
	var zero *T
	typ := reflect.TypeOf(zero).Elem()
	schemas[fullPath] = typ

	fmt.Printf("registered GET %s → %v\n", fullPath, typ)

	r.GET(fullPath, func(c echo.Context) error {
		resp, err := handler(c)
		if err != nil {
			return err
		}
		if resp.Type == JSON {
			return c.JSON(resp.Status, resp.Body)
		}
		return echo.ErrNotImplemented
	})
}

/* =======================
   SCHEMA OUTPUT MODEL
======================= */

type RouteSchema struct {
	Kind       string             `json:"kind"` // "json-schema" | "primitive"
	JSONSchema *jsonschema.Schema `json:"jsonSchema,omitempty"`
	Primitive  string             `json:"primitive,omitempty"`
}

/* =======================
   SCHEMA GENERATION
======================= */

func GenerateRouteSchemas(
	schemas map[string]reflect.Type,
) map[string]RouteSchema {
	r := jsonschema.Reflector{
		DoNotReference: true,
		ExpandedStruct: true,
	}

	out := make(map[string]RouteSchema)

	for route, typ := range schemas {
		if isPrimitiveType(typ) {
			out[route] = RouteSchema{
				Kind:      "primitive",
				Primitive: primitiveName(typ),
			}
			continue
		}

		out[route] = RouteSchema{
			Kind:       "json-schema",
			JSONSchema: r.ReflectFromType(typ),
		}
	}

	return out
}

/* =======================
   TYPE HELPERS
======================= */

func isPrimitiveType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.String,
		reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func primitiveName(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "integer"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "unsigned-integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	default:
		return "unknown"
	}
}

/* =======================
   DOMAIN
======================= */

type User struct {
	ID       int64    `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Role     string   `json:"role"`
	Age      *int     `json:"age"`
	Session  *Session `json:"session,omitempty"`
}
type Session struct {
	Token  string `json:"token"`
	UserID int64  `json:"user_id"`
}

func MeHandler(c echo.Context) (Response[User], error) {
	return JSONResponse(http.StatusOK, User{
		ID:       1,
		Username: "johndoe",
		Email:    "smth@smth.smth",
		Role:     "admin",
		Age:      nil,
	}), nil
}

/* =======================
   MAIN
======================= */

func main() {
	e := NewTypedEcho(echo.New())

	auth := e.Group("/auth")

	GroupGET(auth, "/me", MeHandler)

	GET(e, "/health", func(c echo.Context) (Response[string], error) {
		return JSONResponse(http.StatusOK, "ok"), nil
	})

	fmt.Println("\n=== ROUTE SCHEMAS ===")

	rs := GenerateRouteSchemas(e.Schemas)
	for route, schema := range rs {
		b, _ := json.MarshalIndent(schema, "", "  ")
		fmt.Printf("\n%s\n%s\n", route, b)
	}
}
