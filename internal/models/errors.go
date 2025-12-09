package models

// this is just to mimic the echo error structure
// eg: {"message":"Failed to list virtual-machines"}
// reprod: echo.NewHTTPError(http.StatusInternalServerError, "Failed to list virtual-machines").SetInternal(fmt.Errorf("database connection error"))
type HTTPError struct {
	Message string `json:"message"`
}
