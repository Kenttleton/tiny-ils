package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// LoadOrCreate loads the node keypair from disk, generating it if it doesn't exist.
func LoadOrCreate(privKeyPath, pubKeyPath string) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	if _, err := os.Stat(privKeyPath); os.IsNotExist(err) {
		return generate(privKeyPath, pubKeyPath)
	}
	return load(privKeyPath, pubKeyPath)
}

func generate(privKeyPath, pubKeyPath string) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate keypair: %w", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal private key: %w", err)
	}
	if err := writePEM(privKeyPath, "PRIVATE KEY", privBytes); err != nil {
		return nil, nil, err
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal public key: %w", err)
	}
	if err := writePEM(pubKeyPath, "PUBLIC KEY", pubBytes); err != nil {
		return nil, nil, err
	}

	return priv, pub, nil
}

func load(privKeyPath, pubKeyPath string) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	privPEM, err := os.ReadFile(privKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read private key: %w", err)
	}
	block, _ := pem.Decode(privPEM)
	if block == nil {
		return nil, nil, fmt.Errorf("decode private key PEM")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse private key: %w", err)
	}
	priv, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, nil, fmt.Errorf("unexpected key type")
	}
	return priv, priv.Public().(ed25519.PublicKey), nil
}

func writePEM(path, pemType string, data []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	return pem.Encode(f, &pem.Block{Type: pemType, Bytes: data})
}
