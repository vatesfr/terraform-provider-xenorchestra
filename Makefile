.PHONY: import

import:
	go build -o terraform-provider-xenorchestra
	terraform init
	terraform import xenorchestra_vm.testing 77c6637c-fa3d-0a46-717e-296208c40169
