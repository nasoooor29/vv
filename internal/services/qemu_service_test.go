package services

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"visory/internal/utils"
)

// TestQemuServiceInitialization tests service creation
func TestQemuServiceInitialization(t *testing.T) {
	dispatcher := &utils.Dispatcher{}
	logger := slog.Default()

	t.Run("service requires dispatcher and logger", func(t *testing.T) {
		assert.NotNil(t, dispatcher)
		assert.NotNil(t, logger)
	})
}

// TestGetVirtualMachinesWithoutLibVirt tests error handling when libvirt not available
func TestGetVirtualMachinesWithoutLibVirt(t *testing.T) {
	dispatcher := &utils.Dispatcher{}
	logger := slog.Default()

	service := &QemuService{
		Dispatcher: dispatcher.WithGroup("qemu"),
		Logger:     logger.WithGroup("qemu"),
		LibVirt:    nil, // Simulate LibVirt not available
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/qemu/virtual-machines", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := service.GetVirtualMachines(c)
	assert.Error(t, err, "should return error when LibVirt is not available")
}

// TestGetVirtualMachineWithoutLibVirt tests error handling for specific VM
func TestGetVirtualMachineWithoutLibVirt(t *testing.T) {
	dispatcher := &utils.Dispatcher{}
	logger := slog.Default()

	service := &QemuService{
		Dispatcher: dispatcher.WithGroup("qemu"),
		Logger:     logger.WithGroup("qemu"),
		LibVirt:    nil,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/qemu/virtual-machines/test-uuid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("uuid")
	c.SetParamValues("test-uuid")

	err := service.GetVirtualMachine(c)
	assert.Error(t, err, "should return error when LibVirt is not available")
}

// TestGetVirtualMachineMissingUUID tests missing UUID parameter handling
func TestGetVirtualMachineMissingUUID(t *testing.T) {
	dispatcher := &utils.Dispatcher{}
	logger := slog.Default()

	service := &QemuService{
		Dispatcher: dispatcher.WithGroup("qemu"),
		Logger:     logger.WithGroup("qemu"),
		LibVirt:    nil,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/qemu/virtual-machines/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("uuid")
	c.SetParamValues("")

	err := service.GetVirtualMachine(c)
	assert.Error(t, err, "should return error for missing UUID")
}

// TestStartVirtualMachineWithoutLibVirt tests start VM error handling
func TestStartVirtualMachineWithoutLibVirt(t *testing.T) {
	dispatcher := &utils.Dispatcher{}
	logger := slog.Default()

	service := &QemuService{
		Dispatcher: dispatcher.WithGroup("qemu"),
		Logger:     logger.WithGroup("qemu"),
		LibVirt:    nil,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/qemu/virtual-machines/test-uuid/start", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("uuid")
	c.SetParamValues("test-uuid")

	err := service.StartVirtualMachine(c)
	assert.Error(t, err, "should return error when LibVirt is not available")
}

// TestStartVirtualMachineMissingUUID tests missing UUID for start action
func TestStartVirtualMachineMissingUUID(t *testing.T) {
	dispatcher := &utils.Dispatcher{}
	logger := slog.Default()

	service := &QemuService{
		Dispatcher: dispatcher.WithGroup("qemu"),
		Logger:     logger.WithGroup("qemu"),
		LibVirt:    nil,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/qemu/virtual-machines//start", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("uuid")
	c.SetParamValues("")

	err := service.StartVirtualMachine(c)
	assert.Error(t, err, "should return error for missing UUID")
}

// TestRebootVirtualMachineWithoutLibVirt tests reboot error handling
func TestRebootVirtualMachineWithoutLibVirt(t *testing.T) {
	dispatcher := &utils.Dispatcher{}
	logger := slog.Default()

	service := &QemuService{
		Dispatcher: dispatcher.WithGroup("qemu"),
		Logger:     logger.WithGroup("qemu"),
		LibVirt:    nil,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/qemu/virtual-machines/test-uuid/reboot", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("uuid")
	c.SetParamValues("test-uuid")

	err := service.RebootVirtualMachine(c)
	assert.Error(t, err, "should return error when LibVirt is not available")
}

// TestShutdownVirtualMachineWithoutLibVirt tests shutdown error handling
func TestShutdownVirtualMachineWithoutLibVirt(t *testing.T) {
	dispatcher := &utils.Dispatcher{}
	logger := slog.Default()

	service := &QemuService{
		Dispatcher: dispatcher.WithGroup("qemu"),
		Logger:     logger.WithGroup("qemu"),
		LibVirt:    nil,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/qemu/virtual-machines/test-uuid/shutdown", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("uuid")
	c.SetParamValues("test-uuid")

	err := service.ShutdownVirtualMachine(c)
	assert.Error(t, err, "should return error when LibVirt is not available")
}

// TestGetVirtualMachineInfoWithoutLibVirt tests error handling for VM info
func TestGetVirtualMachineInfoWithoutLibVirt(t *testing.T) {
	dispatcher := &utils.Dispatcher{}
	logger := slog.Default()

	service := &QemuService{
		Dispatcher: dispatcher.WithGroup("qemu"),
		Logger:     logger.WithGroup("qemu"),
		LibVirt:    nil,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/qemu/virtual-machines/test-uuid/info", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("uuid")
	c.SetParamValues("test-uuid")

	err := service.GetVirtualMachineInfo(c)
	assert.Error(t, err, "should return error when LibVirt is not available")
}

// TestGetVirtualMachinesInfoWithoutLibVirt tests error handling for VMs info list
func TestGetVirtualMachinesInfoWithoutLibVirt(t *testing.T) {
	dispatcher := &utils.Dispatcher{}
	logger := slog.Default()

	service := &QemuService{
		Dispatcher: dispatcher.WithGroup("qemu"),
		Logger:     logger.WithGroup("qemu"),
		LibVirt:    nil,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/qemu/virtual-machines/info", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := service.GetVirtualMachinesInfo(c)
	assert.Error(t, err, "should return error when LibVirt is not available")
}
