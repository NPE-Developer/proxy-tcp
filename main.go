package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func readConfig() (host, target string) {
	bytes, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatal("failed to read config.json file: ", err.Error())
	}

	var value map[string]string
	if err := json.Unmarshal(bytes, &value); err != nil {
		log.Fatal("failed to read config.json file: ", err.Error())
	}

	return value["host"], value["target"]
}

func handleConnection(conn net.Conn, target string) error {
	proxyConn, err := net.Dial("tcp", target)
	if err != nil {
		return err
	}
	defer func() {
		proxyConn.Close()
		conn.Close()
		fmt.Println("[INFO] client disconnected")
	}()

	fmt.Println("[INFO] client connected")
	go io.Copy(conn, proxyConn)
	io.Copy(proxyConn, conn)

	return nil
}

func listener(host string, target string) error {
	listener, err := net.Listen("tcp", host)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Printf("[INFO] started proxy at %s with target %s\n", host, target)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("[ERROR] error accepting connection: ", err.Error())
			return err
		}

		// handle connection
		go func() {
			if err := handleConnection(conn, target); err != nil {
				fmt.Println("[ERROR] error handling connection: ", err.Error())
			}
		}()
	}

}

func main() {
	var host, target string
	if len(os.Args) < 3 {
		host, target = readConfig()
	} else {
		host = os.Args[1]
		target = os.Args[2]
	}

	if host == "" || target == "" {
		fmt.Println("[INFO] usage: proxy.exe [host] [target]")
		fmt.Println("[INFO] or create config.json file with {host: string, target: string} value")
		os.Exit(1)
	}

	if err := listener(host, target); err != nil {
		log.Fatal(err)
	}
}
