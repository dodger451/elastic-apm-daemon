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
		log.Printf("Reading line from socket to channel: %d B \n", len(data))
		ch <- strings.TrimSuffix(data, "\n")
	}
	err := conn.Close()
	if err != nil {
		log.Fatal(err)
		return
	}
}

func dataSend(ch chan string, url string) {

	log.Printf("Messages gathered: %d \n", len(ch))
	num := len(ch)
	if num == 0 {
		return
	};
	body := ""
	total := 0
	var message string
	for i := 0; i < num; i++ {
		message = <-ch
		body += fmt.Sprintf("%s\n", message)
		total += len(message)
	}

	log.Printf("Posting %d messages with total %d B", num, total)
	post([]byte(body), url)
}

func post(data []byte, url string) {
	log.Printf("URL:>%s", url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	check(err, "")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print("Failed to send to server.")
		return;
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

func getOpts() (int, string, string, time.Duration) {
	sockPtr := flag.String("socket", "/tmp/.apm.sock", "default \"/tmp/.apm.sock\"")
	urlPtr := flag.String("url", "http://localhost:8000/simple.php", "e.g. http://localhost:8000/intake/v2/events")
	intervalPtr := flag.String("send-every", "60s", "e.g. 2m, 30s,... ")
	bufferPtr := flag.Int("buffer", 10000, "max number of lines")
	flag.Parse()

	intervalLength := *intervalPtr
	interval, err := time.ParseDuration(intervalLength)
	check(err, fmt.Sprintf("Sending data every %s.", interval.String()))
	return *bufferPtr, *sockPtr, *urlPtr, interval
}

func main() {
	buffer, socket, url, interval:= getOpts()

	err := os.Remove(socket)
	l, err := net.Listen("unix", socket)
	check(err, "apm daemon is ready.")
	defer l.Close()

	ch := make(chan string, buffer)
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			go dataSend(ch, url)
		}
	}()

	for {
		if len(ch) >= buffer-1 {
			go dataSend(ch, url)
		}
		conn, err := l.Accept()
		check(err, "")
		go dataReceive(conn, ch)
	}
}
