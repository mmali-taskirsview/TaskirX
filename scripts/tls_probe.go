package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

type hostEntry struct {
	name string
	host string
}

func probeTLS13(entry hostEntry) error {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	config := &tls.Config{
		ServerName: entry.host,
		MinVersion: tls.VersionTLS13,
		MaxVersion: tls.VersionTLS13,
	}
	conn, err := tls.DialWithDialer(dialer, "tcp", fmt.Sprintf("%s:443", entry.host), config)
	if err != nil {
		return err
	}
	defer conn.Close()

	state := conn.ConnectionState()
	cert := state.PeerCertificates

	fmt.Printf("%s: OK (TLS %x)\n", entry.name, state.Version)
	fmt.Printf("  Host     : %s\n", entry.host)
	if len(cert) > 0 {
		fmt.Printf("  Subject  : %s\n", cert[0].Subject.CommonName)
		fmt.Printf("  Issuer   : %s\n", cert[0].Issuer.CommonName)
		fmt.Printf("  NotBefore: %s\n", cert[0].NotBefore.Format(time.RFC3339))
		fmt.Printf("  NotAfter : %s\n", cert[0].NotAfter.Format(time.RFC3339))
	} else {
		fmt.Printf("  Cert     : none\n")
	}
	return nil
}

func main() {
	hosts := []hostEntry{
		{name: "API Backend", host: "api.taskir.com"},
		{name: "Dashboard", host: "dashboard.taskir.com"},
		{name: "Bidding", host: "bidding.taskir.com"},
	}

	fmt.Println("TLS 1.3 probe for TaskirX endpoints")
	fmt.Println("======================================")

	for _, entry := range hosts {
		if err := probeTLS13(entry); err != nil {
			fmt.Printf("%s: FAIL\n", entry.name)
			fmt.Printf("  Host : %s\n", entry.host)
			fmt.Printf("  Error: %s\n\n", err.Error())
			continue
		}
		fmt.Println()
	}
}
