package tls

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadOrGenerate_Off(t *testing.T) {
	cfg := Config{Mode: "off"}
	tc, err := LoadOrGenerate(cfg)
	require.NoError(t, err)
	assert.Nil(t, tc)
}

func TestLoadOrGenerate_Empty(t *testing.T) {
	cfg := Config{Mode: ""}
	tc, err := LoadOrGenerate(cfg)
	require.NoError(t, err)
	assert.Nil(t, tc)
}

func TestLoadOrGenerate_Unknown(t *testing.T) {
	cfg := Config{Mode: "invalid"}
	_, err := LoadOrGenerate(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown mode")
}

func TestLoadOrGenerate_Auto(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{Mode: "auto", CertDir: dir}

	tc, err := LoadOrGenerate(cfg)
	require.NoError(t, err)
	require.NotNil(t, tc)
	assert.Len(t, tc.Certificates, 1)
	assert.Equal(t, uint16(tls.VersionTLS12), tc.MinVersion)

	// Verify cert and key files were created
	assert.FileExists(t, filepath.Join(dir, "nucleus.crt"))
	assert.FileExists(t, filepath.Join(dir, "nucleus.key"))

	// Verify key file permissions (owner-only)
	info, err := os.Stat(filepath.Join(dir, "nucleus.key"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	// Parse the certificate
	leaf, err := x509.ParseCertificate(tc.Certificates[0].Certificate[0])
	require.NoError(t, err)
	assert.Equal(t, "nucleus-node", leaf.Subject.CommonName)
	assert.Contains(t, leaf.DNSNames, "localhost")
	assert.Equal(t, x509.Ed25519, leaf.PublicKeyAlgorithm)
}

func TestLoadOrGenerate_Auto_Reuse(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{Mode: "auto", CertDir: dir}

	// Generate first time
	tc1, err := LoadOrGenerate(cfg)
	require.NoError(t, err)

	// Load again — should reuse
	tc2, err := LoadOrGenerate(cfg)
	require.NoError(t, err)

	// Same certificate
	assert.Equal(t, tc1.Certificates[0].Certificate[0], tc2.Certificates[0].Certificate[0])
}

func TestLoadOrGenerate_Provided(t *testing.T) {
	// First generate a cert to use as "provided"
	dir := t.TempDir()
	autoCfg := Config{Mode: "auto", CertDir: dir}
	_, err := LoadOrGenerate(autoCfg)
	require.NoError(t, err)

	// Now load as "provided"
	providedCfg := Config{
		Mode:     "provided",
		CertFile: filepath.Join(dir, "nucleus.crt"),
		KeyFile:  filepath.Join(dir, "nucleus.key"),
	}
	tc, err := LoadOrGenerate(providedCfg)
	require.NoError(t, err)
	require.NotNil(t, tc)
	assert.Len(t, tc.Certificates, 1)
}

func TestLoadOrGenerate_Provided_MissingFiles(t *testing.T) {
	t.Run("empty paths", func(t *testing.T) {
		cfg := Config{Mode: "provided"}
		_, err := LoadOrGenerate(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "requires cert_file")
	})

	t.Run("nonexistent files", func(t *testing.T) {
		cfg := Config{
			Mode:     "provided",
			CertFile: "/nonexistent/cert.pem",
			KeyFile:  "/nonexistent/key.pem",
		}
		_, err := LoadOrGenerate(cfg)
		assert.Error(t, err)
	})
}

func TestGenerateSelfSigned(t *testing.T) {
	tc, err := GenerateSelfSigned()
	require.NoError(t, err)
	require.NotNil(t, tc)
	assert.Len(t, tc.Certificates, 1)

	leaf, err := x509.ParseCertificate(tc.Certificates[0].Certificate[0])
	require.NoError(t, err)
	assert.Equal(t, "nucleus-test", leaf.Subject.CommonName)
	assert.Equal(t, x509.Ed25519, leaf.PublicKeyAlgorithm)
}
