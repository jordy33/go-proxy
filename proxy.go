package main

import (
	//"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"runtime"
)

var localAddress *string = flag.String("l", "localhost:9999", "Local address")
var remoteAddress *string = flag.String("r", "localhost:5000", "Remote address")


func main() {
	flag.Parse()

	fmt.Printf("Listening: %v\nProxying %v\n", *localAddress, *remoteAddress)

	addr, err := net.ResolveTCPAddr("tcp", *localAddress)
	if err != nil {
		panic(err)
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			panic(err)
		}

		go proxyConnection(conn)
	}

}

func proxyConnection(conn *net.TCPConn) {
	rAddr, err := net.ResolveTCPAddr("tcp", *remoteAddress)
	fmt.Println("Go routines:",runtime.NumGoroutine())
	if err != nil {
		panic(err)
	}

	rConn, err := net.DialTCP("tcp", nil, rAddr)
	if err != nil {
		// If the remote is not available
		defer conn.Close()
	}else {
		defer rConn.Close()
		stopchan := make(chan struct{})
		// a channel to signal that it's stopped
		stoppedchan := make(chan struct{})
		// Request loop
		go func() {
			defer close(stoppedchan)
			for {
				data := make([]byte, 1024*1024)
				n, err := conn.Read(data)
				if err != nil {
					//panic(err)
					fmt.Println("REQUEST Go routines:",runtime.NumGoroutine())
					close(stopchan)
					rConn.Close()
					return
				}else{
					rConn.Write(data[:n])
					//log.Printf("sent:\n%v", hex.Dump(data[:n]))
					//fmt.Println(string(data[:n]))
					var mem runtime.MemStats
					runtime.ReadMemStats(&mem)
					log.Printf("Allocated memory: %fMB. Number of goroutines: %d", float32(mem.Alloc)/1024.0/1024.0, runtime.NumGoroutine())
				}

			}
		}()

		// Response loop
		for {
			data := make([]byte, 1024*1024)
			n, err := rConn.Read(data)
			if err != nil {
				//panic(err)
				fmt.Println("RESPONSE Go routines:",runtime.NumGoroutine())
				defer conn.Close()
				rConn.Close()
				return
			}else{
				conn.Write(data[:n])
				//log.Printf("received:\n%v", hex.Dump(data[:n]))
				fmt.Println(string(n))
			}


		}

	}

}

