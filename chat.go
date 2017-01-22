package main

import (
	"flag"
	"net"
	"log"
	"fmt"
	"io"
	"os"
	"bufio"
)

const (
	port = ":8080"
)

type User struct {
	username string
	ip net.IP
	server bool // this may be unnecessary
}

func HandleAsServer(conn net.Conn) {
	fmt.Println("Client has successfully connected")
	_, err := io.WriteString(conn, "You have connected to the server\n")
	if err != nil {
		log.Println(err)
		return
	}
	done := make(chan struct{})
	go HandleIncoming(conn, done)
	go HandleOutgoing(conn, done)
	<-done
	fmt.Println("Client has been disconnected")
}

func HandleIncoming(conn net.Conn, done chan <- struct{}) {
	defer conn.Close()
	input := bufio.NewScanner(conn)
	for input.Scan() {
		fmt.Println(input.Text())
	}
	done <- struct{}{}
}

func HandleOutgoing(conn net.Conn, done chan <- struct{}) {
	defer conn.Close()
	io.Copy(conn, os.Stdin)
	done <- struct{}{}
}

func main() {
	listen := flag.Bool("listen", false, "set to true when instance acts as server")
	// username := flag.String("username", "anonymous", "provides the other part with your identity")
	flag.Parse()
	connIP := net.ParseIP(flag.Arg(0)).String()
	if connIP == "" {
		log.Fatal("invalid IP address given as argument")
	}

	fmt.Println(connIP)

	if *listen {
		l, err := net.Listen("tcp", connIP+port)
		if err != nil {
			log.Fatal(err)
		}
		for {
			c, err := l.Accept()
			if err != nil {
				log.Println("Could not connect")
			}
			go HandleAsServer(c)
		}
	} else {
		conn, err := net.Dial("tcp", connIP+port)
		if err != nil {
			log.Fatal(err)
		}
		done := make(chan struct{})
		go HandleIncoming(conn, done)
		go HandleOutgoing(conn, done)
		<-done
		fmt.Println("You are now disconnected.  Goodbye!")
	}
}