package api

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ExpandVmSecurities(d *schema.ResourceData) ([]VmSecurity, error) {
	expandedVmSecurities := make([]VmSecurity, 0)

	if v, ok := d.GetOk("vm_security"); ok {
		vmSecurities := v.([]interface{})
		for _, security := range vmSecurities {
			security, ok := security.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("[ERROR][hyperv] vm_security should be a Hash - was '%+v'", security)
			}

			log.Printf("[DEBUG] security = [%+v]", security)

			expandedVmSecurity := VmSecurity{
				EncryptStateAndVmMigrationTraffic: security["encrypt_state_and_vm_migration_traffic"].(bool),
				VirtualizationBasedSecurityOptOut: security["virtualization_based_security_optout"].(bool),
				TpmEnabled:                        security["tpm_enabled"].(bool),
				Shielded:                          security["shielded"].(bool),
				BindToHostTpm:                     security["bind_to_host_tpm"].(bool),
			}

			expandedVmSecurities = append(expandedVmSecurities, expandedVmSecurity)
		}
	}

	if len(expandedVmSecurities) < 1 {
		vmSecurity := VmSecurity{
			EncryptStateAndVmMigrationTraffic: false,
			VirtualizationBasedSecurityOptOut: false,
			TpmEnabled:                        false,
			Shielded:                          false,
			BindToHostTpm:                     false,
		}

		expandedVmSecurities = append(expandedVmSecurities, vmSecurity)
	}

	return expandedVmSecurities, nil
}

func FlattenVmSecurities(vmSecurities *[]VmSecurity) []interface{} {
	if vmSecurities == nil || len(*vmSecurities) < 1 {
		return nil
	}

	flattenedVmSecurities := make([]interface{}, 0)

	for _, vmSecurity := range *vmSecurities {
		flattenedVmSecurity := make(map[string]interface{})
		flattenedVmSecurity["encrypt_state_and_vm_migration_traffic"] = vmSecurity.EncryptStateAndVmMigrationTraffic
		flattenedVmSecurity["virtualization_based_security_optout"] = vmSecurity.VirtualizationBasedSecurityOptOut
		flattenedVmSecurity["tpm_enabled"] = vmSecurity.TpmEnabled
		flattenedVmSecurity["shielded"] = vmSecurity.Shielded
		flattenedVmSecurity["bind_to_host_tpm"] = vmSecurity.BindToHostTpm

		flattenedVmSecurities = append(flattenedVmSecurities, flattenedVmSecurity)
	}

	return flattenedVmSecurities
}

type VmSecurity struct {
	VmName                            string
	EncryptStateAndVmMigrationTraffic bool
	VirtualizationBasedSecurityOptOut bool
	TpmEnabled                        bool
	Shielded                          bool
	BindToHostTpm                     bool
}

type HypervVmSecurityClient interface {
	CreateOrUpdateVmSecurity(
		ctx context.Context,
		vmName string,
		encryptStateAndVmMigrationTraffic bool,
		virtualizationBasedSecurityOptOut bool,
		tpmEnabled bool,
		shielded bool,
		bindToHostTpm bool,
	) (err error)
	GetVmSecurity(ctx context.Context, vmName string) (result VmSecurity, err error)
	GetNoVmSecurities(ctx context.Context) (result []VmSecurity)
	GetVmSecurities(ctx context.Context, vmName string) (result []VmSecurity, err error)
	CreateOrUpdateVmSecurities(ctx context.Context, vmName string, vmFirmwares []VmSecurity) (err error)
}
