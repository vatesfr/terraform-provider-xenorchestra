## Contributing

Any and all contributions are welcome! Don't hesitate to reach out to ask if you have a work in progress (WIP) pull request, an issue without much background, etc. I will do my best to help anyone who is willing to contribute.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.12+
- [Go](https://golang.org/doc/install) 1.16 (to build the provider plugin)
- git lfs must be installed and `git lfs install` must be run after cloning

## Testing the Provider

The provider has two types of tests: client integration and terraform acceptance tests.

*Note:* The client integration and acceptance tests create real resources! The test suite will create and remove resources during the test run but it's possible that crashing the provider will leave resources dangling. Until [#84](https://github.com/terra-farm/terraform-provider-xenorchestra/issues/84) is done you may need to re-run the test suite or clean up some of the state yourself.


### Running the tests

The following environment variables must be set:
- XOA_URL - the url used to connect to your XO server (ws://yourdomain.com)
- XOA_USER - the username of a user with admin privileges
- XOA_PASSWORD - the password of the associated user
- XOA_POOL - The XO pool you want to target when running the tests. VMs, storage repositories and other resources will be created / launched on this pool
- XOA_TEMPLATE - A VM template that has an existing OS **already installed**
- XOA_DISKLESS_TEMPLATE - A VM template that does not have an existing OS (found from `xe template-list`)
- XOA_ISO - The name of an ISO that exists on the same pool as `XOA_POOL`
- XOA_ISO_SR - The name of an ISO storage repository that exists on the same pool as `XOA_POOL`. This SR must be writable since the tests will upload an ISO to it.
- XOA_NETWORK - The name of a network that is PXE capable. If a non PXE capable network is used some tests may fail.
- XOA_PIF - The UUID of a PIF that will be used for testing VLAN network creation. This has the potential to disrupt network traffic, so this PIF should be an unused interface.

I typically keep these in a ~/.xoa file and run the following before running the test suite

```bash
# See the contents of ~/.xoa
$ cat ~/.xoa
export XOA_URL=ws://yourdomain.com
export XOA_USER=username
export XOA_PASSWORD=password
export XOA_POOL=pool-1
export XOA_TEMPLATE='Debian 10 Cloudinit'

[ ... ]

# Source the environment variables inside the file
eval $(cat ~/.xoa)
```

The following command can be used to run to pass a test name into go's `-run` flag (docs [here](https://tip.golang.org/cmd/go/#hdr-Testing_flags)). This is helpful for running a subset of the test when working on new functionality.

```
TEST=TestAccXONetworkDataSource_read make testacc

# Increase terraform's logging to debug for more insight into a failure
TF_LOG=debug TEST=TestAccXONetworkDataSource_read make testacc
```

The following command can be used to run the entire test suite.

```
# This runs the testclient and testacc Makefile targets
make test
```

