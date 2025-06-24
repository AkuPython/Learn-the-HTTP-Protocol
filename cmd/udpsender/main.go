package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

const host = "localhost:42069"

func main() {
	s, err := net.ResolveUDPAddr("udp4", host)
	if err != nil {
		fmt.Println("Could not resolve UDP:", host)
	}

	c, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		fmt.Println("Could Dial UDP:", host)
		os.Exit(1)
	}
	defer c.Close()

	fmt.Printf("Sending to %s.\n", host)

	r := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		inp, err := r.ReadString('\n')
		if err != nil {
			fmt.Println("Could read input:", err)
			os.Exit(1)
		}
		_, err = c.Write([]byte(inp))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could send message: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Message sent: %s", inp)
	}

}
