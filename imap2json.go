package main

import (
	"./go-imap/go1/imap"
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/jhillyerd/go.enmime"
	"io/ioutil"
	"log"
	"net"
	"net/mail"
	"net/url"
	"os"
	"strings"
	"time"
)

type Msg struct {
	Header mail.Header
	UID    int
	Body   string
}

type Conversation struct {
	Id   string
	Msgs []Msg
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s imap://imap.dabase.com\n", os.Args[0])
	os.Exit(2)
}

func dumplist(x interface{}) []int {

	l := []int{}

	switch t := x.(type) {

	case []imap.Field:
		for _, v := range t {
			//fmt.Println(i)
			l = append(l, dumplist(v)...)
		}
	case uint32:
		l = append(l, int(t))
	default:
		fmt.Printf("Unhandled: %T\n", t)
	}
	return l
}

func dumpl(x interface{}) [][]int {

	l := [][]int{}

	switch t := x.(type) {

	case []imap.Field:
		for _, v := range t {
			//fmt.Println(i)
			l = append(l, dumplist(v))
		}
	default:
		fmt.Printf("Unhandled: %T\n", t)
	}
	return l
}

func main() {

	if len(os.Args) != 2 {
		usage()
	}

	iurl, err := url.ParseRequestURI(os.Args[1])
	if err != nil {
		usage()
	}

	if iurl.Scheme != "imaps" && iurl.Scheme != "imap" {
		usage()
	}

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

	// Comment out to turn off debug info
	imap.DefaultLogger = log.New(os.Stdout, "", 0)
	imap.DefaultLogMask = imap.LogConn | imap.LogRaw

	if iurl.Scheme == "imaps" {
		fmt.Println("Making a secure connection to", iurl.Host)
		c, err = imap.DialTLS(iurl.Host, nil)
		if err != nil {
			fmt.Println(err.Error())
		}

	} else { // It's just imap
		c, _ = imap.Dial(iurl.Host)
	}

	// Logout once done
	defer func() { c.Logout(30 * time.Second) }()

	//fmt.Println("Server says hello:", c.Data[0].Info)
	//c.Data = nil

	if iurl.User == nil {
		fmt.Println("Logging in Anonymously...")
		c.Anonymous()
	} else {
		// Authenticate
		if c.State() == imap.Login {
			user := iurl.User.Username()
			pass, _ := iurl.User.Password()
			c.Login(user, pass)
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

	rcmd, err := imap.Wait(c.Send("UID THREAD", "references UTF-8 all"))
	if err != nil {
		fmt.Println("Your IMAPD server", iurl.Host, "sadly does not support UID THREAD (rfc5256)")
		fmt.Println("Please consider exporting your email and serving it via http://dovecot.org/ IMAPD")
		panic(err)
	}

	flat := dumpl(rcmd.Data[0].Fields[1:])
	fmt.Println("Flat:", flat)

	err = os.MkdirAll("cache", 0777)
	if err != nil {
		panic(err)
	}

	// Fetch everything TODO: Only fetch what's in THREAD but not in cache/
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
				id := int(m.UID)
				s := fmt.Sprintf("cache/%d.txt", id)
				// Writing out message ids to cache
				// fmt.Printf("WROTE: %d\n", id)
				err := ioutil.WriteFile(s, entiremsg, 0644)
				if err != nil {
					panic(err)
				}
			}
		}
		cmd.Data = nil
	}

	// Refer to Array based structure in JSON-design.mdwn

	var archive []Conversation
	for _, j := range flat {
		var c Conversation
		for i, k := range j {
			if i == 0 { // First message gets hashed
				s := fmt.Sprintf("cache/%d.txt", k)
				entiremsg, err := ioutil.ReadFile(s)
				if err != nil {
					panic(err) // continue ?
				}
				h := sha1.New()
				h.Write(entiremsg)
				c.Id = fmt.Sprintf("%x", h.Sum(nil))
				m, err := getMsg(k)
				if err != nil {
					m = Msg{Header: nil, Body: "Missing " + string(k)}
				}
				c.Msgs = append(c.Msgs, m)
			} else { // Subsequent messages in the conversation
				m, err := getMsg(k)
				if err != nil {
					m = Msg{Header: nil, Body: "Missing " + string(k)}
				}
				c.Msgs = append(c.Msgs, m)
			}
		}
		archive = append(archive, c)
	}

	// DEBUG CRUFT
	// fmt.Println(archive)
	// for _, v := range archive {
	// fmt.Println("Hash:", v.Id)
	// fmt.Println("Messages:", len(v.Msgs))
	// }

	// Marshall to mail.json
	json, _ := json.MarshalIndent(archive, "", " ")
	err = ioutil.WriteFile("mail.json", json, 0644)
	if err != nil {
		panic(err)
	} else {
		fmt.Println("Built mail.json\n")
	}

}

func getMsg(id int) (m Msg, err error) {
	s := fmt.Sprintf("cache/%d.txt", id)
	entiremsg, err := ioutil.ReadFile(s)
	if err != nil {
		fmt.Println("Not fetched:", id)
		return m, err
	}
	if msg, _ := mail.ReadMessage(bytes.NewReader(entiremsg)); msg != nil {
		if enmime.IsMultipartMessage(msg) {
			mime, err := enmime.ParseMIMEBody(msg)
			if err != nil {
				//fmt.Println("Trying to read", id)
				//panic(err)
				m.Body = err.Error()
			} else {
				m.Body = mime.Text
			}
		} else {
			body, _ := ioutil.ReadAll(msg.Body)
			m.Body = string(body)
		}
		m.UID = id

		// Trying to prune headers we don't need to keep mail.json size down
		delete(msg.Header, "Content-Disposition")
		delete(msg.Header, "Content-Transfer-Encoding")
		delete(msg.Header, "Content-Type")
		delete(msg.Header, "Delivered-To")
		delete(msg.Header, "Dkim-Signature")
		delete(msg.Header, "In-Reply-To")
		delete(msg.Header, "List-Archive")
		delete(msg.Header, "List-Help")
		delete(msg.Header, "List-Id")
		delete(msg.Header, "List-Post")
		delete(msg.Header, "List-Subscribe")
		delete(msg.Header, "List-Unsubscribe")
		delete(msg.Header, "Message-Id")
		delete(msg.Header, "Mime-Version")
		delete(msg.Header, "Precedence")
		delete(msg.Header, "Received")
		delete(msg.Header, "References")
		delete(msg.Header, "Reply-To")
		delete(msg.Header, "Resent-Cc")
		delete(msg.Header, "Resent-Date")
		delete(msg.Header, "Resent-From")
		delete(msg.Header, "Resent-Message-Id")
		delete(msg.Header, "Resent-To")
		delete(msg.Header, "Return-Path")
		delete(msg.Header, "Sender")
		delete(msg.Header, "Thread-Index")
		delete(msg.Header, "Thread-Topic")
		delete(msg.Header, "User-Agent")
		delete(msg.Header, "Resent-Sender")
		delete(msg.Header, "Accept-Language")
		delete(msg.Header, "Content-Language")
		delete(msg.Header, "Errors-To")
		for key := range msg.Header {
			if strings.HasPrefix(key, "X-") {
				delete(msg.Header, key)
			}
		}
		m.Header = msg.Header
	}
	return m, nil
}
