package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"strings"
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

// MakeSSHKeyPair make a pair of public and private keys for SSH access.
// Public key is encoded in the format for inclusion in an OpenSSH authorized_keys file.
// Private Key generated is PEM encoded
func MakeSSHKeyPair() (key []byte, pub []byte, err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	//encode private key as PEM
	privateKeyStr := new(strings.Builder)
	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	if err := pem.Encode(privateKeyStr, privateKeyPEM); err != nil {
		return nil, nil, err
	}

	// generate and write public key
	pubKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	key = []byte(privateKeyStr.String())
	pub = ssh.MarshalAuthorizedKey(pubKey)
	return key, pub, nil
}
