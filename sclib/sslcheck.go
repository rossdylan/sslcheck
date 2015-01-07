package sclib

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	_ "flag"
	"fmt"
	"io"
	"net"
	"net/smtp"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

type CertificateInfoList []*CertificateInfo

type CertificateInfo struct {
	name string
	cert *x509.Certificate
}

//The Len function for CertificateInfoList
//used as part of the implementation of sort.Interface
func (c CertificateInfoList) Len() int {
	return len(c)
}

//Swap 2 elements in a CertificateInfoList type
//used as part of the implementation of sort.Interface
func (c CertificateInfoList) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

//Calculate the order of 2 CertificateInfoList in a CertificateInfoList list
//used as part of the implementation of sort.Interface
func (c CertificateInfoList) Less(i, j int) bool {
	return c[i].name < c[j].name
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
func GenerateReport(certs CertificateInfoList, warningsOnly bool) string {
	sort.Sort(certs)
	pReader, pWriter := io.Pipe()
	var buff bytes.Buffer
	reportWriter := new(tabwriter.Writer)
	reportWriter.Init(pWriter, 0, 8, 0, '\t', 0)
	fmt.Fprintln(reportWriter, "Site\tCommon Name\tStatus\t   \tDays Left\tExpire Date")
	expiredCount := 0
	for _, cert := range certs {
		if cert != nil {
			eDate := cert.cert.NotAfter
			var expired string
			if IsExpired(eDate) {
				expired = "Expired"
				expiredCount++
			} else {
				expired = "Valid"
			}
			daysToExpire := GetExpireDays(eDate)
			cn := cert.cert.Subject.CommonName
			if (warningsOnly && IsExpired(eDate)) || !warningsOnly {
				fmt.Fprintf(reportWriter, "%s\t%s\t%s\t   \t%d\t%s\n", cert.name, cn, expired, daysToExpire, eDate.Local())
			}
		}
	}
	if expiredCount == 0 && warningsOnly {
		return ""
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
	days := int(delta.Hours() / 24)
	if days < 0 {
		return days
	} else {
		return days + 1
	}
	return 0
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
func CertGrabber(uri string, queue chan *CertificateInfo) {
	hostSplit := strings.Split(uri, ":")
	host := hostSplit[0]
	port, err := strconv.Atoi(hostSplit[1])
	if err != nil {
		panic(err)
	}
	queue <- &CertificateInfo{name: uri, cert: GetCerts(host, port)}

}
