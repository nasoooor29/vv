package models

// VirtualMachine represents a QEMU virtual machine
type VirtualMachine struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	UUID string `json:"uuid"`
}

// VirtualMachineInfo contains detailed information about a virtual machine
type VirtualMachineInfo struct {
	State     uint8  `json:"state"`
	MaxMemKB  uint64 `json:"max_mem_kb"`
	MemoryKB  uint64 `json:"memory_kb"`
	VCPUs     uint16 `json:"vcpus"`
	CPUTimeNs uint64 `json:"cpu_time_ns"`
}

// VirtualMachineWithInfo combines VM details with runtime information
type VirtualMachineWithInfo struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	UUID string `json:"uuid"`
	VirtualMachineInfo
}

// CreateVMRequest represents a request to create a new virtual machine
type CreateVMRequest struct {
	Name      string `json:"name" validate:"required"`
	Memory    int64  `json:"memory" validate:"required"`
	VCPUs     int32  `json:"vcpus" validate:"required"`
	DiskSize  int64  `json:"disk_size" validate:"required"`
	OSImage   string `json:"os_image"`
	Autostart bool   `json:"autostart"`
}

// VMActionResponse represents the response from VM control operations
type VMActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// QEMU VM States (libvirt domain states)
const (
	VIR_DOMAIN_NOSTATE     = iota // No state
	VIR_DOMAIN_RUNNING            // The domain is running
	VIR_DOMAIN_BLOCKED            // The domain is blocked on resource
	VIR_DOMAIN_PAUSED             // The domain is paused by user
	VIR_DOMAIN_SHUTDOWN           // The domain is being shut down
	VIR_DOMAIN_SHUTOFF            // The domain is shut off
	VIR_DOMAIN_CRASHED            // The domain is crashed
	VIR_DOMAIN_PMSUSPENDED        // The domain is suspended by guest power management
)
