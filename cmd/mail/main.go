package main

import (
	"flag"
	"html/template"
	"io"
	"log"
	"os"
	"strings"
	"time"

	recruitment "github.com/eternal-flame-AD/email-recruitment"
	"gopkg.in/gomail.v2"
)

var flagEmailTpl = flag.String("t", "template.html", "email template file")
var flagDiverge = flag.String("diverge", "", "diverge all email to another address")
var emailTemplate *template.Template
var mailDialer *gomail.Dialer

func init() {
	flag.Parse()
	recruitment.LoadConfigAndData()
	tplBytes, err := os.ReadFile(*flagEmailTpl)
	if err != nil {
		log.Fatalf("failed to read email template: %v", err)
	}
	emailTemplate = template.Must(template.New("").Funcs(template.FuncMap{
		"set_header": func(msg *gomail.Message, key string, val ...string) string {
			msg.SetHeader(key, val...)
			return ""
		},
		"set_address_header": func(msg *gomail.Message, key, email, name string) string {
			msg.SetAddressHeader(key, email, name)
			return ""
		},
		"date": func(layout string) string {
			return time.Now().Format(layout)
		},
		"first_name": func(name string) string {
			return strings.Split(name, " ")[0]
		},
		"last_name": func(name string) string {
			return name[strings.LastIndexByte(name, ' ')+1:]
		},
	}).Parse(string(tplBytes)))

	mailDialer = gomail.NewDialer(recruitment.Config.SMTP.Host,
		recruitment.Config.SMTP.Port, recruitment.Config.SMTP.Username, recruitment.Config.SMTP.Password)
}

type TemplateContext struct {
	Recruiter recruitment.Recruiter
	Prospect  recruitment.Prospect
	Diverged  bool
	Message   *gomail.Message
}

func main() {
	var mailer gomail.SendCloser
	var err error
	if *flagDiverge == "" || strings.Contains(*flagDiverge, "@") {
		mailer, err = mailDialer.Dial()
		if err != nil {
			log.Panicf("Failed to dial SMTP server: %v", err)
		}
		defer mailer.Close()
	}

	defer recruitment.SaveProspects()

	msg := gomail.NewMessage()
	for i, p := range recruitment.Prospects {
		if p.AlreadySent {
			continue
		}
		log.Printf("sending email to %s", p.Email)

		// look up recruiter
		var recruiter recruitment.Recruiter
		var ok bool
		if recruiter, ok = recruitment.Config.Recruiter[p.AssignedRecruiter]; !ok {
			log.Printf("no recruiter assigned to %s", p.Name)
			continue
		}

		// prepare message context
		msg.Reset()
		ctx := &TemplateContext{
			Recruiter: recruiter,
			Prospect:  p,
			Diverged:  *flagDiverge != "",
			Message:   msg,
		}

		// write headers
		msg.SetAddressHeader("To", p.Email, p.Name)
		msg.SetHeader("From", recruitment.Config.SMTP.From)
		if err := emailTemplate.ExecuteTemplate(io.Discard, "meta", ctx); err != nil {
			log.Printf("failed to execute email template meta section: %v", err)
			continue
		}
		// prepare body
		if err := emailTemplate.Execute(io.Discard, ctx); err != nil /* make sure template runs without errors */ {
			log.Printf("failed to execute email template: %v", err)
			continue
		}
		msg.AddAlternativeWriter("text/html", func(w io.Writer) error {
			return emailTemplate.Execute(w, ctx)
		})

		// send email
		if *flagDiverge != "" && !strings.Contains(*flagDiverge, "@") {
			// diverge to file
			f, err := os.OpenFile(*flagDiverge, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				log.Printf("failed to open diverge file: %v", err)
				continue
			}
			defer f.Close()
			if _, err := msg.WriteTo(f); err != nil {
				log.Printf("failed to write email to diverge file: %v", err)
			}

		} else {
			if *flagDiverge != "" {
				// diverge to email address
				msg.SetHeader("To", *flagDiverge)
				msg.SetAddressHeader("X-Intended-To", p.Email, p.Name)
			}
			if err := gomail.Send(mailer, msg); err != nil {
				log.Printf("failed to send email to %s: %v", p.Name, err)
				continue
			}
		}

		log.Printf("email sent to %s <%s>", p.Name, p.Email)
		if !ctx.Diverged {
			recruitment.Prospects[i].AlreadySent = true
		}

	}
}
