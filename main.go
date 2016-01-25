package main

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/atotto/clipboard"
)

var HtmlTemplate = template.Must(template.New("password").Parse(html))

const CHILD_PASSWORD_ENV = "_SHAREPASS_PASSWORD"

func main() {
	copyFlag := flag.Bool("copy", true, "Copy sharing URL to clipboard")
	daemonFlag := flag.Bool("daemon", true, "Run as background process after entering password")
	timeoutFlag := flag.Duration("timeout", time.Minute*10, "Timeout before exiting (e.g. 60s, 10m)")
	flag.Parse()

	log.SetFlags(0)

	pass := os.Getenv(CHILD_PASSWORD_ENV)
	if pass == "" {
		// Read the password
		inputPass, err := getPass(os.Stdin)
		if err != nil {
			log.Fatalf("Failed to read password: %s\n", err)
		}
		pass = inputPass

		if *daemonFlag {
			forkChild(pass)
			return
		}
	}

	// Get the local network address
	ip, err := getLocalAddr()
	if err != nil {
		log.Fatalln(err)
	}

	// Generate a secret key
	key, err := getSecretKey()
	if err != nil {
		log.Fatalln(err)
	}

	// Open a socket with a randomly assigned local port
	listener, err := net.Listen("tcp", ip+":0")
	if err != nil {
		log.Fatal(err)
	}
	addr := listener.Addr().String()

	// Generate the URL to send to the recipient
	serveUrl := url.URL{
		Scheme: "http",
		Host:   addr,
		Path:   "/" + key,
	}
	url := serveUrl.String()
	log.Printf("Listening on %s for %s\n", url, *timeoutFlag)

	// Copy the URL to the clipboard, if enabled
	if *copyFlag {
		if err := clipboard.WriteAll(url); err != nil {
			log.Printf("Error copying to clipboard: %s\n", err)
		} else {
			log.Printf("Copied to clipboard: \"%s\"\n", url)
		}
	}

	// channel for signalling success or timeout
	done := make(chan bool)
	time.AfterFunc(*timeoutFlag, func() {
		done <- true
	})

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

	// wait until one successful request is complete or timeout occurs
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
	fmt.Fprintf(os.Stderr, "Enter password > ")
	pass, err = reader.ReadString('\n')
	pass = strings.TrimRight(pass, "\n\r ")
	if err != nil {
		return
	}
	return
}

func forkChild(pass string) (err error) {
	attr := &os.ProcAttr{
		Env:   append(os.Environ(), CHILD_PASSWORD_ENV+"="+pass),
		Files: []*os.File{nil, nil, os.Stderr},
	}
	if _, err = os.StartProcess("sharepass", os.Args, attr); err != nil {
		return
	}

	return nil
}
