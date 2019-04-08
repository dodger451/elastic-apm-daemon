package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

func dataReceive(conn net.Conn, ch chan string) {

	buf := bufio.NewReader(conn)
	for {
		data, err := buf.ReadString('\n')
		if err != nil {
			break //Client disconnected.
		}
		fmt.Printf("Reading line from socket to channel: %d B \n", len(data))
		ch <- strings.TrimSuffix(data, "\n")
	}
	err := conn.Close()
	if err != nil {
		log.Fatal(err)
		return
	}
}

func dataSend(ch chan string, url string, contentheader string) {

	num := len(ch)
	if num == 0 {
		return
	}
	body := ""
	total := 0
	var message string
	for i := 0; i < num; i++ {
		message = <-ch
		body += fmt.Sprintf("%s\n", message)
		total += len(message)
	}

	log.Printf("Posting %d messages (%d byte) to %s\n", num, total, url)

	post([]byte(body), url, contentheader)
}

func post(data []byte, url string, contentheader string) {

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	check(err, "")

	req.Header.Set("Content-Type", contentheader)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print("Failed to send to server.")
		return
	}
	defer resp.Body.Close()

	log.Printf("Response Status:%s", resp.Status)
}

func check(err error, message string) {
	if err != nil {
		panic(err)
	}
	if len(message) > 0 {
		log.Printf("%s\n", message)
	}
}

func getOpts() (int, string, string, time.Duration, string) {
	sockPtr := flag.String("socket", "/tmp/.apm.sock", "default \"/tmp/.apm.sock\"")
	urlPtr := flag.String("url", "http://localhost:8200/v1/transmissions", "e.g. , "+
		"http://localhost:8200/v1/transmissions, http://localhost:8200/v1/errors "+
		"or http://localhost:8000/intake/v2/events")
	intervalPtr := flag.String("send-every", "60s", "e.g. 2m, 30s,... ")
	bufferPtr := flag.Int("buffer", 10000, "max number of lines")
	headerPtr := flag.String("content-type-header", "application/json", "v1 api:'application/json', v2: 'application/x-ndjson'")

	flag.Parse()

	intervalLength := *intervalPtr
	interval, err := time.ParseDuration(intervalLength)
	check(err, "")
	return *bufferPtr, *sockPtr, *urlPtr, interval,*headerPtr
}

func main() {
	buffer, socket, url, interval, contentheader := getOpts()

	err := os.Remove(socket)
	l, err := net.Listen("unix", socket)
	check(err, "")
	if err = os.Chmod(socket, 0766); err != nil {
		log.Fatal(err)
	}
	check(err, fmt.Sprintf("apm daemon is ready listing on socket '%s' (buffer: %d), sending data every '%s' to server at '%s' with content-type '%s'.", socket, buffer, interval.String(), url, contentheader))
	fmt.Println("Listen to ", socket)
	fmt.Println("Send to ", url)
	defer l.Close()

	ch := make(chan string, buffer)
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			go dataSend(ch, url, contentheader)
		}
	}()

	for {
		if len(ch) >= buffer-1 {
			go dataSend(ch, url, contentheader)
		}
		conn, err := l.Accept()
		check(err, "")
		go dataReceive(conn, ch)
	}
}
