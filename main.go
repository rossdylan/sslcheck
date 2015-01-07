package main

import (
	"flag"
	"fmt"
	"github.com/rossdylan/sslcheck/sclib"
	"os"
	"strings"
	"path/filepath"
	"bufio"
	"io"
)

var email string
var warning bool
var file string
// Longest possible domain is 253
// Additional 5 characters for port
var maxReadBuffer int = 259 // max domain length

func init() {
	flag.StringVar(&email, "email", "", "Send a full report to the given email")
	flag.BoolVar(&warning, "warning", false, "Send a warning report listing certs close to expiration")
  flag.StringVar(&file, "file", "", "File containing one endpoint per line")
}

//Main function, sets up all the channels needed for communication
//and grabs the uris to parse out of the arguments. We also monitor
//a channel to make sure all our workers die
func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [host:port ...]\n", filepath.Base(os.Args[0]))
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
	}
	flag.Parse()
	queue := make(chan *sclib.CertificateInfo, 10)
	var numCerts int
	numCerts = 0

	// If file is specified then use that as an input.
  if file != "" {
		linenum := 0

		f, err := os.Open(file)
		if err != nil {
			fatal_error(err.Error())
		}

    defer f.Close()

		r := bufio.NewReaderSize(f, maxReadBuffer)
		for {
			linenum++
			bytes, prefix, err := r.ReadLine()

			if err == io.EOF { break }
			if err != nil {
				fatal_error(err.Error())
			}

			line := string(bytes)

			// Do not support lines longer than maxReadBuffer.
			if prefix {
				fatal_error(fmt.Sprintf("Error: Line %d in file '%s' exceeds max length of %d bytes.", linenum, filepath.Base(file), maxReadBuffer))
			}

			// fmt.Printf("%d: %s\n", linenum, line)
			if strings.Contains(line, ":") {
				go sclib.CertGrabber(line, queue)
				numCerts += 1
			}
  	}
	}

	// Also check for additional commandline input
	for _, arg := range os.Args[1:] {
		if strings.Contains(arg, ":") {
			go sclib.CertGrabber(arg, queue)
			numCerts += 1
		}
	}
	if numCerts == 0 {
		fatal_usage("You must provide at least one host:port to be checked.")
	}
	var count int
	count = 0
	var certs sclib.CertificateInfoList
	certs = make(sclib.CertificateInfoList, numCerts)
	for cert := range queue {
		if cert != nil {
			certs[count] = cert
			count++
		}
		if count == numCerts {
			break
		}
	}
	var report string
	report = sclib.GenerateReport(certs, warning)
	if email != "" {
		if report == "" {
			return
		}
		sclib.MailReport(report, email)
	} else {
		fmt.Println(report)
	}
}

func fatal_error(e string) {
	defer os.Exit(1)
	fmt.Println(e)
}

func fatal_usage(e string) {
  defer fatal_error(fmt.Sprint("\n", e))
	flag.Usage()
}
