package server

import (
	"log/slog"
	"net/http"

	"visory/internal/database"
	"visory/internal/database/user"
	"visory/internal/models"

	"github.com/labstack/echo/v4"
)

type SerilizerType int

const (
	JSON SerilizerType = iota
	XML
	HTML
	BLOB
)

type Serilizable[T any] interface {
	Serilize(int, any) error
	GetWithStatus() (int, T)
	GetSerilizerType() SerilizerType
}
type JsonSerlizer[T any] struct {
	status int
	jj     T
	e      echo.Context
	_type  SerilizerType
}

func (t *JsonSerlizer[T]) GetWithStatus() (int, T) {
	return t.status, t.jj
}

func (t *JsonSerlizer[T]) GetSerilizerType() SerilizerType {
	return t._type
}

func NewJsonSerilizer[T any](e echo.Context, status int, jj T) *JsonSerlizer[T] {
	return &JsonSerlizer[T]{
		status: status,
		jj:     jj,
		e:      e,
		_type:  JSON,
	}
}

func (t *JsonSerlizer[T]) Serilize(status int, v any) error {
	return t.e.JSON(t.status, t.jj)
}

// this can go inside echo.Echo ?
func ToEchoHandlerFunc[T any](f func(echo.Context) (Serilizable[T], error)) echo.HandlerFunc {
	return func(c echo.Context) error {
		serilizer, err := f(c)
		if err != nil {
			return err
		}
		s, v := serilizer.GetWithStatus()
		return serilizer.Serilize(s, v)
	}
}

func (s *Server) Me2(c echo.Context) (Serilizable[user.GetUserAndSessionByTokenRow], error) {
	cookie, err := c.Cookie(models.COOKIE_NAME)
	if err != nil {
		slog.Error("error happened", "err", err)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Failed to get user by session token").SetInternal(err)
	}
	userWithSession, err := s.db.User.GetUserAndSessionByToken(c.Request().Context(), cookie.Value)
	if err != nil {
		slog.Error("error happened", "err", err)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Failed to get user by session token").SetInternal(err)
	}
	// return c.JSON(http.StatusOK, userWithSession), nil
	return NewJsonSerilizer(c, http.StatusOK, userWithSession), nil
}

func Dummy_() {
	e := echo.New()
	s := &Server{
		db:   database.New(),
		port: 3338,
	}
	e.GET("/me2", ToEchoHandlerFunc(s.Me2))
}
