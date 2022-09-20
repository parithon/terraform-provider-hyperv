$ModuleVersion = "1.0.3"
$Path = "../Terraform/terraform.d/plugins/registry.terraform.io/taliesins/hyperv"
go build -o "$Path/$ModuleVersion/windows_amd64/terraform-provider-hyperv_${ModuleVersion}.exe"
