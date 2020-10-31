package xoa

import "testing"

// map[auto_poweron:false cloud_config: core_os:false cpu_cap:0 cpu_weight:0 cpus:1 disk:[map[name_label:xo provider root size:1e+10 sr_id:86a9757d-9c05-9fe0-e79a-8243cb1f37f3]] high_availability: id:df123bde-d52a-bdae-c763-6887b4822886 memory_max:4.295e+09 name_description:description name_label:Terraform - vif testing network:[map[attached:true device:0 mac_address:f2:f3:e9:d5:f5:a3 network_id:a12df741-f34f-7d05-f120-462f0ab39a48]] resource_set: template:08701be4-52f0-c201-a208-16e765d33e5e timeouts:<nil>]
func Test_migrateXOVmStateV0ToV1(t *testing.T) {

	rawState := map[string]interface{}{
		"network": []map[string]interface{}{
			{
				"attached":    true,
				"device":      "1",
				"mac_address": "00:00:00:00:00:00",
				"network_id":  "id",
			},
			{
				"attached":    true,
				"device":      "0",
				"mac_address": "00:00:00:00:00:00",
				"network_id":  "id",
			},
		},
	}

	newState, _ := migrateXOVmStateV0ToV1(rawState, nil)

	var networks []map[string]interface{}
	var ok bool
	if networks, ok = newState["network"].([]map[string]interface{}); !ok {
		t.Fatalf("expected terraform state to have a 'network' attribute. Instead received %+v", newState)
	}

	if networks[0]["device"].(string) != "0" || networks[1]["device"].(string) != "1" {
		t.Errorf("expected the terraform state to be ordered by device ID. Instead received %+v", networks)
	}
}
