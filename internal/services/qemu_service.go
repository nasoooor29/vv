package services

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"visory/internal/models"
	"visory/internal/utils"

	"github.com/digitalocean/go-libvirt"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
)

type QemuService struct {
	Dispatcher *utils.Dispatcher
	Logger     *slog.Logger
	LibVirt    *libvirt.Libvirt
}

// NewQemuService creates a new QemuService with dependency injection
func NewQemuService(dispatcher *utils.Dispatcher, logger *slog.Logger) *QemuService {
	// Connect to libvirt QEMU system
	uri, _ := url.Parse(string(libvirt.QEMUSystem))
	l, err := libvirt.ConnectToURI(uri)
	if err != nil {
		logger.Error("Failed to connect to libvirt", "error", err)
		return nil
	}

	return &QemuService{
		Dispatcher: dispatcher.WithGroup("qemu"),
		Logger:     logger.WithGroup("qemu"),
		LibVirt:    l,
	}
}

//	@Summary      List virtual machines
//	@Description  Get a list of all QEMU virtual machines
//	@Tags         qemu
//	@Produce      json
//	@Success      200  {array}   models.VirtualMachine
//	@Failure      401  {object}  models.HTTPError
//	@Failure      403  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /qemu/virtual-machines [get]
//
// GetVirtualMachines returns list of virtual machines
func (s *QemuService) GetVirtualMachines(c echo.Context) error {
	if s.LibVirt == nil {
		return s.Dispatcher.NewInternalServerError("LibVirt connection not available", nil)
	}

	flags := libvirt.ConnectListDomainsActive | libvirt.ConnectListDomainsInactive
	domains, _, err := s.LibVirt.ConnectListAllDomains(1, flags)
	if err != nil {
		s.Logger.Error("Failed to list domains", "error", err)
		return s.Dispatcher.NewInternalServerError("Failed to list virtual machines", err)
	}

	vms := make([]models.VirtualMachine, 0, len(domains))
	for _, domain := range domains {
		domainUUID, err := uuid.FromBytes(domain.UUID[:])
		if err != nil {
			s.Logger.Warn("Failed to parse domain UUID", "error", err)
			continue
		}
		vms = append(vms, models.VirtualMachine{
			ID:   domain.ID,
			Name: domain.Name,
			UUID: domainUUID.String(),
		})
	}

	return c.JSON(http.StatusOK, vms)
}

//	@Summary      Get virtual machine info
//	@Description  Get detailed information about all virtual machines
//	@Tags         qemu
//	@Produce      json
//	@Success      200  {array}   models.VirtualMachineWithInfo
//	@Failure      401  {object}  models.HTTPError
//	@Failure      403  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /qemu/virtual-machines/info [get]
//
// GetVirtualMachinesInfo returns list of VMs with info
func (s *QemuService) GetVirtualMachinesInfo(c echo.Context) error {
	if s.LibVirt == nil {
		return s.Dispatcher.NewInternalServerError("LibVirt connection not available", nil)
	}

	flags := libvirt.ConnectListDomainsActive | libvirt.ConnectListDomainsInactive
	domains, _, err := s.LibVirt.ConnectListAllDomains(1, flags)
	if err != nil {
		s.Logger.Error("Failed to list domains", "error", err)
		return s.Dispatcher.NewInternalServerError("Failed to list virtual machines", err)
	}

	vms := make([]models.VirtualMachineWithInfo, 0, len(domains))
	for _, domain := range domains {
		domainUUID, err := uuid.FromBytes(domain.UUID[:])
		if err != nil {
			s.Logger.Warn("Failed to parse domain UUID", "error", err)
			continue
		}

		// Get domain info
		rState, rMaxMem, rMemory, rNrVirtCPU, rCPUTime, err := s.LibVirt.DomainGetInfo(domain)
		if err != nil {
			s.Logger.Warn("Failed to get domain info", "domain", domain.Name, "error", err)
			continue
		}

		vms = append(vms, models.VirtualMachineWithInfo{
			ID:   domain.ID,
			Name: domain.Name,
			UUID: domainUUID.String(),
			VirtualMachineInfo: models.VirtualMachineInfo{
				State:     rState,
				MaxMemKB:  rMaxMem,
				MemoryKB:  rMemory,
				VCPUs:     rNrVirtCPU,
				CPUTimeNs: rCPUTime,
			},
		})
	}

	return c.JSON(http.StatusOK, vms)
}

//	@Summary      Get specific virtual machine
//	@Description  Get basic information about a specific virtual machine
//	@Tags         qemu
//	@Param        uuid  path  string  true  "Virtual Machine UUID"
//	@Produce      json
//	@Success      200  {object}  models.VirtualMachine
//	@Failure      401  {object}  models.HTTPError
//	@Failure      403  {object}  models.HTTPError
//	@Failure      404  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /qemu/virtual-machines/{uuid} [get]
//
// GetVirtualMachine returns a specific virtual machine
func (s *QemuService) GetVirtualMachine(c echo.Context) error {
	if s.LibVirt == nil {
		return s.Dispatcher.NewInternalServerError("LibVirt connection not available", nil)
	}

	vmUUID := c.Param("uuid")
	if vmUUID == "" {
		return s.Dispatcher.NewBadRequest("Virtual machine UUID is required", nil)
	}

	flags := libvirt.ConnectListDomainsActive | libvirt.ConnectListDomainsInactive
	domains, _, err := s.LibVirt.ConnectListAllDomains(1, flags)
	if err != nil {
		s.Logger.Error("Failed to list domains", "error", err)
		return s.Dispatcher.NewInternalServerError("Failed to list virtual machines", err)
	}

	for _, domain := range domains {
		domainUUID, err := uuid.FromBytes(domain.UUID[:])
		if err != nil {
			continue
		}
		if domainUUID.String() == vmUUID {
			return c.JSON(http.StatusOK, models.VirtualMachine{
				ID:   domain.ID,
				Name: domain.Name,
				UUID: domainUUID.String(),
			})
		}
	}

	return s.Dispatcher.NewNotFound("Virtual machine not found", nil)
}

//	@Summary      Get virtual machine detailed info
//	@Description  Get detailed information about a specific virtual machine
//	@Tags         qemu
//	@Param        uuid  path  string  true  "Virtual Machine UUID"
//	@Produce      json
//	@Success      200  {object}  models.VirtualMachineWithInfo
//	@Failure      401  {object}  models.HTTPError
//	@Failure      403  {object}  models.HTTPError
//	@Failure      404  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /qemu/virtual-machines/{uuid}/info [get]
//
// GetVirtualMachineInfo returns detailed info for a specific VM
func (s *QemuService) GetVirtualMachineInfo(c echo.Context) error {
	if s.LibVirt == nil {
		return s.Dispatcher.NewInternalServerError("LibVirt connection not available", nil)
	}

	vmUUID := c.Param("uuid")
	if vmUUID == "" {
		return s.Dispatcher.NewBadRequest("Virtual machine UUID is required", nil)
	}

	flags := libvirt.ConnectListDomainsActive | libvirt.ConnectListDomainsInactive
	domains, _, err := s.LibVirt.ConnectListAllDomains(1, flags)
	if err != nil {
		s.Logger.Error("Failed to list domains", "error", err)
		return s.Dispatcher.NewInternalServerError("Failed to list virtual machines", err)
	}

	for _, domain := range domains {
		domainUUID, err := uuid.FromBytes(domain.UUID[:])
		if err != nil {
			continue
		}
		if domainUUID.String() == vmUUID {
			// Get domain info
			rState, rMaxMem, rMemory, rNrVirtCPU, rCPUTime, err := s.LibVirt.DomainGetInfo(domain)
			if err != nil {
				s.Logger.Error("Failed to get domain info", "domain", domain.Name, "error", err)
				return s.Dispatcher.NewInternalServerError("Failed to get virtual machine info", err)
			}

			return c.JSON(http.StatusOK, models.VirtualMachineWithInfo{
				ID:   domain.ID,
				Name: domain.Name,
				UUID: domainUUID.String(),
				VirtualMachineInfo: models.VirtualMachineInfo{
					State:     rState,
					MaxMemKB:  rMaxMem,
					MemoryKB:  rMemory,
					VCPUs:     rNrVirtCPU,
					CPUTimeNs: rCPUTime,
				},
			})
		}
	}

	return s.Dispatcher.NewNotFound("Virtual machine not found", nil)
}

//	@Summary      Start virtual machine
//	@Description  Start a stopped virtual machine
//	@Tags         qemu
//	@Param        uuid  path  string  true  "Virtual Machine UUID"
//	@Produce      json
//	@Success      200  {object}  models.VMActionResponse
//	@Failure      401  {object}  models.HTTPError
//	@Failure      403  {object}  models.HTTPError
//	@Failure      404  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /qemu/virtual-machines/{uuid}/start [post]
//
// StartVirtualMachine starts a virtual machine
func (s *QemuService) StartVirtualMachine(c echo.Context) error {
	if s.LibVirt == nil {
		return s.Dispatcher.NewInternalServerError("LibVirt connection not available", nil)
	}

	vmUUID := c.Param("uuid")
	if vmUUID == "" {
		return s.Dispatcher.NewBadRequest("Virtual machine UUID is required", nil)
	}

	domain, err := s.getDomainByUUID(vmUUID)
	if err != nil {
		s.Logger.Error("Failed to find domain", "uuid", vmUUID, "error", err)
		return s.Dispatcher.NewNotFound("Virtual machine not found", err)
	}

	if err := s.LibVirt.DomainCreate(domain); err != nil {
		s.Logger.Error("Failed to start domain", "name", domain.Name, "error", err)
		return s.Dispatcher.NewInternalServerError("Failed to start virtual machine", err)
	}

	return c.JSON(http.StatusOK, models.VMActionResponse{
		Success: true,
		Message: fmt.Sprintf("Virtual machine '%s' started successfully", domain.Name),
	})
}

//	@Summary      Reboot virtual machine
//	@Description  Reboot a running virtual machine
//	@Tags         qemu
//	@Param        uuid  path  string  true  "Virtual Machine UUID"
//	@Produce      json
//	@Success      200  {object}  models.VMActionResponse
//	@Failure      401  {object}  models.HTTPError
//	@Failure      403  {object}  models.HTTPError
//	@Failure      404  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /qemu/virtual-machines/{uuid}/reboot [post]
//
// RebootVirtualMachine reboots a virtual machine
func (s *QemuService) RebootVirtualMachine(c echo.Context) error {
	if s.LibVirt == nil {
		return s.Dispatcher.NewInternalServerError("LibVirt connection not available", nil)
	}

	vmUUID := c.Param("uuid")
	if vmUUID == "" {
		return s.Dispatcher.NewBadRequest("Virtual machine UUID is required", nil)
	}

	domain, err := s.getDomainByUUID(vmUUID)
	if err != nil {
		s.Logger.Error("Failed to find domain", "uuid", vmUUID, "error", err)
		return s.Dispatcher.NewNotFound("Virtual machine not found", err)
	}

	if err := s.LibVirt.DomainReboot(domain, libvirt.DomainRebootDefault); err != nil {
		s.Logger.Error("Failed to reboot domain", "name", domain.Name, "error", err)
		return s.Dispatcher.NewInternalServerError("Failed to reboot virtual machine", err)
	}

	return c.JSON(http.StatusOK, models.VMActionResponse{
		Success: true,
		Message: fmt.Sprintf("Virtual machine '%s' rebooted successfully", domain.Name),
	})
}

//	@Summary      Shutdown virtual machine
//	@Description  Gracefully shutdown a running virtual machine
//	@Tags         qemu
//	@Param        uuid  path  string  true  "Virtual Machine UUID"
//	@Produce      json
//	@Success      200  {object}  models.VMActionResponse
//	@Failure      401  {object}  models.HTTPError
//	@Failure      403  {object}  models.HTTPError
//	@Failure      404  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /qemu/virtual-machines/{uuid}/shutdown [post]
//
// ShutdownVirtualMachine shuts down a virtual machine
func (s *QemuService) ShutdownVirtualMachine(c echo.Context) error {
	if s.LibVirt == nil {
		return s.Dispatcher.NewInternalServerError("LibVirt connection not available", nil)
	}

	vmUUID := c.Param("uuid")
	if vmUUID == "" {
		return s.Dispatcher.NewBadRequest("Virtual machine UUID is required", nil)
	}

	domain, err := s.getDomainByUUID(vmUUID)
	if err != nil {
		s.Logger.Error("Failed to find domain", "uuid", vmUUID, "error", err)
		return s.Dispatcher.NewNotFound("Virtual machine not found", err)
	}

	if err := s.LibVirt.DomainShutdown(domain); err != nil {
		s.Logger.Error("Failed to shutdown domain", "name", domain.Name, "error", err)
		return s.Dispatcher.NewInternalServerError("Failed to shutdown virtual machine", err)
	}

	return c.JSON(http.StatusOK, models.VMActionResponse{
		Success: true,
		Message: fmt.Sprintf("Virtual machine '%s' shutdown initiated", domain.Name),
	})
}

//	@Summary      Create virtual machine
//	@Description  Create a new virtual machine
//	@Tags         qemu
//	@Accept       json
//	@Param        body  body  models.CreateVMRequest  true  "VM creation parameters"
//	@Produce      json
//	@Success      201  {object}  models.VirtualMachine
//	@Failure      400  {object}  models.HTTPError
//	@Failure      401  {object}  models.HTTPError
//	@Failure      403  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /qemu/virtual-machines [post]
//
// CreateVirtualMachine creates a new virtual machine
func (s *QemuService) CreateVirtualMachine(c echo.Context) error {
	if s.LibVirt == nil {
		return s.Dispatcher.NewInternalServerError("LibVirt connection not available", nil)
	}

	req := new(models.CreateVMRequest)
	if err := c.Bind(req); err != nil {
		return s.Dispatcher.NewBadRequest("Invalid request format", err)
	}

	// Validate request
	if req.Name == "" {
		return s.Dispatcher.NewBadRequest("Virtual machine name is required", nil)
	}
	if req.Memory <= 0 {
		return s.Dispatcher.NewBadRequest("Memory must be greater than 0", nil)
	}
	if req.VCPUs <= 0 {
		return s.Dispatcher.NewBadRequest("VCPUs must be greater than 0", nil)
	}
	if req.DiskSize <= 0 {
		return s.Dispatcher.NewBadRequest("Disk size must be greater than 0", nil)
	}

	// Log the creation attempt
	s.Logger.Info("Creating virtual machine",
		"name", req.Name,
		"memory_mb", req.Memory,
		"vcpus", req.VCPUs,
		"disk_size_gb", req.DiskSize,
		"autostart", req.Autostart,
	)

	// TODO: Implement actual VM creation via libvirt
	// This would require XML domain definition generation and DefineDOM + Create
	// For now, return a placeholder response indicating the feature is pending implementation

	return c.JSON(http.StatusCreated, models.VirtualMachine{
		ID:   0,
		Name: req.Name,
		UUID: "pending-implementation",
	})
}

// Helper function to get domain by UUID
func (s *QemuService) getDomainByUUID(vmUUID string) (libvirt.Domain, error) {
	flags := libvirt.ConnectListDomainsActive | libvirt.ConnectListDomainsInactive
	domains, _, err := s.LibVirt.ConnectListAllDomains(1, flags)
	if err != nil {
		return libvirt.Domain{}, err
	}

	for _, domain := range domains {
		domainUUID, err := uuid.FromBytes(domain.UUID[:])
		if err != nil {
			continue
		}
		if domainUUID.String() == vmUUID {
			return domain, nil
		}
	}

	return libvirt.Domain{}, fmt.Errorf("domain not found")
}
