package main

import (
	"bytes"
	"encoding/pem"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseFlagsHelpUsesDefaultHelpFlag(t *testing.T) {
	t.Parallel()

	_, err := parseFlags([]string{"-h"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected help error")
	}
	if err != flag.ErrHelp {
		t.Fatalf("expected flag.ErrHelp, got %v", err)
	}
}

func TestBuildCertificateTemplate(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 2, 12, 0, 0, 0, time.UTC)
	cfg := Config{
		Days:     30,
		Domain:   "example.com",
		IP:       "127.0.0.1",
		Hostname: "localhost",
	}

	template, err := buildCertificateTemplate(cfg, now)
	if err != nil {
		t.Fatalf("build certificate template: %v", err)
	}
	if template.SerialNumber == nil || template.SerialNumber.Sign() <= 0 {
		t.Fatalf("expected positive serial number, got %v", template.SerialNumber)
	}
	if got, want := template.NotAfter, now.Add(30*24*time.Hour); !got.Equal(want) {
		t.Fatalf("unexpected expiry: got %s want %s", got, want)
	}
	if len(template.IPAddresses) != 1 || template.IPAddresses[0].String() != "127.0.0.1" {
		t.Fatalf("unexpected IP SANs: %#v", template.IPAddresses)
	}
	if len(template.DNSNames) != 2 || template.DNSNames[0] != "localhost" || template.DNSNames[1] != "example.com" {
		t.Fatalf("unexpected DNS SANs: %#v", template.DNSNames)
	}
}

func TestWritePEMFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "cert.pem")
	if err := writePEMFile(path, "CERTIFICATE", []byte("payload"), 0644); err != nil {
		t.Fatalf("write PEM file: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read PEM file: %v", err)
	}
	block, rest := pem.Decode(data)
	if block == nil {
		t.Fatal("expected PEM block")
	}
	if len(rest) != 0 {
		t.Fatalf("unexpected trailing data: %q", string(rest))
	}
	if block.Type != "CERTIFICATE" {
		t.Fatalf("unexpected block type: %s", block.Type)
	}
	if string(block.Bytes) != "payload" {
		t.Fatalf("unexpected block payload: %q", string(block.Bytes))
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat PEM file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0644 {
		t.Fatalf("unexpected file mode: %o", got)
	}
}

func TestValidateConfigRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Days:     1,
		Domain:   "example.com",
		IP:       "127.0.0.1",
		Hostname: "localhost",
		KeySize:  1024,
	}

	if err := validateConfig(cfg); err == nil {
		t.Fatal("expected validation error")
	}
}
