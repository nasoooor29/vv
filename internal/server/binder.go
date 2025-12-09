package server

import (
	"encoding/hex"
	"net/http"
	"reflect"
	"strings"

	"github.com/digitalocean/go-libvirt"
	"github.com/labstack/echo/v4"
)

type CustomBinder struct{}

func (pb *CustomBinder) Bind(i any, c echo.Context) error {
	db := new(echo.DefaultBinder)
	if err := db.Bind(i, c); err != nil {
		if err.Error() != "code=400, message=unknown type, internal=unknown type" {
			return err
		}
	}
	sVal := reflect.ValueOf(i).Elem()
	typ := sVal.Type()
	for i := range sVal.NumField() {
		field := typ.Field(i)
		val := sVal.Field(i)
		switch field.Type.String() {
		case "libvirt.UUID":
			if val.CanSet() {
				uuidStr := strings.ReplaceAll(strings.TrimSpace(c.Param(field.Tag.Get("param"))), "-", "")
				if uuidStr == "" {
					return echo.NewHTTPError(http.StatusBadRequest, "uuid is required")
				}
				uuid, err := hex.DecodeString(uuidStr)
				if err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, "failed to decode uuid").SetInternal(err)
				}
				val.Set(reflect.ValueOf(libvirt.UUID([]byte(uuid))))
			}
		}
	}

	return nil
}
