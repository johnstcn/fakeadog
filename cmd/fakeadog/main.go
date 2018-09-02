package main

import (
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/johnstcn/fakeadog/pkg/parser"

	"github.com/sirupsen/logrus"
)

func main() {
	var log = logrus.New()
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
			log.Fatalf("PORT was not set to valid int: %q", envPort)
		}
		port = envPortI
	}

	hostport := fmt.Sprintf("%s:%d", host, port)
	addr, err := net.ResolveUDPAddr("udp", hostport)
	if err != nil {
		log.Fatalf("could not resolve address %s: %s\n", hostport, err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("could not listen on %s: %s\n", hostport, err)
	}

	log.Info("listening on ", hostport)
	// from datadog-go/statsd
	buf := make([]byte, 65467)
	p := parser.NewDatadogParser()

	defer conn.Close()
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Error("reading from udp:", err)
			continue
		}
		payload := buf[:n]
		// payload may contain multiple metrics separated by newlines
		splitted := bytes.Split(payload, []byte("\n"))
		for _, sp := range splitted {
			m, err := p.Parse(sp)
			if err != nil {
				log.Errorf("parsing payload %q: %s", string(payload), err)
				continue
			}
			log.WithFields(logrus.Fields{
				"type":  m.Type,
				"name":  m.Name,
				"value": m.Value,
				"tags":  m.Tags,
			}).Info("received datadog metric")
		}
	}
}
