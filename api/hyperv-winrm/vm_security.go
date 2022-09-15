package hyperv_winrm

import (
	"context"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/taliesins/terraform-provider-hyperv/api"
)

type createOrUpdateVmSecurityArgs struct {
	VmSecurityJson string
}

var createOrUpdateVmSecurityTemplate = template.Must(template.New("CreateOrUpdateVmSecurity").Parse(`
$ErrorActionPreference = 'Stop'
Import-Module Hyper-V
$vmSecurity = '{{.VmSecurityJson}}' | ConvertFrom-Json

$currentKeyProtector = Get-VMKeyProtector -VMName $vmSecurity.VmName

if (($vmSecurity.TpmEnabled -or $vmSecurity.Shielded) -and (($null -eq $currentKeyProtector) -or ("AAAABA==" -eq [Convert]::ToBase64String($currentKeyProtector)))) {
	Set-VMKeyProtector -VMName $vmSecurity.VmName -NewLocalKeyProtector
}

if ($vmSecurity.TpmEnabled -eq $true) {
	Enable-VMTPM -VMName $vmSecurity.VmName
}
else {
	Disable-VMTPM -VMName $vmSecurity.VmName
}

$SetVMSecurityArgs = @{
	VMName = $vmSecurity.VmName
	EncryptStateAndVmMigrationTraffic = $vmSecurity.EncryptStateAndVmMigrationTraffic
	VirtualizationBasedSecurityOptOut = $vmSecurity.VirtualizationBasedSecurityOptOut
}

Set-VMSecurity @SetVMSecurityArgs

$SetVMSecurityPolicy = @{
	VMName = $vmSecurity.VmName
	Shielded = $vmSecurity.Shielded
	BindToHostTpm = $vmSecurity.BindToHostTpm
}

Set-VMSecurityPolicy @SetVMSecurityPolicy
`))

func (c *ClientConfig) CreateOrUpdateVmSecurity(
	ctx context.Context,
	vmName string,
	encryptStateAndVmMigrationTraffic bool,
	virtualizationBasedSecurityOptOut bool,
	tpmEnabled bool,
	shielded bool,
	bindToHostTpm bool,
) (err error) {
	vmSecurityJson, err := json.Marshal(api.VmSecurity{
		VmName:                            vmName,
		EncryptStateAndVmMigrationTraffic: encryptStateAndVmMigrationTraffic,
		VirtualizationBasedSecurityOptOut: virtualizationBasedSecurityOptOut,
		TpmEnabled:                        tpmEnabled,
		Shielded:                          shielded,
		BindToHostTpm:                     bindToHostTpm,
	})

	if err != nil {
		return err
	}

	err = c.WinRmClient.RunFireAndForgetScript(ctx, createOrUpdateVmSecurityTemplate, createOrUpdateVmSecurityArgs{
		VmSecurityJson: string(vmSecurityJson),
	})

	return err
}

type getVmSecurityArgs struct {
	VmName string
}

var getVmSecurityTemplate = template.Must(template.New("GetVmSecurity").Parse(`
$ErrorActionPreference = 'Stop'

$vmSecurityObject =  Get-VMSecurity -VMName '{{.VmName}}' | Select-Object EncryptStateAndVmMigrationTraffic,VirtualizationBasedSecurityOptOut,TpmEnabled,Shielded,BindToHostTpm

if ($vmSecurityObject) {
	$vmSecurity = ConvertTo-Json -InputObject $vmSecurityObject
	$vmSecurity
} else {
	"{}"
}
`))

func (c *ClientConfig) GetVmSecurity(ctx context.Context, vmName string) (result api.VmSecurity, err error) {
	err = c.WinRmClient.RunScriptWithResult(ctx, getVmSecurityTemplate, getVmSecurityArgs{
		VmName: vmName,
	}, &result)

	return result, err
}

func (c *ClientConfig) GetNoVmSecurities(ctx context.Context) (result []api.VmSecurity) {
	result = make([]api.VmSecurity, 0)
	return result
}

func (c *ClientConfig) GetVmSecurities(ctx context.Context, vmName string) (result []api.VmSecurity, err error) {
	result = make([]api.VmSecurity, 0)
	vmSecurity, err := c.GetVmSecurity(ctx, vmName)
	if err != nil {
		return result, err
	}
	result = append(result, vmSecurity)
	return result, err
}

func (c *ClientConfig) CreateOrUpdateVmSecurities(ctx context.Context, vmName string, vmSecurities []api.VmSecurity) (err error) {
	if len(vmSecurities) == 0 {
		return nil
	}
	if len(vmSecurities) > 1 {
		return fmt.Errorf("Only 1 vm security setting allowed per a vm")
	}

	vmSecurity := vmSecurities[0]

	return c.CreateOrUpdateVmSecurity(ctx, vmName,
		vmSecurity.EncryptStateAndVmMigrationTraffic,
		vmSecurity.VirtualizationBasedSecurityOptOut,
		vmSecurity.TpmEnabled,
		vmSecurity.Shielded,
		vmSecurity.BindToHostTpm,
	)
}
