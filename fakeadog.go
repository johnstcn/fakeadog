package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/johnstcn/fakeadog/parser"
)

func main() {
	host := flag.String("host", "localhost", "address to bind to, default is localhost")
	port := flag.Int("port", 8125, "port to bind to, default is 8125")
	flag.Parse()

	hostport := fmt.Sprintf("%s:%d", *host, *port)
	addr, err := net.ResolveUDPAddr("udp", hostport)
	if err != nil {
		fmt.Printf("could not resolve address %s: %s\n", hostport, err)
		os.Exit(1)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Printf("could not listen on %s: %s\n", hostport, err)
		os.Exit(1)
	}

	log.Println("[INFO] listening on", hostport)
	// from datadog-go/statsd
	buf := make([]byte, 65467)
	p := parser.NewDatadogParser()

	defer conn.Close()
	for {
		n, _, err := conn.ReadFromUDP(buf)
		payload := buf[:n]
		m, err := p.Parse(payload)
		if err != nil {
			log.Println("[ERROR] parsing payload:", string(payload))
			log.Println("[ERROR]", err)
			continue
		}
		fmt.Println("[INFO]", m)
	}
}
