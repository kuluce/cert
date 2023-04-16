package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/toolkits/logger"
)

var (
	days   *int    // 有效期
	domain *string // 域名
	ip     *string // ip
	h      *string // a
)

func main() {

	// logger.Info("begin")
	// logger.Info("end")
	days = flag.Int("t", 3650, "有效期天数，默认3650天")
	domain = flag.String("d", "example.com", "DOMAIN")
	ip = flag.String("i", "127.0.0.1", "IP")
	h = flag.String("h", "localhost", "IP")

	fmt.Printf("len(os.Args): %v\n", len(os.Args))

	flag.Parse()

	if *days <= 0 {
		logger.Errorln("有效期不能小于零")
		os.Exit(1)
	}
	logger.Info("有效期:%v天", *days)

	if *domain == "" {
		logger.Errorln("domain can not be empty")
		os.Exit(1)
	}
	logger.Info("域名:%v", *domain)

	if *ip == "" {
		logger.Errorln("ip can not be empty")
		os.Exit(1)
	}
	logger.Info("IP:%v", *ip)

	if *h == "" {
		logger.Errorln("host can not be empty")
		os.Exit(1)
	}
	logger.Info("主机:%v", *h)

	createCert(*days, *domain, []string{*ip}, []string{*h})
}

func createCert(days int, domain string, ipList []string, aList []string) {
	// 生成私钥
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatalf("failed to generate private key: %s", err)
	}

	// 创建证书模板
	now := time.Now()
	periodOfValidity := now.Add(time.Duration(days) * 24 * time.Hour)
	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: domain},
		NotBefore:             now,
		NotAfter:              periodOfValidity,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// 添加IP地址和DNS名到证书模板
	ips := []net.IP{}
	for _, ip := range ipList {
		ips = append(ips, net.ParseIP(ip))
	}
	template.IPAddresses = ips
	template.DNSNames = aList

	// 生成证书
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
	}

	// 保存证书到文件
	certFileName := fmt.Sprintf("%s.cert.pem", domain)
	certOut, err := os.Create(certFileName)
	if err != nil {
		log.Fatalf("failed to open cert.pem for writing: %s", err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()
	log.Print("written cert.pem\n")

	// 保存私钥到文件
	privFileName := fmt.Sprintf("%s.key.pem", domain)
	keyOut, err := os.OpenFile(privFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("failed to open key.pem for writing: %s", err)
	}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyOut.Close()
	log.Print("written key.pem\n")

}
