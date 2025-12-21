package services

import (
	"log/slog"
	"net/http"
	"os"

	"visory/internal/database"
	"visory/internal/utils"

	_ "visory/docs"

	"github.com/labstack/echo/v4"
)

type DocsService struct {
	db         *database.Service
	Dispatcher *utils.Dispatcher
	Logger     *slog.Logger
}

// NewDocsService creates a new DocsService with dependency injection
func NewDocsService(db *database.Service, dispatcher *utils.Dispatcher, logger *slog.Logger) *DocsService {
	return &DocsService{
		db:         db,
		Dispatcher: dispatcher.WithGroup("docs"),
		Logger:     logger.WithGroup("docs"),
	}
}

//	@Summary      swagger ui
//	@Description  serve swagger ui for api documentation
//	@Tags         documentation
//	@Router       /docs/swagger [get]
//
// ServeSwagger serves the Swagger UI
func (s *DocsService) ServeSwagger(c echo.Context) error {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Visory API - Swagger UI</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@3/swagger-ui.css">
    <link rel="icon" type="image/png" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@3/favicon-32x32.png" sizes="32x32" />
    <link rel="icon" type="image/png" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@3/favicon-16x16.png" sizes="16x16" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            padding: 0;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@3/swagger-ui-bundle.js" charset="UTF-8"></script>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@3/swagger-ui-standalone-preset.js" charset="UTF-8"></script>
    <script>
        const ui = SwaggerUIBundle({
            url: "/api/docs/spec",
            dom_id: '#swagger-ui',
            deepLinking: true,
            presets: [
                SwaggerUIBundle.presets.apis,
                SwaggerUIStandalonePreset
            ],
            plugins: [
                SwaggerUIBundle.plugins.DownloadUrl
            ],
            layout: "StandaloneLayout"
        });
        window.onload = function() {
            window.ui = ui;
        }
    </script>
</body>
</html>`
	return c.HTML(http.StatusOK, html)
}

//	@Summary      redoc ui
//	@Description  serve redoc ui for api documentation
//	@Tags         documentation
//	@Router       /docs/redoc [get]
//
// ServeRedoc serves the ReDoc UI
func (s *DocsService) ServeRedoc(c echo.Context) error {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Visory API - ReDoc</title>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">
    <style>
        body {
            margin: 0;
            padding: 0;
        }
    </style>
</head>
<body>
    <redoc spec-url='/api/docs/spec'></redoc>
    <script src="https://cdn.jsdelivr.net/npm/redoc@latest/bundles/redoc.standalone.js"> </script>
</body>
</html>`
	return c.HTML(http.StatusOK, html)
}

//	@Summary      openapi spec
//	@Description  serve openapi/swagger specification in json format
//	@Tags         documentation
//	@Produce      json
//	@Success      200  {object}  map[string]interface{}
//	@Failure      500  {object}  models.HTTPError
//	@Router       /docs/spec [get]
//
// ServeSpec serves the OpenAPI specification
func (s *DocsService) ServeSpec(c echo.Context) error {
	// Try multiple paths in case tests are run from different directories
	possiblePaths := []string{
		"./docs/swagger.json",
		"docs/swagger.json",
		"../../docs/swagger.json",
	}

	var data []byte
	var err error
	var successPath string

	for _, path := range possiblePaths {
		data, err = os.ReadFile(path)
		if err == nil {
			successPath = path
			break
		}
	}

	if err != nil {
		s.Logger.Error("failed to read swagger.json", "error", err, "paths", possiblePaths)
		return s.Dispatcher.NewInternalServerError("Failed to read API specification", err)
	}

	s.Logger.Debug("loaded swagger spec", "path", successPath)
	c.Response().Header().Set("Content-Type", "application/json")
	return c.Blob(http.StatusOK, "application/json", data)
}
