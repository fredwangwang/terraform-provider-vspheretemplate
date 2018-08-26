package options

import (
	"encoding/json"
	"fmt"
	"github.com/vmware/govmomi/ovf"
	"github.com/vmware/govmomi/vim25/types"
)

type Property struct {
	types.KeyValue
	Spec *ovf.Property `json:",omitempty"`
}

type Network struct {
	Name    string
	Network string
}

type Options struct {
	AllDeploymentOptions []string `json:",omitempty"`
	Deployment           string   `json:",omitempty"`

	AllDiskProvisioningOptions []string `json:",omitempty"`
	DiskProvisioning           string

	AllIPAllocationPolicyOptions []string `json:",omitempty"`
	IPAllocationPolicy           string

	AllIPProtocolOptions []string `json:",omitempty"`
	IPProtocol           string

	PropertyMapping []Property `json:",omitempty"`

	NetworkMapping []Network `json:",omitempty"`

	Annotation string `json:",omitempty"`

	MarkAsTemplate bool
	PowerOn        bool
	InjectOvfEnv   bool
	WaitForIP      bool
	Name           *string
}

func FromInterface(i interface{}) (Options, error) {
	var options Options

	optionsContent, err := json.Marshal(i)
	if err != nil {
		return options, fmt.Errorf("failed to marshal the given options: %s", err)
	}

	err = json.Unmarshal(optionsContent, &options)
	if err != nil {
		return options, fmt.Errorf("failed to unmarshal the given options into Options struct: %s", err)
	}

	return options, nil
}
