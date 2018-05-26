package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ChannelMeter/iso8601duration"
	"github.com/RaghavSood/taskmail/config"
	"github.com/arschles/go-bindata-html-template"
	"gopkg.in/gomail.v2"
	"log"
	"os"
	"runtime"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

type task []struct {
	ID          int          `json:"id"`
	Description string       `json:"description"`
	Due         taskwTime    `json:"due"`
	Entry       taskwTime    `json:"entry"`
	Estimate    taskDuration `json:"estimate"`
	Modified    taskwTime    `json:"modified"`
	Project     string       `json:"project"`
	Status      string       `json:"status"`
	UUID        string       `json:"uuid"`
	Urgency     float64      `json:"urgency"`
}

type taskwTime struct {
	time.Time
}

type taskDuration struct {
	time.Duration
}

type EmailRequest struct {
	from      string
	to        []string
	subject   string
	body      string
	plaintext string
}

const (
	MIME = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
)

func init() {
	if err := config.LoadConfig(UserHomeDir() + "/.taskmail/"); err != nil {
		log.Panic(fmt.Errorf("Invalid application configuration: %s", err))
	}

}

func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func NewRequest(from string, to []string, subject string) *EmailRequest {
	return &EmailRequest{
		from:    from,
		to:      to,
		subject: subject,
	}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var input string
	for scanner.Scan() {
		input += scanner.Text() // Println will add back the final '\n'
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
	var tasklist task
	err := json.Unmarshal([]byte(input), &tasklist)
	if err != nil {
		fmt.Printf("Error: %s", err)
	}

	sort.Slice(tasklist, func(i, j int) bool {
		return tasklist[i].Urgency > tasklist[j].Urgency
	})

	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 0, 1, ' ', 0)
	for i := 0; i < len(tasklist); i++ {
		fmt.Fprintf(w, "%s\t%s\t%s\t%f\n", tasklist[i].Description, tasklist[i].Due.Format("15:04 02-01"), tasklist[i].Estimate.String(), tasklist[i].Urgency)
	}
	w.Flush()
	fmt.Println(buf.String())
	send(buf.String())
}

func send(body string) {
	loc, _ := time.LoadLocation("Pacific/Auckland")
	datenow := time.Now().In(loc).Format("02-01-2006")
	subject := "Daily Report - " + datenow

	r := NewRequest(config.Config.From, config.Config.To, subject)
	r.Send("templates/daily.tmpl", map[string]string{"maildate": datenow, "tasklist": body})
}

func (r *EmailRequest) sendMail() bool {
	m := gomail.NewMessage()
	m.SetHeader("From", "Jay <"+r.from+">")
	m.SetHeader("To", r.to[0])
	m.SetHeader("Subject", r.subject)
	m.SetBody("text/plain", r.plaintext)
	m.AddAlternative("text/html", r.body)

	d := gomail.NewPlainDialer(config.Config.SMTPHost, config.Config.SMTPPort, r.from, config.Config.SMTPPass)
	if err := d.DialAndSend(m); err != nil {
		log.Printf("Error: %s", err)
		return false
	}
	return true
}

func (r *EmailRequest) Send(templatefile string, items interface{}) {
	err := r.parseTemplate(templatefile, items)
	if err != nil {
		log.Fatal(err)
	}
	if ok := r.sendMail(); ok {
		log.Printf("Email has been sent to %s\n", r.to)
	} else {
		log.Printf("Failed to send the email to %s\n", r.to)
	}
}

func (r *EmailRequest) parseTemplate(templatefile string, data interface{}) error {
	t, err := template.New("emailtemplate", Asset).Parse(templatefile)
	if err != nil {
		return err
	}
	buffer := new(bytes.Buffer)
	if err = t.Execute(buffer, data); err != nil {
		return err
	}
	r.body = buffer.String()

	switch v := data.(type) {
	case map[string]string:
		r.plaintext = v["tasklist"]
	}

	return nil
}

func (twTime *taskwTime) UnmarshalJSON(buf []byte) error {
	loc, _ := time.LoadLocation("Pacific/Auckland")
	tt, err := time.Parse("20060102T150405Z", strings.Trim(string(buf), `"`))
	tt = tt.In(loc)
	if err != nil {
		return err
	}
	twTime.Time = tt
	return nil
}

func (twDuration *taskDuration) UnmarshalJSON(buf []byte) error {
	tt, err := duration.FromString(string(buf))
	if err != nil {
		return err
	}
	twDuration.Duration = tt.ToDuration()
	return nil
}
