package main

import (
	"strings"
	"testing"
)

func Test__readEmail(t *testing.T) {
	headers, body, err := readEmail([]byte(`To: Foobar <mail@host>

body`))
	if err != nil {
		t.Fatal("error", err)
	}

	if string(body) != `body` {
		t.Fatal("wrong body")
	}

	addresses, err := headers.AddressList(`To`)
	if err != nil {
		t.Fatal("error", err)
	}

	if len(addresses) != 1 {
		t.Fatal("no To header")
	}
	if addresses[0].Name != `Foobar` || addresses[0].Address != `mail@host` {
		t.Fatal("wrong to address")
	}
}

func Test__forward(t *testing.T) {
	var forwardedMails []struct {
		from, to, body string
	}
	err := forward("foo@bar.target", []byte("From: someone@tld\r\nSubject: foobar\r\n\r\nHello World"), func(from, to string, body []byte) error {
		forwardedMails = append(forwardedMails, struct {
			from string
			to   string
			body string
		}{from, to, string(body)})
		return nil
	})
	if err != nil {
		t.Fatal("error", err)
	}

	if len(forwardedMails) != 1 {
		t.Fatal("not forwarded")
	}
	if !strings.Contains(forwardedMails[0].body, "From: forwarder@localnet.cc\r\n") ||
		!strings.Contains(forwardedMails[0].body, "Subject: Forwarded: foobar from someone@tld\r\n") ||
		!strings.Contains(forwardedMails[0].body, "To: foo@bar.target\r\n") ||
		!strings.Contains(forwardedMails[0].body, "\r\n\r\nHello World") {
		t.Fatal("body", forwardedMails[0].body)
	}
	if forwardedMails[0].from != "forwarder@localnet.cc" {
		t.Fatal("from", forwardedMails[0].from)
	}
	if forwardedMails[0].to != "foo@bar.target" {
		t.Fatal("to", forwardedMails[0].to)
	}
}
