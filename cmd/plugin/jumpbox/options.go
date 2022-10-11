package main

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"github.com/pkg/errors"
	"os"
	"text/template"
)

type (
	VMOptions struct {
		Name             string
		Namespace        string
		UserData         string
		StorageClassName string
		ImageName        string
		ClassName        string
		NetworkType      string
		NetworkName      string
		SshPubPath       string
		SshKeyPath       string
		User             string
		Password         string
		WaitCreate       bool

		pvcName    string
		configName string
		svcName    string
		SshPubKey  string
	}
)

func setup(args []string) {
	vmName := args[0]
	options.Name = vmName
	options.pvcName = vmName + "-pvc"
	options.configName = vmName + "-cm"
	options.svcName = vmName + "-svc"
}

func buildUserdata() error {
	hash := md5.Sum([]byte(options.Password))
	options.Password = hex.EncodeToString(hash[:])

	sshPubKey, err := os.ReadFile(os.ExpandEnv(options.SshPubPath))
	if err != nil {
		return errors.WithMessage(err, "err reading ssh pub key")
	}
	options.SshPubKey = string(sshPubKey)

	t := template.Must(template.New("userdata").Parse(userdata))

	buf := new(bytes.Buffer)
	err = t.Execute(buf, options)
	if err != nil {
		return errors.WithMessage(err, "err building userdata template")
	}
	options.UserData = base64.StdEncoding.EncodeToString(buf.Bytes())

	return nil
}

var userdata = `
#cloud-config
## Required syntax at the start of user-data file
users:
## Create the default user for the OS
  - default

  - name: {{ .User }}
    sudo: ALL=(ALL) NOPASSWD:ALL
    groups: users, admins
    shell: /bin/bash
    passwd: {{ .Password }}
    ssh_authorized_keys:
      - {{ .SshPubKey }}


## Enable DHCP on the default network interface provisioned in the VM
network:
  version: 2
  ethernets:
      ens192:
          dhcp4: true

## Setup Filesystem and Mount PV disk
fs_setup:
  - label: workspace
    filesystem: 'ext4'
    device: '/dev/sdb'
    partition: 'auto'

mounts:
 - [ sdb, /workspace ]

apt_upgrade: true
packages:
    - traceroute
    - unzip
    - tree
    - jq

runcmd:
  - chmod 774 /workspace
  - chown -R root:admins /workspace
`
