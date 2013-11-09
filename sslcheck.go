package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	_ "flag"
	"fmt"
	"io"
	"net"
	"net/smtp"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

//Alias []*x509.Certificate to Certificates
type Certificates []*x509.Certificate

//The Len function for Certificates
//used as part of the implementation of sort.Interface
func (c Certificates) Len() int {
	return len(c)
}

//Swap 2 elements in a Certificates type
//used as part of the implementation of sort.Interface
func (c Certificates) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

//Calculate the order of 2 certificates in a Certificates list
//used as part of the implementation of sort.Interface
func (c Certificates) Less(i, j int) bool {
	return c[i].Subject.CommonName < c[j].Subject.CommonName
}

//Give a report (just a string) and an adresss to send it to
//Send that report out via email
func MailReport(report, to_addr string) {
	smtpAuth := smtp.PlainAuth("", "", "", "localhost")
	err := smtp.SendMail("localhost:25", smtpAuth, "no-reply@csh.rit.edu", []string{to_addr}, []byte(report))
	if err != nil {
		panic(err)
	}
}

//Given a Certificates list, create a tabular report of
//the relevant information in string format
func GenerateReport(certs Certificates) string {
	sort.Sort(certs)
	pReader, pWriter := io.Pipe()
	var buff bytes.Buffer
	reportWriter := new(tabwriter.Writer)
	reportWriter.Init(pWriter, 0, 8, 0, '\t', 0)
	fmt.Fprintln(reportWriter, "Site\tStatus\t   \tDays Left\tExpire Date")
	for _, cert := range certs {
		if cert != nil {
			eDate := cert.NotAfter
			var expired string
			if IsExpired(eDate) {
				expired = "Expired"
			} else {
				expired = "Valid"
			}
			daysToExpire := GetExpireDays(eDate)
			cn := cert.Subject.CommonName
			fmt.Fprintf(reportWriter, "%s\t%s\t   \t%d\t%s\n", cn, expired, daysToExpire, eDate.Local())
		}
	}
	go buff.ReadFrom(pReader)
	reportWriter.Flush()
	pWriter.Close()
	pReader.Close()
	return buff.String()
}

//Given a time.Time struct check to see if the current time is after it
func IsExpired(then time.Time) bool {
	return time.Now().After(then)
}

//Get the number days until a certificate expires
//Calculated based on the time.Time the certificate expires
func GetExpireDays(then time.Time) int {
	now := time.Now()
	delta := then.Sub(now)
	return int(delta.Hours() / 24)
}

//Give a hostname and a port grab the certificate for the service
//running there
func GetCerts(host string, port int) *x509.Certificate {
	uri := host + fmt.Sprintf(":%d", port)
	connection, err := net.Dial("tcp", uri)
	if err != nil {
		panic(err)
	}
	defer connection.Close()
	config := tls.Config{ServerName: host, InsecureSkipVerify: true}
	secConn := tls.Client(connection, &config)
	defer secConn.Close()

	handshakeError := secConn.Handshake()
	if handshakeError != nil {
		return nil
	}
	certs := secConn.ConnectionState().PeerCertificates
	if certs == nil || len(certs) < 1 {
		return nil
	}
	return certs[0]

}

//Grab a certificate from the given uri and send it down a channel to be processed
//Then report our exit to the main thread using another channel
//This is intended to be run via a goroutine
func CertGrabber(uri string, queue chan *x509.Certificate, exitQueue chan bool) {
	hostSplit := strings.Split(uri, ":")
	host := hostSplit[0]
	port, err := strconv.Atoi(hostSplit[1])
	if err != nil {
		panic(err)
	}
	queue <- GetCerts(host, port)
	exitQueue <- true

}

//Main function, sets up all the channels needed for communication
//and grabs the uris to parse out of the arguments. We also monitor
//a channel to make sure all our workers die
func main() {
	queue := make(chan *x509.Certificate, 10)
	exitQueue := make(chan bool)
	var numThreads int
	numThreads = 0
	for _, arg := range os.Args[1:] {
		go CertGrabber(arg, queue, exitQueue)
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
	var certs Certificates
	certs = make(Certificates, 0)
	for cert := range queue {
		if cert != nil {
			certs = append(certs, cert)
		} else {
			break
		}
	}
	report := GenerateReport(certs)
	fmt.Println(report)
}
