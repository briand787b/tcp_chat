package main

import (
	"flag"
	"net"
	"log"
	"fmt"
	"io"
	"os"
	"bufio"
	"errors"
)

const (
	port = ":8080"
)

type User struct {
	hostname string // may allow users to set their own username instead
	ip net.IP
	server bool // this may be unnecessary
}

func (u *User) String() string {
	ipStr := string([]byte(u.ip))
	return fmt.Sprint(u.hostname, "@", ipStr, ": ")
}

// This is a non-essential function, so just
// log errors and keep the program going
func ConfigureUser() (user *User) {
	user = &User{
		hostname: "Anonymous",
		ip: []byte("0.0.0.0"),
		server: false,
	}

	// Configure host name
	hn, err := os.Hostname()
	if err != nil {
		log.Println(err)
	}
	user.hostname = hn

	// Configure external IP address
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Println(err)
		return
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			log.Println(err)
			return
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}

			user.ip = []byte(ip.String())
			return
		}
	}

	log.Println(errors.New("are you connected to the network?"))
	return
}

func HandleAsServer(conn net.Conn) {
	uname := ConfigureUser()
	uname.server = true
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
	uname := ConfigureUser()
	input := bufio.NewScanner(conn)
	input.Split(bufio.ScanLines)
	for input.Scan() {
		fmt.Println("\r" + input.Text())
		fmt.Print(uname)
	}
	done <- struct{}{}
}

func HandleOutgoing(conn net.Conn, done chan <- struct{}) {
	defer conn.Close()
	uname := ConfigureUser()
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		io.WriteString(conn, uname.String()+input.Text()+"\n")
		fmt.Print(uname)
	}
	done <- struct{}{}
}

func main() {
	listen := flag.Bool("listen", false, "set to true when instance acts as server")
	flag.Parse()
	connIP := net.ParseIP(flag.Arg(0)).String()
	if connIP == "" {
		log.Fatal("invalid IP address given as argument")
	}

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