package client

import (
	"bytes"
	"crypto"
	"fmt"
	"github.com/curltech/go-colla-core/logger"
	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-msgauth/dkim"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/emersion/go-message"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
)

type SmtpClient struct {
	*smtp.Client
	addr string
	data io.WriteCloser
}

var smtpClient *SmtpClient

func init() {
	smtpClient = &SmtpClient{}
}

func GetSmtpClient() *SmtpClient {
	return smtpClient
}

func (this *SmtpClient) Send(addr string, username string, password string, from string, to []string, message string) error {
	// Set up authentication information.
	auth := sasl.NewPlainClient("", username, password)

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	msg := strings.NewReader(message)
	err := smtp.SendMail(addr, auth, from, to, msg)
	if err != nil {
		logger.Errorf(err.Error())

		return err
	}

	return nil
}

func (this *SmtpClient) Dial(addr string, sender string, recipient string) error {
	// Connect to the remote SMTP server.
	var err error
	this.addr = addr
	this.Client, err = smtp.Dial(this.addr)
	if err != nil {
		logger.Errorf(err.Error())

		return err
	}

	// Set the sender and recipient first
	if err := this.Client.Mail(sender, nil); err != nil {
		logger.Errorf(err.Error())

		return err
	}
	if err := this.Client.Rcpt(recipient); err != nil {
		logger.Errorf(err.Error())

		return err
	}

	this.data, err = this.Client.Data()
	if err != nil {
		logger.Errorf(err.Error())

		return err
	}

	return nil
}

func (this *SmtpClient) Print(format string, a ...interface{}) error {
	var err error
	_, err = fmt.Fprintf(this.data, format, a...)
	if err != nil {
		logger.Errorf(err.Error())

		return err
	}

	return nil
}

func (this *SmtpClient) Quit() error {
	var err error
	err = this.data.Close()
	if err != nil {
		logger.Errorf(err.Error())

		return err
	}

	// Send the QUIT command and close the connection.
	err = this.Client.Quit()
	if err != nil {
		logger.Errorf(err.Error())

		return err
	}

	return nil
}

func (this *SmtpClient) Sign(msg string, signer crypto.Signer) error {
	r := strings.NewReader(msg)

	options := &dkim.SignOptions{
		Domain:   "example.org",
		Selector: "brisbane",
		Signer:   signer,
	}

	var b bytes.Buffer
	if err := dkim.Sign(&b, r, options); err != nil {
		logger.Errorf(err.Error())

		return err
	}

	return nil
}

func (this *SmtpClient) Verify(msg string) error {
	r := strings.NewReader(msg)

	verifications, err := dkim.Verify(r)
	if err != nil {
		logger.Errorf(err.Error())

		return err
	}

	for _, v := range verifications {
		if v.Err == nil {
			logger.Infof("Valid signature for:", v.Domain)
		} else {
			logger.Errorf(v.Err.Error())

			return err
		}
	}

	return nil
}

func ExampleRead() {
	// Let's assume r is an io.Reader that contains a message.
	var r io.Reader

	m, err := message.Read(r)
	if message.IsUnknownCharset(err) {
		// This error is not fatal
		log.Println("Unknown encoding:", err)
	} else if err != nil {
		log.Fatal(err)
	}

	if mr := m.MultipartReader(); mr != nil {
		// This is a multipart message
		log.Println("This is a multipart message containing:")
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}

			t, _, _ := p.Header.ContentType()
			log.Println("A part with type", t)
		}
	} else {
		t, _, _ := m.Header.ContentType()
		log.Println("This is a non-multipart message with type", t)
	}
}

func ExampleWriter() {
	var b bytes.Buffer

	var h message.Header
	h.SetContentType("multipart/alternative", nil)
	w, err := message.CreateWriter(&b, h)
	if err != nil {
		log.Fatal(err)
	}

	var h1 message.Header
	h1.SetContentType("text/html", nil)
	w1, err := w.CreatePart(h1)
	if err != nil {
		log.Fatal(err)
	}
	io.WriteString(w1, "<h1>Hello World!</h1><p>This is an HTML part.</p>")
	w1.Close()

	var h2 message.Header
	h1.SetContentType("text/plain", nil)
	w2, err := w.CreatePart(h2)
	if err != nil {
		log.Fatal(err)
	}
	io.WriteString(w2, "Hello World!\n\nThis is a text part.")
	w2.Close()

	w.Close()

	log.Println(b.String())
}

func Example_transform() {
	// Let's assume r is an io.Reader that contains a message.
	var r io.Reader

	m, err := message.Read(r)
	if message.IsUnknownCharset(err) {
		log.Println("Unknown encoding:", err)
	} else if err != nil {
		log.Fatal(err)
	}

	// We'll add "This message is powered by Go" at the end of each text entity.
	poweredBy := "\n\nThis message is powered by Go."

	var b bytes.Buffer
	w, err := message.CreateWriter(&b, m.Header)
	if err != nil {
		log.Fatal(err)
	}

	// Define a function that transforms message.
	var transform func(w *message.Writer, e *message.Entity) error
	transform = func(w *message.Writer, e *message.Entity) error {
		if mr := e.MultipartReader(); mr != nil {
			// This is a multipart entity, transform each of its parts
			for {
				p, err := mr.NextPart()
				if err == io.EOF {
					break
				} else if err != nil {
					return err
				}

				pw, err := w.CreatePart(p.Header)
				if err != nil {
					return err
				}

				if err := transform(pw, p); err != nil {
					return err
				}

				pw.Close()
			}
			return nil
		} else {
			body := e.Body
			if strings.HasPrefix(m.Header.Get("Content-Type"), "text/") {
				body = io.MultiReader(body, strings.NewReader(poweredBy))
			}
			_, err := io.Copy(w, body)
			return err
		}
	}

	if err := transform(w, m); err != nil {
		log.Fatal(err)
	}
	w.Close()

	log.Println(b.String())
}

func write() {
	var b bytes.Buffer

	from := []*mail.Address{{"Mitsuha Miyamizu", "mitsuha.miyamizu@example.org"}}
	to := []*mail.Address{{"Taki Tachibana", "taki.tachibana@example.org"}}

	// Create our mail header
	var h mail.Header
	h.SetDate(time.Now())
	h.SetAddressList("From", from)
	h.SetAddressList("To", to)

	// Create a new mail writer
	mw, err := mail.CreateWriter(&b, h)
	if err != nil {
		log.Fatal(err)
	}

	// Create a text part
	tw, err := mw.CreateInline()
	if err != nil {
		log.Fatal(err)
	}
	var th mail.InlineHeader
	th.Set("Content-Type", "text/plain")
	w, err := tw.CreatePart(th)
	if err != nil {
		log.Fatal(err)
	}
	io.WriteString(w, "Who are you?")
	w.Close()
	tw.Close()

	// Create an attachment
	var ah mail.AttachmentHeader
	ah.Set("Content-Type", "image/jpeg")
	ah.SetFilename("picture.jpg")
	w, err = mw.CreateAttachment(ah)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: write a JPEG file to w
	w.Close()

	mw.Close()

	log.Println(b.String())
}

func read() {
	// Let's assume r is an io.Reader that contains a mail.
	var r io.Reader

	// Create a new mail reader
	mr, err := mail.CreateReader(r)
	if err != nil {
		log.Fatal(err)
	}

	// Read each mail's part
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			b, _ := ioutil.ReadAll(p.Body)
			log.Printf("Got text: %v\n", string(b))
		case *mail.AttachmentHeader:
			filename, _ := h.Filename()
			log.Printf("Got attachment: %v\n", filename)
		}
	}
}
