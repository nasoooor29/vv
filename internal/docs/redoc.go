package docs

import (
	"bytes"
	_ "embed"
	"html/template"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

// HTML represents the redoc index.html page
//
//go:embed redoc.html
var HTML string

// JavaScript represents the redoc standalone javascript
//
//go:embed redoc.js
var JavaScript string

func RedocHandler(spec *openapi3.T) echo.HandlerFunc {
	specJson, err := spec.MarshalJSON()
	if err != nil {
		panic(err)
	}

	buf := bytes.NewBuffer(nil)
	tpl, err := template.New("redoc").Parse(HTML)
	if err != nil {
		panic(err)
	}

	if err = tpl.Execute(buf, map[string]any{
		"title": spec.Info.Title,
		"spec":  specJson,
	}); err != nil {
		panic(err)
	}

	html := buf.String()
	return func(c echo.Context) error {
		// if content type is application/json, return the spec as json
		if c.Request().Header.Get(echo.HeaderAccept) == "application/json" {
			c.JSON(http.StatusOK, spec)
		}
		// if html is accepted, return the redoc html
		if strings.Contains(c.Request().Header.Get(echo.HeaderAccept), "text/html") {
			return c.HTML(http.StatusOK, html)
		}
		return nil
	}
}
