package client

import (
	"reflect"
	"testing"
)

func TestHostCompare(t *testing.T) {
	tests := []struct {
		other  Host
		host   Host
		result bool
	}{
		{
			other: Host{
				Id:        "788e1dce-44f6-4db7-ae62-185c69fecd3b",
				NameLabel: "xcp-host1-k8s.domain.eu",
				Pool:      "pool id",
			},
			host:   Host{NameLabel: "xcp-host1-k8s.domain.eu"},
			result: true,
		},
		{
			other: Host{
				Id:        "788e1dce-44f6-4db7-ae62-185c69fecd3b",
				NameLabel: "xcp-host2-k8s.domain.eu",
				Pool:      "pool id",
			},
			host:   Host{NameLabel: "xcp-host1-k8s.domain.eu"},
			result: false,
		},
	}

	for _, test := range tests {
		host := test.host
		other := test.other
		result := test.result
		if host.Compare(other) != result {
			t.Errorf("Expected Host %v to Compare %t to %v", host, result, other)
		}
	}
}

func TestGetHostByName(t *testing.T) {
	c, err := NewClient(GetConfigFromEnv())
	if err != nil {
		t.Fatalf("failed to create client with error: %v", err)
	}

	nameLabel := accTestHost.NameLabel
	hosts, err := c.GetHostByName(nameLabel)
	if err != nil {
		t.Fatalf("failed to get host with error: %v", err)
	}

	host := hosts[0]
	if host.NameLabel != nameLabel {
		t.Errorf("expected host to have name `%s` received `%s` instead.", nameLabel, host.NameLabel)
	}

}

func TestGetSortedHosts(t *testing.T) {
	c, err := NewClient(GetConfigFromEnv())
	if err != nil {
		t.Fatalf("failed to create client with error: %v", err)
	}

	poolName := accTestHost.Pool
	hosts, err := c.GetSortedHosts(Host{Pool: poolName}, "id", "asc")
	if err != nil {
		t.Fatalf("failed to get host with error: %v", err)
	}
	if len(hosts) == 0 {
		t.Errorf("failed to find any host for pool `%s`.", poolName)
	}
	if len(hosts) > 1 {
		if hosts[0].Id > hosts[1].Id {
			t.Errorf("failed to sort hosts. expected %s to be smaller than %s.", hosts[0].Id, hosts[1].Id)
		}
	}
	for _, host := range hosts {
		if host.Pool != poolName {
			t.Errorf("expected pool to have name `%s` received `%s` instead.", poolName, host.Pool)
		}
	}
}

func Test_sortHostsByField(t *testing.T) {
	type args struct {
		hosts []Host
		by    string
		order string
	}
	tests := []struct {
		name string
		args args
		want []Host
	}{
		{name: "sortIdAsc", args: args{hosts: []Host{{Id: "acb"}, {Id: "abc"}}, order: "asc", by: "id"}, want: []Host{{Id: "abc"}, {Id: "acb"}}},
		{name: "sortIdDesc", args: args{hosts: []Host{{Id: "acb"}, {Id: "abc"}}, order: "desc", by: "id"}, want: []Host{{Id: "acb"}, {Id: "abc"}}},
		{name: "sortWrongOrder", args: args{hosts: []Host{{Id: "acb"}, {Id: "abc"}}, order: "TRASH", by: "id"}, want: []Host{{Id: "acb"}, {Id: "abc"}}},
		{name: "sortWrongField", args: args{hosts: []Host{{Id: "acb"}, {Id: "abc"}, {Id: "xyz"}}, order: "asc", by: "UNKNOWN"}, want: []Host{{Id: "acb"}, {Id: "abc"}, {Id: "xyz"}}},
		{name: "sortNameLabelAsc", args: args{hosts: []Host{{NameLabel: "acb"}, {NameLabel: "abc"}, {NameLabel: "xyz"}}, order: "asc", by: "name_label"}, want: []Host{{NameLabel: "abc"}, {NameLabel: "acb"}, {NameLabel: "xyz"}}},
		{name: "noSort", args: args{hosts: []Host{{Id: "1", NameLabel: "xyz"}, {Id: "2", NameLabel: "acb"}, {Id: "3", NameLabel: "abc"}}, order: "", by: ""}, want: []Host{{Id: "1", NameLabel: "xyz"}, {Id: "2", NameLabel: "acb"}, {Id: "3", NameLabel: "abc"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sortHostsByField(tt.args.hosts, tt.args.by, tt.args.order); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sortHostsByField() = %v, want %v", got, tt.want)
			}
		})
	}
}
