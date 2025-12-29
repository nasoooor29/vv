package services

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/evangwt/go-vncproxy"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

type VNCProxy struct {
	logger *slog.Logger
}

// NewVNCProxy creates a new VNC proxy service
func NewVNCProxy(logger *slog.Logger) *VNCProxy {
	return &VNCProxy{
		logger: logger.WithGroup("vnc-proxy"),
	}
}

func (p *VNCProxy) ConnectVNC(c echo.Context, vncIP string, vncPort int) error {
	vncAddr := net.JoinHostPort(vncIP, fmt.Sprintf("%d", vncPort))
	p.logger.Info("Connecting to VNC server", "address", vncAddr)

	vncProxy := vncproxy.New(&vncproxy.Config{
		LogLevel: vncproxy.DebugLevel,
		TokenHandler: func(r *http.Request) (string, error) {
			return vncAddr, nil
		},
	})

	websocket.Handler(vncProxy.ServeWS).ServeHTTP(
		c.Response().Writer,
		c.Request(),
	)

	return nil
}
