// Package tls provides TLS certificate management for Open Nucleus.
//
// Supports three modes:
//   - "auto": generates a self-signed Ed25519 certificate (development/field deployment)
//   - "provided": loads user-supplied cert and key files
//   - "off": no TLS (plain HTTP)
package tls

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// Config controls TLS behavior.
type Config struct {
	Mode     string `koanf:"mode"`      // "auto", "provided", "off"
	CertFile string `koanf:"cert_file"` // path to PEM certificate (mode=provided)
	KeyFile  string `koanf:"key_file"`  // path to PEM private key (mode=provided)
	CertDir  string `koanf:"cert_dir"`  // directory for auto-generated certs (mode=auto)
}

// LoadOrGenerate returns a *tls.Config based on the configuration mode.
// Returns nil for mode "off".
func LoadOrGenerate(cfg Config) (*tls.Config, error) {
	switch cfg.Mode {
	case "off", "":
		return nil, nil

	case "provided":
		return loadProvided(cfg.CertFile, cfg.KeyFile)

	case "auto":
		return autoGenerate(cfg.CertDir)

	default:
		return nil, fmt.Errorf("tls: unknown mode %q (use auto, provided, or off)", cfg.Mode)
	}
}

// loadProvided loads a TLS certificate and key from files.
func loadProvided(certFile, keyFile string) (*tls.Config, error) {
	if certFile == "" || keyFile == "" {
		return nil, fmt.Errorf("tls: mode=provided requires cert_file and key_file")
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("tls: loading cert/key: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// autoGenerate creates a self-signed Ed25519 certificate, storing it in certDir.
// Reuses existing cert if found and not expired.
func autoGenerate(certDir string) (*tls.Config, error) {
	if certDir == "" {
		certDir = "."
	}
	if err := os.MkdirAll(certDir, 0700); err != nil {
		return nil, fmt.Errorf("tls: creating cert dir: %w", err)
	}

	certPath := filepath.Join(certDir, "nucleus.crt")
	keyPath := filepath.Join(certDir, "nucleus.key")

	// Try loading existing cert
	if cert, err := tls.LoadX509KeyPair(certPath, keyPath); err == nil {
		leaf, parseErr := x509.ParseCertificate(cert.Certificate[0])
		if parseErr == nil && time.Now().Before(leaf.NotAfter) {
			return &tls.Config{
				Certificates: []tls.Certificate{cert},
				MinVersion:   tls.VersionTLS12,
			}, nil
		}
	}

	// Generate new self-signed cert
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("tls: generating Ed25519 key: %w", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("tls: generating serial number: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Open Nucleus"},
			CommonName:   "nucleus-node",
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 year
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost", "nucleus-node"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, pub, priv)
	if err != nil {
		return nil, fmt.Errorf("tls: creating certificate: %w", err)
	}

	// Write cert PEM
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		return nil, fmt.Errorf("tls: writing cert: %w", err)
	}

	// Write key PEM
	keyDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("tls: marshaling private key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		return nil, fmt.Errorf("tls: writing key: %w", err)
	}

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("tls: loading generated cert: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// GenerateSelfSigned creates a self-signed Ed25519 cert in memory (no disk).
// Useful for testing.
func GenerateSelfSigned() (*tls.Config, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      pkix.Name{CommonName: "nucleus-test"},
		NotBefore:    time.Now().Add(-1 * time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, pub, priv)
	if err != nil {
		return nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, _ := x509.MarshalPKCS8PrivateKey(priv)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}
