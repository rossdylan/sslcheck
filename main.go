package main

import (
	"crypto/x509"
	"fmt"
	"github.com/rossdylan/sslcheck/sclib"
	"os"
)

//Main function, sets up all the channels needed for communication
//and grabs the uris to parse out of the arguments. We also monitor
//a channel to make sure all our workers die
func main() {
	queue := make(chan *x509.Certificate, 10)
	exitQueue := make(chan bool)
	var numThreads int
	numThreads = 0
	for _, arg := range os.Args[1:] {
		go sclib.CertGrabber(arg, queue, exitQueue)
		numThreads += 1
	}
	var count int
	count = 0
	for val := range exitQueue {
		if val {
			count += 1
		}
		if count == numThreads {
			break
		}
	}
	queue <- nil
	var certs sclib.Certificates
	certs = make(sclib.Certificates, 0)
	for cert := range queue {
		if cert != nil {
			certs = append(certs, cert)
		} else {
			break
		}
	}
	report := sclib.GenerateReport(certs)
	fmt.Println(report)
}
