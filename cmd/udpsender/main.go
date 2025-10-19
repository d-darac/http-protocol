package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const lport = ":8080"
const rport = ":42069"

func main() {
	laddr, err := net.ResolveUDPAddr("udp", lport)
	if err != nil {
		log.Fatalf("couldn't resolve udp address: %s", err)
	}
	raddr, err := net.ResolveUDPAddr("udp", rport)
	if err != nil {
		log.Fatalf("couldn't resolve udp address: %s", err)
	}

	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		log.Fatalf("couldn't establish connection: %s", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		str, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("error: %s", err)
		}
		_, err = conn.Write([]byte(str))
		if err != nil {
			log.Printf("error: %s", err)
		}
	}
}
