package smtp

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"strings"
)

type Client struct {
	conn      net.Conn
	localName string
	br        *bufio.Reader
}

func Dial(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	c := &Client{conn: conn, localName: "localhost", br: bufio.NewReader(conn)}
	code, msg, _, err := c.parseResp()
	if err != nil {
		conn.Close()
		return nil, err
	}
	log.Printf("Dial: %s#%s\n", code, msg)
	return c, nil
}

func (c *Client) EHLO(localName string) error {
	if localName != "" {
		c.localName = localName
	}
	c.conn.Write([]byte("EHLO " + c.localName + "\r\n"))
	code, msg, more, err := c.parseResp()
	if err != nil {
		return err
	}
	log.Printf("EHLO: %s:%s\n", code, msg)
	if code != "250" {
		return fmt.Errorf("EHLO: %s#%s", code, msg)
	}
	for _, m := range more {
		log.Printf("EHLO: %s\n", m)
	}
	return nil
}

func (c *Client) Auth(user, password string) error {
	var code, message string
	var err error
	c.conn.Write([]byte("AUTH LOGIN\r\n"))
	code, message, _, err = c.parseResp()
	if err != nil {
		return fmt.Errorf("Auth failed: %v", err)
	}
	if code != "334" && message != "VXNlcm5hbWU6" {
		return fmt.Errorf("Auth failed: %s#%s\n", code, message)
	}

	c.conn.Write([]byte(base64.StdEncoding.EncodeToString([]byte(user)) + "\r\n"))
	code, message, _, err = c.parseResp()
	if err != nil {
		return fmt.Errorf("Auth failed: %v", err)
	}
	if code != "334" && message != "UGFzc3dvcmQ6" {
		return fmt.Errorf("Auth failed: %s#%s\n", code, message)
	}

	c.conn.Write([]byte(base64.StdEncoding.EncodeToString([]byte(password)) + "\r\n"))
	code, message, _, err = c.parseResp()
	if err != nil {
		return fmt.Errorf("Auth failed: %v", err)
	}
	if code != "235" {
		return fmt.Errorf("Auth failed: %s#%s\n", code, message)
	}
	return nil
}

func (c *Client) Close() error {
	c.conn.Write([]byte("QUIT\r\n"))
	code, message, _, err := c.parseResp()
	if err != nil {
		return nil
	}
	if code != "221" {
		return fmt.Errorf("Close: failed to quit: %s#%s", code, message)
	}
	return c.conn.Close()
}

func (c *Client) parseResp() (code, msg string, more []string, err error) {
	for {
		resp, _ := c.br.ReadString('\n')
		if len(resp) > 4 {
			code = resp[0:3]
			if resp[3:4] == "-" {
				more = append(more, strings.TrimSpace(resp))
			} else {
				msg = strings.TrimSpace(resp[4:])
				return
			}
		} else {
			err = fmt.Errorf("parseResp: invalid message")
			return
		}
	}
}
