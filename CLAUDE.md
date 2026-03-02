# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a certificate management tool (证书管理) - a CLI utility that generates self-signed SSL/TLS certificates. It creates RSA 4096-bit keys with configurable domain, IP, and validity period.

## Build Commands

Build the binary (includes UPX compression):
```bash
make build
```

Build and run:
```bash
make run
```

Clean build artifacts:
```bash
make clean
```

## Usage

The compiled binary is output to `dest/cert`. Run with flags:

```bash
./dest/cert -d example.com -i 127.0.0.1 -h localhost -t 3650
```

Flags:
- `-d string` - Domain name (default: "example.com")
- `-i string` - IP address (default: "127.0.0.1")
- `-h string` - Hostname (default: "localhost")
- `-t int` - Validity period in days (default: 3650)

Output files (in working directory):
- `{domain}.cert.pem` - The certificate
- `{domain}.key.pem` - The private key

## Architecture

The project follows a minimal Go CLI structure:

- `cmd/mgr/main.go` - Single-file application containing all logic
- `go.mod` - Module definition (Go 1.20)
- `makefile` - Build automation with UPX compression
- `dest/` - Build output directory

The certificate generation process:
1. Generate RSA 4096-bit private key
2. Create X.509 certificate template with configured domain/IP
3. Add Subject Alternative Names (IP addresses and DNS names)
4. Self-sign the certificate
5. Write certificate and key to PEM files

## Dependencies

- `github.com/toolkits/logger` - Simple logging utility

## Notes

- The build process uses UPX for binary compression (level 9)
- Cross-compilation for Windows (amd64) is included in `make build`
- The `make push` command handles SSH agent setup for git push
