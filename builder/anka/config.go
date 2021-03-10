//go:generate mapstructure-to-hcl2 -type Config
package anka

import (
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/mitchellh/mapstructure"
)

const DEFAULT_BOOT_DELAY = "10s"

var random *rand.Rand

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	AnkaUser     string `mapstructure:"anka_user"`
	AnkaPassword string `mapstructure:"anka_password"`

	InstallerApp string `mapstructure:"installer_app"`
	SourceVMName string `mapstructure:"source_vm_name"`
	SourceVMTag  string `mapstructure:"source_vm_tag"`

	VMName   string `mapstructure:"vm_name"`
	DiskSize string `mapstructure:"disk_size"`
	RAMSize  string `mapstructure:"ram_size"`
	CPUCount string `mapstructure:"cpu_count"`

	AlwaysFetch bool `mapstructure:"always_fetch"`

	UpdateAddons bool `mapstructure:"update_addons"`

	RegistryName string `mapstructure:"registry_name"`
	RegistryURL  string `mapstructure:"registry_path"`
	NodeCertPath string `mapstructure:"cert"`
	NodeKeyPath  string `mapstructure:"key"`
	CaRootPath   string `mapstructure:"cacert"`
	IsInsecure   bool   `mapstructure:"insecure"`

	PortForwardingRules []struct {
		PortForwardingGuestPort int    `mapstructure:"port_forwarding_guest_port"`
		PortForwardingHostPort  int    `mapstructure:"port_forwarding_host_port"`
		PortForwardingRuleName  string `mapstructure:"port_forwarding_rule_name"`
	} `mapstructure:"port_forwarding_rules,omitempty"`

	HWUUID     string `mapstructure:"hw_uuid,omitempty"`
	BootDelay  string `mapstructure:"boot_delay"`
	EnableHtt  bool   `mapstructure:"enable_htt"`
	DisableHtt bool   `mapstructure:"disable_htt"`
	UseAnkaCP  bool   `mapstructure:"use_anka_cp"`

	StopVM bool `mapstructure:"stop_vm"`

	ctx interpolate.Context //nolint:structcheck
}

func init() {
	random = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func NewConfig(raws ...interface{}) (*Config, error) {
	var c Config

	var md mapstructure.Metadata
	err := config.Decode(&c, &config.DecodeOpts{
		Metadata:    &md,
		Interpolate: true,
	}, raws...)
	if err != nil {
		return nil, err
	}

	if c.BootDelay == "" {
		c.BootDelay = DEFAULT_BOOT_DELAY
	}

	var errs *packer.MultiError

	if c.Comm.Type == "" {
		c.Comm.Type = "anka"
	}

	if c.InstallerApp == "" && c.SourceVMName == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("installer_app or source_vm_name must be specified"))
	}

	if c.InstallerApp != "" && c.SourceVMName != "" {
		errs = packer.MultiErrorAppend(errs, errors.New("cannot specify both an installer_app and source_vm_name"))
	}

	if c.VMName == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("please specify a name for your vm"))
	}

	if c.SourceVMName != "" && strings.ContainsAny(c.SourceVMName, " \n") {
		errs = packer.MultiErrorAppend(errs, errors.New("source_vm_name name contains spaces"))
	}

	if len(c.PortForwardingRules) > 0 {
		for index, rule := range c.PortForwardingRules {
			if rule.PortForwardingGuestPort == 0 {
				errs = packer.MultiErrorAppend(errs, errors.New("guest port is required"))
			}
			if rule.PortForwardingRuleName == "" {
				c.PortForwardingRules[index].PortForwardingRuleName = randSeq(10)
			}
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	return &c, nil
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[random.Intn(len(letters))]
	}
	return string(b)
}
