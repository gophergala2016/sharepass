package main

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/atotto/clipboard"
)

var HtmlTemplate = template.Must(template.New("password").Parse(html))

func main() {
	ip, err := getLocalAddr()
	if err != nil {
		log.Fatalln(err)
	}

	key, err := getSecretKey()
	if err != nil {
		log.Fatalln(err)
	}

	listener, err := net.Listen("tcp", ip+":0")
	if err != nil {
		log.Fatal(err)
	}
	addr := listener.Addr().String()

	pass, err := getPass(os.Stdin)
	if err != nil {
		log.Fatalf("Failed to read password: %s\n", err)
	}

	serveUrl := url.URL{
		Scheme: "http",
		Host:   addr,
		Path:   "/" + key,
	}
	log.Printf("Listening on %s\n", serveUrl.String())

	// TODO: add a boolean flag for this
	if err := clipboard.WriteAll(serveUrl.String()); err != nil {
		log.Printf("Error copying to clipboard: %s\n", err)
	}

	done := make(chan bool)

	http.HandleFunc("/"+key, func(res http.ResponseWriter, req *http.Request) {
		err := HtmlTemplate.Execute(res, pass)
		if err != nil {
			log.Printf("Error serving HTML: %s\n", err)
			return
		}
		done <- true
	})

	server := &http.Server{Addr: addr, Handler: nil}
	go server.Serve(listener)

	// wait until one successful request is complete
	<-done
}

func getLocalAddr() (ip string, err error) {
	host, err := os.Hostname()
	if err != nil {
		return
	}

	addrs, err := net.LookupIP(host)
	if err != nil {
		return
	}

	for _, addr := range addrs {
		ipv4 := addr.To4()
		if ipv4 != nil && !ipv4.IsLoopback() {
			ip = ipv4.String()
			break
		}
	}
	if ip == "" {
		err = errors.New("Failed to find local IP address")
	}
	return
}

func getSecretKey() (string, error) {
	randBytes := make([]byte, 16)
	_, err := rand.Read(randBytes)
	if err != nil {
		log.Fatal(err)
	}

	encoding := base64.URLEncoding.WithPadding(base64.NoPadding)
	key := encoding.EncodeToString(randBytes)

	return key, nil
}

func getPass(input io.Reader) (pass string, err error) {
	reader := bufio.NewReader(input)
	fmt.Printf("Enter password > ")
	pass, err = reader.ReadString('\n')
	pass = strings.TrimRight(pass, "\n\r ")
	if err != nil {
		return
	}
	return
}
