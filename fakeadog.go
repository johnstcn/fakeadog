package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/johnstcn/fakeadog/parser"
)

func main() {
	var host string
	var port int

	flag.StringVar(&host, "host", "localhost", "address to bind to, default is localhost")
	flag.IntVar(&port, "port", 8125, "port to bind to, default is 8125")
	flag.Parse()

	if envHost := os.Getenv("HOST"); envHost != "" {
		host = envHost
	}

	if envPort := os.Getenv("PORT"); envPort != "" {
		envPortI, err := strconv.Atoi(envPort)
		if err != nil {
			fmt.Printf("PORT was not set to valid int: %q", envPort)
			os.Exit(1)
		}
		port = envPortI
	}

	hostport := fmt.Sprintf("%s:%d", host, port)
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
		if err != nil {
			log.Println("[ERROR] reading from udp:", err)
			continue
		}
		payload := buf[:n]
		// payload may contain multiple metrics separated by newlines
		splitted := bytes.Split(payload, []byte("\n"))
		for _, sp := range splitted {
			m, err := p.Parse(sp)
			if err != nil {
				log.Println("[ERROR] parsing payload:", string(payload))
				log.Println("[ERROR]", err)
				continue
			}
			fmt.Println("[INFO]", m)
		}
	}
}
