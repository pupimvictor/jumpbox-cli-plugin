package main

import (
	"bytes"
	"encoding/base64"
	"github.com/pkg/errors"
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
		SshPublicKey     string
		SshPrivateKey    string
		User             string

		pvcName           string
		configName        string
		svcName           string
		sshSecretName     string
		sshPrivateKeyPath string
		tanzuDir          string
	}
)

var userdata = `
#cloud-config

ssh_authorized_keys:
  - {{ .SshPublicKey }} 

fs_setup:
  - label: workspace
    filesystem: 'ext4'
    device: '/dev/sdb'
    partition: 'auto'

mounts:
  - [ sdb, /workspace ]

runcmd:
  - sudo chmod 777 /workspace
`

func setup(args []string) {
	vmName := args[0]
	options.Name = vmName
	options.pvcName = vmName + "-pvc"
	options.sshSecretName = vmName + "-ssh"
	options.configName = vmName + "-cm"
	options.svcName = vmName + "-svc"
}

func buildUserdata() error {
	sshPrivateKey, sshPubKey, err := MakeSSHKeyPair()
	if err != nil {
		return errors.WithMessage(err, "err creating ssh key pair")
	}
	options.SshPublicKey = string(sshPubKey)
	options.SshPrivateKey = string(sshPrivateKey)

	t := template.Must(template.New("userdata").Parse(userdata))

	buf := new(bytes.Buffer)
	err = t.Execute(buf, options)
	if err != nil {
		return errors.WithMessage(err, "err building userdata template")
	}
	options.UserData = base64.StdEncoding.EncodeToString(buf.Bytes())

	return nil
}
