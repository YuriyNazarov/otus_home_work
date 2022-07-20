package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

var errConnFailed = errors.New("connection error")

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

type telnet struct {
	address string
	timeout time.Duration
	in      io.ReadCloser
	out     io.Writer
	Conn    net.Conn
}

func (t *telnet) Connect() error {
	log.Println(fmt.Sprintf("connecting to %s", t.address))
	conn, err := net.DialTimeout("tcp", t.address, t.timeout)
	if err != nil {
		log.Println("connection error: ", err)
		return errConnFailed
	}
	t.Conn = conn
	log.Println("__CONNECTED__")
	return nil
}

func (t *telnet) Close() error {
	log.Println("connection closed")
	return t.Conn.Close()
}

func (t *telnet) Send() error {
	if _, err := io.Copy(t.Conn, t.in); err != nil {
		t.Close()
		return fmt.Errorf("error on sending: %w", err)
	}
	return nil
}

func (t *telnet) Receive() error {
	if _, err := io.Copy(t.out, t.Conn); err != nil {
		t.Close()
		return fmt.Errorf("error on receiving: %w", err)
	}
	return nil
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &telnet{
		address: address,
		timeout: timeout,
		in:      in,
		out:     out,
	}
}
