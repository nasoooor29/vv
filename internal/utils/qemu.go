package utils

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/gofrs/uuid"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
)

// Create Qcow2 disk image in the spicified path/size size in Megabytes
func createDiskImage(path string, size uint) (string, error) {
	format := "qcow2"
	filename := fmt.Sprintf("%v.%v", path, format)
	cmd := exec.Command("qemu-img", "create", "-f", format, filename, fmt.Sprintf("%vM", size))
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return filename, nil
}

// Builds domain xml for libvirt
type LibVirtDomainParams struct {
	Name                  string
	DiskLocation          string
	InstallationMediaPath string
	MemorySize            uint
	VirtualCpus           uint
	DiskSize              uint
	SpiceListenPort       int
	SpiceListenIpAddr     string
	VNCListenPort         int
	VNCListenIpAddr       string
}

// BuildLibVirtDomain and create disk image using provided params, returns DomainXML for libvirt
func BuildLibVirtDomain(p *LibVirtDomainParams) (string, error) {
	uuid, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	diskImage, err := createDiskImage(filepath.Join(p.DiskLocation, uuid.String()), p.DiskSize)
	if err != nil {
		return "", err
	}
	graphics := []libvirtxml.DomainGraphic{
		{
			Spice: &libvirtxml.DomainGraphicSpice{
				Port:     int(p.SpiceListenPort),
				AutoPort: "yes",
				Listen:   p.SpiceListenIpAddr,
				Image: &libvirtxml.DomainGraphicSpiceImage{
					Compression: "off",
				},
			},

			VNC: &libvirtxml.DomainGraphicVNC{
				Port:     int(p.VNCListenPort),
				AutoPort: "yes",
				Listen:   p.VNCListenIpAddr,
			},
		},
	}
	if p.VNCListenPort < 0 {
		graphics = []libvirtxml.DomainGraphic{
			{
				VNC: &libvirtxml.DomainGraphicVNC{
					Port:     int(p.VNCListenPort),
					AutoPort: "yes",
					Listen:   p.VNCListenIpAddr,
				},
			},
		}
	}
	if p.SpiceListenPort < 0 {
		graphics = []libvirtxml.DomainGraphic{
			{
				Spice: &libvirtxml.DomainGraphicSpice{
					Port:     int(p.SpiceListenPort),
					AutoPort: "yes",
					Listen:   p.SpiceListenIpAddr,
					Image: &libvirtxml.DomainGraphicSpiceImage{
						Compression: "off",
					},
				},
			},
		}
	}

	dom := libvirtxml.Domain{
		UUID: uuid.String(),
		Type: "kvm",
		Name: p.Name,
		Memory: &libvirtxml.DomainMemory{
			Value: p.MemorySize,
			Unit:  "MiB",
		},
		CurrentMemory: &libvirtxml.DomainCurrentMemory{
			Value: p.MemorySize,
			Unit:  "MiB",
		},
		VCPU: &libvirtxml.DomainVCPU{
			Placement: "static",
			Value:     p.VirtualCpus,
		},
		OS: &libvirtxml.DomainOS{
			Type: &libvirtxml.DomainOSType{
				Arch:    "x86_64",
				Machine: "pc-q35-10.0",
				Type:    "hvm",
			},
		},
		CPU: &libvirtxml.DomainCPU{
			Mode: "host-passthrough",
		},
		Devices: &libvirtxml.DomainDeviceList{
			Emulator: "/usr/bin/qemu-system-x86_64",
			Disks: []libvirtxml.DomainDisk{
				{
					Driver: &libvirtxml.DomainDiskDriver{
						Name: "qemu",
						Type: "qcow2",
					},
					Source: &libvirtxml.DomainDiskSource{
						File: &libvirtxml.DomainDiskSourceFile{
							File: diskImage,
						},
						Index: 0,
					},
					BackingStore: &libvirtxml.DomainDiskBackingStore{},
					Target: &libvirtxml.DomainDiskTarget{
						Dev: "vda",
						Bus: "virtio",
					},
					Alias: &libvirtxml.DomainAlias{
						Name: "virtio-disk0",
					},
				},
				{
					Driver: &libvirtxml.DomainDiskDriver{
						Name: "qemu",
					},
					Device: "cdrom",
					Boot: &libvirtxml.DomainDeviceBoot{
						Order: 1,
					},
					Target: &libvirtxml.DomainDiskTarget{
						Dev: "sda",
						Bus: "sata",
					},
					Source: &libvirtxml.DomainDiskSource{
						File: &libvirtxml.DomainDiskSourceFile{
							File: p.InstallationMediaPath,
						},
						Index: 1,
					},
					ReadOnly: &libvirtxml.DomainDiskReadOnly{},
					Alias: &libvirtxml.DomainAlias{
						Name: "sata0-0-0",
					},
					Address: &libvirtxml.DomainAddress{
						Drive: &libvirtxml.DomainAddressDrive{},
					},
				},
			},
			Controllers: []libvirtxml.DomainController{
				{
					Type: "sata",
					Alias: &libvirtxml.DomainAlias{
						Name: "ide",
					},
					Index: new(uint),
					Address: &libvirtxml.DomainAddress{
						PCI: &libvirtxml.DomainAddressPCI{
							Domain:   new(uint),
							Bus:      new(uint),
							Slot:     new(uint),
							Function: new(uint),
						},
					},
				},
			},
			Interfaces: []libvirtxml.DomainInterface{
				{
					Source: &libvirtxml.DomainInterfaceSource{
						Network: &libvirtxml.DomainInterfaceSourceNetwork{
							Network: "default",
							Bridge:  "virbr0",
						},
					},
					Target: &libvirtxml.DomainInterfaceTarget{
						Dev: "vnet2",
					},
					Model: &libvirtxml.DomainInterfaceModel{
						Type: "virtio",
					},
					Alias: &libvirtxml.DomainAlias{
						Name: "net0",
					},
					Address: &libvirtxml.DomainAddress{
						PCI: &libvirtxml.DomainAddressPCI{
							Domain:   new(uint),
							Bus:      new(uint),
							Slot:     new(uint),
							Function: new(uint),
						},
					},
				},
			},
			Graphics: graphics,
		},
	}
	return dom.Marshal()
}

// Extract VNC configuration from domain XML
func VNCFromDomainXML(domainXML string) (string, int, error) {
	var dom libvirtxml.Domain
	if err := dom.Unmarshal(domainXML); err != nil {
		return "", 0, err
	}
	for _, g := range dom.Devices.Graphics {
		if g.VNC != nil {
			return g.VNC.Listen, g.VNC.Port, nil
		}
	}
	return "", 0, fmt.Errorf("no VNC configuration found in domain XML")
}
