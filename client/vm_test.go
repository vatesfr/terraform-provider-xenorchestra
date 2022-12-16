package client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
)

var vmObjectData string = `
{
  "type": "VM",
  "addresses": {},
  "auto_poweron": false,
  "boot": {
    "order": "ncd"
  },
  "CPUs": {
    "max": 1,
    "number": 1
  },
  "current_operations": {},
  "expNestedHvm": false,
  "high_availability": "",
  "memory": {
    "dynamic": [
      1073741824,
      1073741824
    ],
    "static": [
      536870912,
      1073741824
    ],
    "size": 1073733632
  },
  "installTime": 1552287083,
  "resourceSet": "U8kmJKszJC0",
  "name_description": "Testingsdfsdf",
  "name_label": "Hello from terraform!",
  "other": {
    "xo:resource_set": "\"U8kmJKszJC0\"",
    "base_template_name": "Ubuntu Bionic Beaver 18.04",
    "import_task": "OpaqueRef:a1c9c64b-eeec-48cd-b587-51a8fc7924d0",
    "mac_seed": "bc583da8-d7a4-9437-7630-ce5ecce7efd0",
    "install-methods": "cdrom,nfs,http,ftp",
    "linux_template": "true"
  },
  "os_version": {},
  "power_state": "Running",
  "hasVendorDevice": false,
  "snapshots": [],
  "startTime": 1552445802,
  "tags": [],
  "VIFs": [
    "13793e84-110e-7f0d-8544-cb2f39adf2f4"
  ],
  "virtualizationMode": "hvm",
  "xenTools": false,
  "$container": "a5c7d15c-2724-47ce-8e30-77f21f08bb4c",
  "$VBDs": [
    "43202d8a-c2ba-963e-54d0-b8cf770b2725",
    "5b8adb53-36a7-8489-3a5f-f4ec8aa79568",
    "5a12a8dd-f217-9f90-5aab-8f434893a6e3",
    "567d6474-62ac-2d98-66ed-53cf468fc8b3",
    "42eeb5e0-b14e-99c0-36a6-61b671112353"
  ],
  "VGPUs": [],
  "$VGPUs": [],
  "vga": "cirrus",
  "videoram": "8",
  "id": "77c6637c-fa3d-0a46-717e-296208c40169",
  "uuid": "77c6637c-fa3d-0a46-717e-296208c40169",
  "$pool": "cadf25ab-91ff-6fc0-041f-5a7033c4bc78",
  "$poolId": "cadf25ab-91ff-6fc0-041f-5a7033c4bc78"
}`

var vmObjectWithNumericVideoram string = `
{
  "videoram": 8
}`

var data string = `
{
  "6944cce9-5ce0-a853-ee9c-bcc8281b597f": {
    "body": "VM 'Test VM' started on host: xenserver-ddelnano (uuid: a5c7d15c-2724-47ce-8e30-77f21f08bb4c)",
    "name": "VM_STARTED",
    "time": 1547577637,
    "$object": "5f318ba2-2300-cc34-3710-f64f53634ac0",
    "id": "6944cce9-5ce0-a853-ee9c-bcc8281b597f",
    "type": "message",
    "uuid": "6944cce9-5ce0-a853-ee9c-bcc8281b597f",
    "$pool": "cadf25ab-91ff-6fc0-041f-5a7033c4bc78",
    "$poolId": "cadf25ab-91ff-6fc0-041f-5a7033c4bc78"
  },
  "52e21132-2f4f-3b80-ef65-7d43750eb6db": {
    "body": "VM 'XOA' started on host: xenserver-ddelnano (uuid: a5c7d15c-2724-47ce-8e30-77f21f08bb4c)",
    "name": "VM_STARTED",
    "time": 1547578119,
    "$object": "9df00260-c6d8-6cd4-4b24-d9e23602bf95",
    "id": "52e21132-2f4f-3b80-ef65-7d43750eb6db",
    "type": "message",
    "uuid": "52e21132-2f4f-3b80-ef65-7d43750eb6db",
    "$pool": "cadf25ab-91ff-6fc0-041f-5a7033c4bc78",
    "$poolId": "cadf25ab-91ff-6fc0-041f-5a7033c4bc78"
  },
  "3bdc3bed-2a91-12af-5154-097706009593": {
    "body": "VM 'XOA' started on host: xenserver-ddelnano (uuid: a5c7d15c-2724-47ce-8e30-77f21f08bb4c)",
    "name": "VM_STARTED",
    "time": 1547578434,
    "$object": "2e793135-3b3d-354e-46c4-ef99fe0a37d0",
    "id": "3bdc3bed-2a91-12af-5154-097706009593",
    "type": "message",
    "uuid": "3bdc3bed-2a91-12af-5154-097706009593",
    "$pool": "cadf25ab-91ff-6fc0-041f-5a7033c4bc78",
    "$poolId": "cadf25ab-91ff-6fc0-041f-5a7033c4bc78"
  },
  "ad97f23d-3bf0-7fac-e70c-ab98661450c6": {
    "body": "VM 'XOA' shutdown forcibly",
    "name": "VM_SHUTDOWN",
    "time": 1547578721,
    "$object": "9df00260-c6d8-6cd4-4b24-d9e23602bf95",
    "id": "ad97f23d-3bf0-7fac-e70c-ab98661450c6",
    "type": "message",
    "uuid": "ad97f23d-3bf0-7fac-e70c-ab98661450c6",
    "$pool": "cadf25ab-91ff-6fc0-041f-5a7033c4bc78",
    "$poolId": "cadf25ab-91ff-6fc0-041f-5a7033c4bc78"
  },
  "77c6637c-fa3d-0a46-717e-296208c40169": {
    "body": "VM 'XOA' shutdown forcibly",
    "name": "VM_SHUTDOWN",
    "time": 1547578721,
    "$object": "9df00260-c6d8-6cd4-4b24-d9e23602bf95",
    "id": "77c6637c-fa3d-0a46-717e-296208c40169",
    "type": "VM",
    "uuid": "ad97f23d-3bf0-7fac-e70c-ab98661450c6",
    "$pool": "cadf25ab-91ff-6fc0-041f-5a7033c4bc78",
    "$poolId": "cadf25ab-91ff-6fc0-041f-5a7033c4bc78"
  }
}
`

func TestUnmarshal(t *testing.T) {
	var allObjectRes allObjectResponse
	err := json.Unmarshal([]byte(data), &allObjectRes.Objects)

	if err != nil || allObjectRes.Objects["77c6637c-fa3d-0a46-717e-296208c40169"].Id != "77c6637c-fa3d-0a46-717e-296208c40169" {
		t.Fatalf("error: %v. Need to have VM object: %v", err, allObjectRes)
	}
}

func TestUnmarshalingVmObject(t *testing.T) {
	var vmObj Vm

	err := json.Unmarshal([]byte(vmObjectData), &vmObj)

	if err != nil {
		t.Fatalf("error: %v. Need to have VM object: %v", err, vmObj)
	}

	if !validateVmObject(vmObj) {
		t.Fatalf("VmObject has not passed validation")
	}

	var vmObjVideoramNumeric Vm
	err = json.Unmarshal([]byte(vmObjectWithNumericVideoram), &vmObjVideoramNumeric)

	if err != nil {
		t.Fatalf("error: %v. Need to have VM object: %v", err, vmObjVideoramNumeric)
	}

	if vmObjVideoramNumeric.Videoram.Value != 8 {
		t.Fatalf("Expected vm to Unmarshal from numerical videoram value")
	}

	if vmObj.ManagementAgentDetected != false {
		t.Fatalf("expected vm to Unmarshal to 'false' when `managementAgentDetected` not present")
	}
}

func TestFlatResourceSetStringerInterface(t *testing.T) {
	rs := &FlatResourceSet{
		Id: "id",
	}
	v := fmt.Sprintf("%s", rs)
	if v != rs.Id {
		t.Errorf("expected FlatResourceSet to print Id '%s' value rather than value '%s'", rs.Id, v)
	}
}

func TestFlatResourceSetMarshalling(t *testing.T) {
	rs := &FlatResourceSet{
		Id: "id",
	}

	b, err := json.Marshal(rs)

	if err != nil {
		t.Fatalf("failed to marshal resource set with error: %v", err)
	}

	expected := fmt.Sprintf(`"%s"`, rs.Id)
	if s := string(b); s != expected {
		t.Errorf("expected '%s' to equal '%s' after marshalling", s, expected)
	}
}

func TestFlatResourceSetMarshallingForEmptyString(t *testing.T) {
	rs := &FlatResourceSet{
		Id: "",
	}

	b, err := json.Marshal(rs)

	if err != nil {
		t.Fatalf("failed to marshal resource set with error: %v", err)
	}

	expected := `null`
	if s := string(b); s != expected {
		t.Errorf("expected '%s' to equal '%s' after marshalling", s, expected)
	}
}

func TestFlatResourceSetUnmarshalling(t *testing.T) {
	id := "id"
	s := struct {
		ResourceSet *FlatResourceSet `json:"resourceSet,omitempty"`
	}{
		ResourceSet: &FlatResourceSet{},
	}
	data := []byte(fmt.Sprintf(`{"resourceSet": "%s"}`, id))
	err := json.Unmarshal(data, &s)

	if err != nil {
		t.Fatalf("failed to unmarshal into FlatResourceSet with error: %v", err)
	}

	rs := s.ResourceSet
	if rs.Id != id {
		t.Errorf("expected '%s' but received resource set '%v' with actual value '%s'", id, rs, rs.Id)
	}
}

func TestUpdateVmWithUpdatesThatRequireHalt(t *testing.T) {
	c, err := NewClient(GetConfigFromEnv())
	if err != nil {
		t.Fatalf("failed to create client with error: %v", err)
	}

	prevCPUs := accVm.CPUs.Number
	updatedCPUs := prevCPUs + 1
	err = c.HaltVm(accVm.Id)

	if err != nil {
		t.Fatalf("failed to halt vm ahead of update: %v", err)
	}

	vm, err := c.UpdateVm(Vm{Id: accVm.Id, CPUs: CPUs{Number: updatedCPUs}, NameLabel: "terraform testing", Memory: MemoryObject{Static: []int{0, 4294967296}}})

	if err != nil {
		t.Fatalf("failed to update vm with error: %v", err)
	}

	err = c.StartVm(vm.Id)

	if err != nil {
		t.Fatalf("failed to start vm after update: %v", err)
	}

	if vm.CPUs.Number != updatedCPUs {
		t.Errorf("failed to update VM cpus to %d", updatedCPUs)
	}
}

func validateVmObject(o Vm) bool {
	if o.Type != "VM" {
		return false
	}

	if o.CPUs.Number != 1 {
		return false
	}

	if o.Memory.Size != 1073733632 {
		return false
	}

	if o.PowerState != "Running" {
		return false
	}
	if o.VirtualizationMode != "hvm" {
		return false
	}

	if o.VIFs[0] != "13793e84-110e-7f0d-8544-cb2f39adf2f4" {
		return false
	}

	if o.PoolId != "cadf25ab-91ff-6fc0-041f-5a7033c4bc78" {
		return false
	}

	if o.ResourceSet.Id == "" {
		return false
	}

	if o.Videoram.Value != 8 {
		return false
	}

	return true
}

var cloudConfig string = "#cloud-config"

func Test_warnOnInvalidCloudConfigLogsWarningOnInvalidContent(t *testing.T) {
	var logBuf bytes.Buffer
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	log.SetOutput(&logBuf)
	warnOnInvalidCloudConfig("This is not valid")
	if !strings.Contains(logBuf.String(), "WARNING") {
		t.Errorf("This shoud trigger a warning log")
	}
}

func Test_warnOnInvalidCloudConfigRecognizesRawText(t *testing.T) {
	var logBuf bytes.Buffer
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	log.SetOutput(&logBuf)
	warnOnInvalidCloudConfig(cloudConfig)
	if strings.Contains(logBuf.String(), "WARNING") {
		t.Errorf("#cloud-config should be recognized as valid cloud config")
	}
}

func Test_warnOnInvalidCloudConfigRecognizesGzip(t *testing.T) {
	var logBuf bytes.Buffer
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)

	_, err := zw.Write([]byte(cloudConfig))

	if err != nil {
		t.Fatalf("failed to compress data necessary for test: %v", err)
	}

	log.SetOutput(&logBuf)
	warnOnInvalidCloudConfig(buf.String())
	if strings.Contains(logBuf.String(), "WARNING") {
		t.Errorf("gzipped content should be recognized as valid cloud config")
	}
}

func Test_warnOnInvalidCloudConfigRecognizesMultipartMIME(t *testing.T) {
	var logBuf bytes.Buffer
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	// This was generated by using the terraform provider for cloudinit. Terraform code used
	// provided below:
	//
	// data "cloudinit_config" "foo" {
	//      gzip = false
	//      base64_encode = false
	//
	//      part {
	//	  content_type = "text/x-shellscript"
	//        content = "baz"
	//      }
	// }
	s := "Content-Type: multipart/mixed; boundary=\"MIMEBOUNDARY\"\nMIME-Version: 1.0\r\n\r\n--MIMEBOUNDARY\r\nContent-Transfer-Encoding: 7bit\r\nContent-Type: text/x-shellscript\r\nMime-Version: 1.0\r\n\r\nbaz\r\n--MIMEBOUNDARY--\r\n"
	log.SetOutput(&logBuf)
	warnOnInvalidCloudConfig(s)
	if strings.Contains(logBuf.String(), "WARNING") {
		t.Errorf("multipart MIME archives should be recognized as valid cloud config")
	}
}
