package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path/filepath"
	"strings"
)

// MakeSSHKeyPair make a pair of public and private keys for SSH access.
// Public key is encoded in the format for inclusion in an OpenSSH authorized_keys file.
// Private Key generated is PEM encoded
func MakeSSHKeyPair() (key []byte, pub []byte, err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// encode private key as PEM
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

// todo: refactor below
func getSSHKeyFromSecret(ctx context.Context, err error) (string, error) {
	secret, err := c.CoreV1().Secrets(options.Namespace).Get(ctx, options.sshSecretName, v1.GetOptions{})
	if err != nil {
		return "", errors.Wrap(err, "error getting ssh secret")
	}

	b64key := secret.Data["ssh-privatekey"]
	var key []byte
	_, err = base64.StdEncoding.Decode(b64key, key)
	if err != nil {
		return "", errors.Wrap(err, "error decoding private key")
	}

	keyPath, err := writeSSHKeyToHost(b64key, err)
	if err != nil {
		return "", errors.Wrap(err, "error writing key to host")
	}
	return keyPath, nil
}

func writeSSHKeyToHost(key []byte, err error) (string, error) {

	mode := 0760
	err = os.MkdirAll(options.tanzuDir, os.FileMode(mode))
	if err != nil {
		return "", errors.Wrap(err, "err creating jumpbox dir")
	}

	keyPath := filepath.Join(options.tanzuDir, options.sshSecretName)
	file, err := os.Create(keyPath)
	if err != nil {
		return "", errors.Wrap(err, "error creating ssh key file")
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	mode = 0600
	err = file.Chmod(os.FileMode(mode))
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("error chmoding file: %s", keyPath))
	}
	_, err = file.Write(key)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("error writing file at: %s", keyPath))
	}

	return keyPath, nil
}
