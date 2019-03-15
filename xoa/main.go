package xoa

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/rpc"

	"github.com/ddelnano/terraform-provider-xenorchestra/common"
	"github.com/gorilla/websocket"
	"github.com/powerman/rpc-codec/jsonrpc2"
)

var cloudConfig = `
#cloud-config
packages:
 - make
 - build-essential
write_files:
-   content: |
        #!/bin/bash

        # Wrapper around puppet "ominbus" bundle
        # Example usage:
        #   run bundle install => puppet-bundle install
        #   run r10k install   => puppet-bundle exec r10k install

        export GEM_HOME=vendor/gem
        export GEM_PATH=$GEM_HOME:/opt/puppetlabs/puppet/lib/ruby/gems/2.4.0:$GEM_PATH
        export PATH=/opt/puppetlabs/puppet/bin:$PATH

        bundle $*
    path: /usr/local/bin/puppet-bundle
    permissions: '0774'
    owner: root:admin

-   content: |
        # This file can be used to override the default puppet settings.
        # See the following links for more details on what settings are available:
        # - https://puppet.com/docs/puppet/latest/config_important_settings.html
        # - https://puppet.com/docs/puppet/latest/config_about_settings.html
        # - https://puppet.com/docs/puppet/latest/config_file_main.html
        # - https://puppet.com/docs/puppet/latest/configuration.html
        environment = masterbranch
        environmentpath = /etc/puppetlabs/puppet/environments
        basemodulepath = /home/ddelnano/.puppetlabs/etc/code/modules:/opt/puppetlabs/puppet/modules:/etc/puppetlabs/puppet/environments/masterbranch/modules:/etc/puppetlabs/puppet/environments/masterbranch/vendor/modules

    path: /etc/puppetlabs/puppet/puppet.conf
runcmd:
 - mkdir -p /etc/puppetlabs/puppet/environments/masterbranch
 - GIT_SSH_COMMAND="ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no" git clone git@gitlab.com:ddelnano/puppet /etc/puppetlabs/puppet/environments/masterbranch
 - cd /etc/puppetlabs/puppet/environments/masterbranch
 - git checkout remove-unnecessary-junk-create-profile_server
 - /opt/puppetlabs/puppet/bin/gem install bundler
 - puppet-bundle install
 - make r10k
 - /opt/puppetlabs/bin/puppet apply --execute 'include profile_server' --test
`

func esting() {
	ws, res, err := dialer.Dial("ws://xoa.internal.ddelnano.com/api/", http.Header{})
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	defer ws.Close()

	log.Println(res)

	handle(ws)
}

var dialer = websocket.Dialer{
	ReadBufferSize:  common.MaxMessageSize,
	WriteBufferSize: common.MaxMessageSize,
}

type clientRequest struct {
	Version string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      *uint64     `json:"id,omitempty"`
}

type clientResponse struct {
	Email string `json:"email,omitempty"`
	Id    string `json:"id,omitempty"`
	// password string `json:"password"`
	// Version  string `json:"jsonrpc"`
	// // ID      *uint64          `json:"id"`
	// Result *json.RawMessage `json:"result,omitempty"`
	// Error  *jsonrpc2.Error  `json:"error,omitempty"`
}

type rwc struct {
	r io.Reader
	c *websocket.Conn
}

func (c *rwc) Write(p []byte) (int, error) {
	fmt.Println("Writing!")
	fmt.Println(string(p))
	err := c.c.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (c *rwc) Read(p []byte) (int, error) {
	for {
		if c.r == nil {
			// Advance to next message.
			var err error
			_, c.r, err = c.c.NextReader()
			if err != nil {
				return 0, err
			}
		}
		n, err := c.r.Read(p)
		if err == io.EOF {
			// At end of message.
			c.r = nil
			if n > 0 {
				return n, nil
			} else {
				// No data read, continue to next message.
				continue
			}
		}
		return n, err
	}
}

func (c *rwc) Close() error {
	return c.c.Close()
}

func handle(ws *websocket.Conn) {
	defer func() {
		ws.Close()
	}()

	codec := jsonrpc2.NewClientCodec(&rwc{c: ws})
	c := rpc.NewClientWithCodec(codec)

	params := map[string]interface{}{
		"email":    "admin@admin.net",
		"password": "admin",
	}
	var reply clientResponse
	err := c.Call("session.signInWithPassword", params, &reply)
	if err != nil {
		log.Printf("%v", err)
	}

	params = map[string]interface{}{
		"bootAfterCreate":  true,
		"name_label":       "Hello from terraform!",
		"name_description": "Hello from golang",
		"template":         "2dd0373e-0ed5-7413-a57f-1958d03b698c",
		"cloudConfig":      cloudConfig,
		"coreOs":           false,
		"cpuCap":           nil,
		"cpuWeight":        nil,
		"CPUs":             1,
		"memoryMax":        1073741824,
		"existingDisks": map[string]interface{}{
			"0": map[string]interface{}{
				"$SR":              "7f469400-4a2b-5624-cf62-61e522e50ea1",
				"name_description": "Created by XO",
				"name_label":       "Ubuntu Bionic Beaver 18.04_imavo",
				"size":             32212254720,
			},
			"1": map[string]interface{}{
				"$SR":              "7f469400-4a2b-5624-cf62-61e522e50ea1",
				"name_description": "",
				"name_label":       "XO CloudConfigDrive",
				"size":             10485760,
			},
			"2": map[string]interface{}{
				"$SR":              "7f469400-4a2b-5624-cf62-61e522e50ea1",
				"name_description": "",
				"name_label":       "XO CloudConfigDrive",
				"size":             10485760,
			},
		},
		"VIFs": []interface{}{
			map[string]string{
				"network": "d225cf00-36f8-e6d6-6a29-02636d4de56b",
			},
		},
	}
	err = c.Call("vm.create", params, &reply)
	if err != nil {
		log.Printf("%v", err)
	}
	log.Printf("%v", reply)
}

// func (r *clientResponse) reset() {
// 	r.Version = ""
// 	r.Result = nil
// 	r.Error = nil
// }

func (r *clientResponse) UnmarshalJSON(raw []byte) error {
	log.Printf("raw bytes: %s", raw)
	type resp *clientResponse
	if err := json.Unmarshal(raw, resp(r)); err != nil {
		return errors.New("bad response: " + string(raw))
	}

	return nil
}
