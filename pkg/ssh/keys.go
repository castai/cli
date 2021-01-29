package ssh

import (
	"bytes"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/ssh"
)

type Keys struct {
	Public  []byte
	Private []byte
}

func GenerateKeys(name string) (*Keys, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generating ssh keys: %w", err)
	}
	pubEncoded, err := encodePublicKey(pub, name)
	if err != nil {
		return nil, fmt.Errorf("encoding ssh public key: %w", err)
	}
	privEncoded, err := encodePrivateKey(priv)
	if err != nil {
		return nil, err
	}
	return &Keys{
		Public:  pubEncoded,
		Private: privEncoded,
	}, nil
}

func encodePublicKey(k ed25519.PublicKey, name string) ([]byte, error) {
	key, err := ssh.NewPublicKey(k)
	if err != nil {
		return nil, err
	}
	b := &bytes.Buffer{}
	b.WriteString(key.Type())
	b.WriteByte(' ')
	e := base64.NewEncoder(base64.StdEncoding, b)
	e.Write(key.Marshal())
	e.Close()
	b.WriteByte(' ')
	b.WriteString(name)
	b.WriteByte('\n')
	return b.Bytes(), nil
}

func encodePrivateKey(k ed25519.PrivateKey) ([]byte, error) {
	key, err := x509.MarshalPKCS8PrivateKey(k)
	if err != nil {
		return nil, err
	}
	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: key,
	}
	buf := &bytes.Buffer{}
	if err := pem.Encode(buf, block); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
