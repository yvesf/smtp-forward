package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"flag"
	"io"
	"log"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"

	"github.com/mhale/smtpd"
	"github.com/pkg/errors"
)

var flagListen = flag.String(`l`, `:25`, `Address to listen on`)
var flagHostname = flag.String(`h`, `HOSTNAME-NOT-SET`, `Server flagHostname`)
var flagMap = flag.String(`m`, ``, `-m prefix-matcher1:target@targethost,prefix-matcher2:target@targethost`)
var flagCertFile = flag.String(`c`, ``, ``)
var flagKeyFile = flag.String(`k`, ``, ``)
var flagFrom = flag.String(`f`, `forwarder@localnet.cc`, `From of forwarded email`)

func logErrorSMTPMiddleware(handler smtpd.Handler) smtpd.Handler {
	return func(remoteAddr net.Addr, from string, to []string, data []byte) (err error) {
		log.Printf(`received email remoteAddr=%v from=%v`, remoteAddr, from)
		err = handler(remoteAddr, from, to, data)
		if err != nil {
			log.Printf(`failed to forward: %v`, err.Error())
			return err
		}
		log.Printf(`handled email remoteAddr=%v from=%v`, remoteAddr, from)
		return nil
	}
}

func readEmail(data []byte) (headers mail.Header, msgData []byte, err error) {
	msg, err := mail.ReadMessage(bytes.NewReader(data))
	if err != nil {
		return nil, nil, errors.Wrap(err, `failed to parse email`)
	}
	headers = make(mail.Header)

	msgData, err = io.ReadAll(msg.Body)
	if err != nil {
		return nil, nil, errors.Wrap(err, `failed to parse email body`)
	}

	for headerKey, headerVal := range msg.Header {
		headers[headerKey] = headerVal
	}

	return headers, msgData, nil
}

func forward(targetEmail string, data []byte, sendEmail func(string, string, []byte) error) error {
	headers, msgData, err := readEmail(data)
	if err != nil {
		return errors.Wrap(err, `failed readEmail`)
	}
	headers[textproto.CanonicalMIMEHeaderKey(`Subject`)] = []string{
		`Forwarded: ` + headers.Get(`subject`) + ` from ` + headers.Get("From")}
	headers[textproto.CanonicalMIMEHeaderKey(`To`)] = []string{targetEmail}
	headers[textproto.CanonicalMIMEHeaderKey(`From`)] = []string{*flagFrom}

	var builder bytes.Buffer
	for headerName, headerValues := range headers {
		for _, value := range headerValues {
			builder.WriteString(textproto.CanonicalMIMEHeaderKey(headerName))
			builder.WriteString(`: `)
			builder.WriteString(value)
			builder.WriteString("\r\n")
		}
	}
	builder.WriteString("\r\n")
	builder.Write(msgData)

	var retryCount = 5
	for retryCount > 0 {
		err = sendEmail(*flagFrom, targetEmail, builder.Bytes())
		if err, ok := err.(*textproto.Error); ok {
			if 400 <= err.Code && err.Code < 500 {
				log.Printf(`retry sleep 120s count=%v code=%v msg=%v`, retryCount, err.Code, err.Msg)
				time.Sleep(120 * time.Second)
				retryCount--
				continue
			}
		}
		if err != nil {
			return errors.Wrap(err, `failed targetEmail send mail via smtp`)
		}
		break
	}
	log.Printf("forwarded targetEmail=%v", targetEmail)

	return nil
}

func makeEmailHandler(mapping map[string]string) smtpd.Handler {
	return func(remoteAddr net.Addr, from string, to []string, data []byte) error {
		for prefix, targetEmail := range mapping {
			for _, to := range to {
				if strings.HasPrefix(to, prefix) {
					err := forward(targetEmail, data, sendEmail)
					if err != nil {
						log.Print(`forwarded failed `, to, ` to `, targetEmail, err.Error())
					}
				}
			}
		}
		return nil
	}
}

func sendEmail(from string, to string, body []byte) error {
	toSplitLocalPart := strings.SplitN(to, `@`, 2)
	if len(toSplitLocalPart) != 2 {
		return errors.New(`invalid targetEmail address: ` + to)
	}

	mxes, err := net.DefaultResolver.LookupMX(context.Background(), toSplitLocalPart[1])
	if err != nil || len(mxes) == 0 {
		return errors.Wrap(err, `failed targetEmail resolve mx`)
	}
	err = smtp.SendMail(mxes[0].Host+":25", nil, from, []string{to}, body)
	return err
}

func main() {
	flag.Parse()

	// parse mapping given by commandline flag separated by colon and comma
	var mapping = make(map[string]string)
	for _, m := range strings.Split(*flagMap, `,`) {
		if m == `` {
			continue
		}
		m := strings.SplitN(m, `:`, 2)
		if len(m) != 2 {
			panic(`invalid flag -m: ` + *flagMap)
		}
		mapping[m[0]] = m[1]
	}

	var tlsConfig *tls.Config
	if *flagCertFile != `` && *flagKeyFile != `` {
		cert, err := tls.LoadX509KeyPair(*flagCertFile, *flagKeyFile)
		if err != nil {
			panic(err)
		}
		tlsConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	}

	server := &smtpd.Server{
		TLSConfig: tlsConfig,
		Addr:      *flagListen,
		Hostname:  *flagHostname,
		Handler:   logErrorSMTPMiddleware(makeEmailHandler(mapping)),
		AuthMechs: map[string]bool{"LOGIN": true, "PLAIN": true},
		AuthHandler: func(remoteAddr net.Addr, mechanism string, username []byte, password []byte, shared []byte) (bool, error) {
			log.Printf("accept auth %v %v", string(username), string(password))
			return true, nil
		},
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Printf("server exited with error: %v", err)
	}
}
