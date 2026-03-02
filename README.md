# 证书管理

# 使用方式

```bash
[dev@dev dest]$ ./cert -h
Usage of cert:
  -d string
        Domain name (default "example.com")
  -host string
        Hostname (default "localhost")
  -i string
        IP address (default "127.0.0.1")
  -k int
        RSA key size (2048, 3072, or 4096) (default 4096)
  -o string
        Output directory for certificate files (default ".")
  -t int
        Validity period in days (default 3650)
```

生成证书示例：

```bash
./dest/cert -d example.com -i 127.0.0.1 -host localhost -o dest
```

输出文件：

- `{domain}.cert.pem`
- `{domain}.key.pem`
