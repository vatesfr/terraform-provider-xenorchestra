module github.com/ddelnano/terraform-provider-xenorchestra

go 1.16

require (
	github.com/ddelnano/terraform-provider-xenorchestra/client v0.0.0-00010101000000-000000000000
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.4.3
)

replace github.com/ddelnano/terraform-provider-xenorchestra/client => ./client
