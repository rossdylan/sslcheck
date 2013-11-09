package main

import (
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/rossdylan/sslcheck/sclib"
	"os"
	"strings"
)

var email string
var warning bool

func init() {
	flag.StringVar(&email, "email", "", "Send a full report to the given email")
	flag.BoolVar(&warning, "warning", false, "Send a warning report listing certs close to expiration")
}

//Main function, sets up all the channels needed for communication
//and grabs the uris to parse out of the arguments. We also monitor
//a channel to make sure all our workers die
func main() {
	flag.Parse()
	queue := make(chan *x509.Certificate, 10)
	var numCerts int
	numCerts = 0
	for _, arg := range os.Args[1:] {
		if strings.Contains(arg, ":") {
			go sclib.CertGrabber(arg, queue)
			numCerts += 1
		}
	}
	if numCerts == 0 {
		fmt.Println("Enter at least 1 service")
		return
	}
	var count int
	count = 0
	var certs sclib.Certificates
	certs = make(sclib.Certificates, numCerts)
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
