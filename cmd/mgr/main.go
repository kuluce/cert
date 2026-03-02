package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// Config holds command line configuration
type Config struct {
	Days      int
	Domain    string
	IP        string
	Hostname  string
	OutputDir string
	KeySize   int
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		log.Fatalf("failed to create certificate: %v", err)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	cfg, err := parseFlags(args, stdout, stderr)
	if err != nil {
		return err
	}
	if err := validateConfig(cfg); err != nil {
		return err
	}
	if err := ensureOutputDir(cfg.OutputDir); err != nil {
		return err
	}

	printConfig(stdout, cfg)
	return createCert(cfg)
}

func parseFlags(args []string, stdout, stderr io.Writer) (Config, error) {
	var cfg Config
	flags := flag.NewFlagSet("cert", flag.ContinueOnError)
	if helpRequested(args) {
		flags.SetOutput(stdout)
	} else {
		flags.SetOutput(stderr)
	}
	flags.IntVar(&cfg.Days, "t", 3650, "Validity period in days")
	flags.StringVar(&cfg.Domain, "d", "example.com", "Domain name")
	flags.StringVar(&cfg.IP, "i", "127.0.0.1", "IP address")
	flags.StringVar(&cfg.Hostname, "host", "localhost", "Hostname")
	flags.StringVar(&cfg.OutputDir, "o", ".", "Output directory for certificate files")
	flags.IntVar(&cfg.KeySize, "k", 4096, "RSA key size (2048, 3072, or 4096)")
	if err := flags.Parse(args); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func helpRequested(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "-help" || arg == "--help" {
			return true
		}
	}
	return false
}

func validateConfig(cfg Config) error {
	if cfg.Days <= 0 {
		return fmt.Errorf("validity period must be greater than 0")
	}
	if cfg.Domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}
	if cfg.IP == "" {
		return fmt.Errorf("IP cannot be empty")
	}
	if cfg.Hostname == "" {
		return fmt.Errorf("hostname cannot be empty")
	}
	if cfg.KeySize != 2048 && cfg.KeySize != 3072 && cfg.KeySize != 4096 {
		return fmt.Errorf("key size must be 2048, 3072, or 4096")
	}
	return nil
}

func ensureOutputDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("create output directory %s: %w", path, err)
	}
	return nil
}

func printConfig(w io.Writer, cfg Config) {
	fmt.Fprintf(w, "Validity: %d days\n", cfg.Days)
	fmt.Fprintf(w, "Domain: %s\n", cfg.Domain)
	fmt.Fprintf(w, "IP: %s\n", cfg.IP)
	fmt.Fprintf(w, "Hostname: %s\n", cfg.Hostname)
	fmt.Fprintf(w, "Output: %s\n", cfg.OutputDir)
	fmt.Fprintf(w, "Key Size: %d bits\n", cfg.KeySize)
}

func createCert(cfg Config) error {
	priv, err := rsa.GenerateKey(rand.Reader, cfg.KeySize)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	template, err := buildCertificateTemplate(cfg, time.Now())
	if err != nil {
		return err
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	certPath := filepath.Join(cfg.OutputDir, cfg.Domain+".cert.pem")
	if err := writePEMFile(certPath, "CERTIFICATE", derBytes, 0644); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}
	log.Printf("Certificate written to: %s", certPath)

	keyPath := filepath.Join(cfg.OutputDir, cfg.Domain+".key.pem")
	keyBytes := x509.MarshalPKCS1PrivateKey(priv)
	if err := writePEMFile(keyPath, "RSA PRIVATE KEY", keyBytes, 0600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}
	log.Printf("Private key written to: %s", keyPath)

	return nil
}

func buildCertificateTemplate(cfg Config, now time.Time) (*x509.Certificate, error) {
	serialNumber, err := randomSerialNumber()
	if err != nil {
		return nil, fmt.Errorf("generate certificate serial number: %w", err)
	}

	parsedIP := net.ParseIP(cfg.IP)
	if parsedIP == nil {
		return nil, fmt.Errorf("invalid IP address: %s", cfg.IP)
	}

	return &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: cfg.Domain,
		},
		NotBefore:             now,
		NotAfter:              now.Add(time.Duration(cfg.Days) * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{parsedIP},
		DNSNames:              uniqueStrings(cfg.Hostname, cfg.Domain),
	}, nil
}

func randomSerialNumber() (*big.Int, error) {
	limit := new(big.Int).Lsh(big.NewInt(1), 128)
	return rand.Int(rand.Reader, limit)
}

func uniqueStrings(values ...string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func writePEMFile(path, blockType string, bytes []byte, perm os.FileMode) (err error) {
	file, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file for %s: %w", path, err)
	}

	tmpPath := file.Name()
	defer func() {
		if err != nil {
			_ = os.Remove(tmpPath)
		}
	}()

	if err = file.Chmod(perm); err != nil {
		_ = file.Close()
		return fmt.Errorf("set permissions on %s: %w", tmpPath, err)
	}
	if err = pem.Encode(file, &pem.Block{Type: blockType, Bytes: bytes}); err != nil {
		_ = file.Close()
		return fmt.Errorf("encode PEM for %s: %w", path, err)
	}
	if err = file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync %s: %w", tmpPath, err)
	}
	if err = file.Close(); err != nil {
		return fmt.Errorf("close %s: %w", tmpPath, err)
	}
	if err = os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}
	return nil
}
