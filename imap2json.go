package main

import (
	"./go-imap/go1/imap"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/mail"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Msg struct {
	Header mail.Header
	Body   string
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s imap://imap.dabase.com\n", os.Args[0])
	os.Exit(2)
}

func main() {

	if len(os.Args) != 2 {
		usage()
	}

	iurl, err := url.ParseRequestURI(os.Args[1])
	if err != nil {
		usage()
	}

	if iurl.Scheme != "imap" {
		usage()
	}

	//fmt.Println(iurl.User)
	//fmt.Println("Host:", iurl.Host)
	//fmt.Println(iurl.Path)

	var (
		c   *imap.Client
		cmd *imap.Command
		rsp *imap.Response
	)

	// Lets check if we can reach the host
	tc, err := net.Dial("tcp", iurl.Host+":"+iurl.Scheme)
	if err == nil {
		tc.Close()
		fmt.Printf("Dial to %s succeeded\n", iurl.Host)
	} else {
		panic(err)
	}

	// Temporary data structure that will be marshalled into JSON
	e2j := map[string]interface{}{}

	// Comment out to turn off debug info
	imap.DefaultLogger = log.New(os.Stdout, "", 0)
	imap.DefaultLogMask = imap.LogConn | imap.LogRaw

	c, _ = imap.Dial(iurl.Host)

	defer func() { c.Logout(30 * time.Second) }()

	// Not sure why this has to be nulled
	c.Data = nil

	if iurl.User == nil {
		fmt.Println("Logging in Anonymously...")
		c.Anonymous()
	} else {
		// Authenticate
		if c.State() == imap.Login {
			user := iurl.User.Username()
			pass, _ := iurl.User.Password()
			fmt.Println("Logging in as user:", user)
			_, err = c.Login(user, pass)
		} else {
			fmt.Printf("Login not presented")
			return
		}

		if err != nil {
			fmt.Printf("login failed, exiting...\n")
			return
		}
	}

	if iurl.Path != "" {
		// Remove / prefix
		mailbox := iurl.Path[1:]
		fmt.Println("Selecting mailbox:", mailbox)
		c.Select(mailbox, true)
	} else {
		c.Select("INBOX", true)
	}

	rcmd, err := imap.Wait(c.Send("THREAD", "references UTF-8 all")) // Do we need UID option here?
	if err != nil {
		panic(err)
	}

	// Export thread information
	e2j["THREAD"] = rcmd.Data[0].Fields[1:]

	// Fetch everything
	set, _ := imap.NewSeqSet("1:*")
	cmd, _ = c.Fetch(set, "UID", "BODY[]")

	// Process responses while the command is running
	for cmd.InProgress() {
		// Wait for the next response (no timeout)
		c.Recv(-1)

		// Process message response into temporary data structure
		for _, rsp = range cmd.Data {
			m := rsp.MessageInfo()
			entiremsg := imap.AsBytes(m.Attrs["BODY[]"])
			if msg, _ := mail.ReadMessage(bytes.NewReader(entiremsg)); msg != nil {
				body, _ := ioutil.ReadAll(msg.Body)
				id := int64(m.UID)
				e2j[strconv.FormatInt(id, 10)] = Msg{Header: msg.Header, Body: string(body)}
			}
		}
		cmd.Data = nil

	}

	// Marshall to mail.json
	backtoj, _ := json.MarshalIndent(e2j, "", " ")
	err = ioutil.WriteFile("mail.json", backtoj, 0644)
	if err != nil {
		panic(err)
	} else {
		fmt.Println("Wrote mail.json")
	}

}
