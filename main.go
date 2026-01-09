package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func readConfig() (host, target string, err error) {
	bytes, err := os.ReadFile("config.json")
	if err != nil {
		return "", "", fmt.Errorf("failed to read config.json file: %s", err.Error())
	}

	var value map[string]string
	if err := json.Unmarshal(bytes, &value); err != nil {
		return "", "", fmt.Errorf("failed to read config.json file: %s", err.Error())
	}

	return value["host"], value["target"], nil
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

func scanPorts(target string, from int, to int, timeout time.Duration) []string {
	open := []string{}
	for i := range to - from {
		port := strconv.Itoa(i + from)

		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", target, port), timeout)
		if err != nil {
			continue
		}
		conn.Close()

		fmt.Printf("found open port %s:%s\n", target, port)
		open = append(open, port)
	}

	return open
}

func main() {
	var host, target string
	var timeout_ms int

	if len(os.Args) < 3 {
		host, target, _ = readConfig() // no need to handle err
	} else {
		host = os.Args[1]
		target = os.Args[2]

		if len(os.Args) == 4 {
			timeout_ms, _ = strconv.Atoi(os.Args[3])
		}
		if timeout_ms == 0 {
			timeout_ms = 100
		}
	}

	if host == "" || target == "" {
		fmt.Println(`[INFO] usage: proxy.exe [host] [target] [timeout_ms]`)
		fmt.Println(`[INFO] or create config.json file with {host: string, target: string} value`)
		fmt.Println(`[INFO] example: "./proxy.exe 127.0.0.1:80 192.168.100.10:3000"`)
		os.Exit(1)
	}

	if len(strings.Split(host, ":")) == 1 && len(strings.Split(target, ":")) == 1 {
		for {
			mut := sync.Mutex{}
			wg := sync.WaitGroup{}

			ports := []string{}
			max_port := 49152 // well known ports (0-1023) sampai registered ports (1024-49151)
			divided := max_port / 8

			for i := range 8 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					result := scanPorts(target, i*divided, (i+1)*divided, time.Duration(timeout_ms)*time.Millisecond)

					mut.Lock()
					defer mut.Unlock()

					ports = append(ports, result...)
				}()
			}
			wg.Wait()

			for _, port := range ports {
				go func() {
					for {
						if err := listener(fmt.Sprintf("%s:%s", host, port), fmt.Sprintf("%s:%s", target, port)); err != nil {
							fmt.Printf("[INFO ] closed proxy to %s:%s: %s\n", target, port, err.Error())
						}
						fmt.Printf("[INFO ] restarting proxy to %s:%s in %s\n", target, port, (3 * time.Second).String())
						time.Sleep(3 * time.Second)
					}
				}()
			}

			select {}
		}
	} else {
		if err := listener(host, target); err != nil {
			log.Fatal(err)
		}
	}
}
