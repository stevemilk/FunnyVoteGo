package email

import (
	"hyperbaas/src/util"
	"strings"

	"bytes"
	"html/template"
	"hyperbaas/src/api/vm"
	"io/ioutil"

	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

// SendMail send email
func SendMail(subject, content string, to, cc []string) error {
	toList := make([]string, 0, len(to))
	ccList := make([]string, 0, len(cc))

	for _, v := range to {
		v = strings.TrimSpace(v)
		if util.MatchEmail([]byte(v)) {
			exists := false
			for _, vv := range toList {
				if v == vv {
					exists = true
					break
				}
			}
			if !exists {
				toList = append(toList, v)
			}
		}
	}
	for _, v := range cc {
		v = strings.TrimSpace(v)
		if util.MatchEmail([]byte(v)) {
			exists := false
			for _, vv := range ccList {
				if v == vv {
					exists = true
					break
				}
			}
			if !exists {
				ccList = append(ccList, v)
			}
		}
	}

	m := gomail.NewMessage()
	m.SetHeader("From", viper.GetString("email.from"))
	m.SetHeader("To", toList...)
	if len(ccList) > 0 {
		m.SetHeader("Cc", ccList...)
	}
	m.SetHeader("Subject", subject)
	//content = confirm
	// TODO: make content from template `confirm.html` with param content
	m.SetBody("text/html", content)

	d := gomail.NewPlainDialer(viper.GetString("email.host"), viper.GetInt("email.port"), viper.GetString("email.user"), viper.GetString("email.password"))

	return d.DialAndSend(m)
}

//SendInviteMail send invite mail
func SendInviteMail(subject string, content vm.ResInviteToken, to, cc []string) error {
	toList := make([]string, 0, len(to))
	ccList := make([]string, 0, len(cc))

	for _, v := range to {
		v = strings.TrimSpace(v)
		if util.MatchEmail([]byte(v)) {
			exists := false
			for _, vv := range toList {
				if v == vv {
					exists = true
					break
				}
			}
			if !exists {
				toList = append(toList, v)
			}
		}
	}
	for _, v := range cc {
		v = strings.TrimSpace(v)
		if util.MatchEmail([]byte(v)) {
			exists := false
			for _, vv := range ccList {
				if v == vv {
					exists = true
					break
				}
			}
			if !exists {
				ccList = append(ccList, v)
			}
		}
	}

	m := gomail.NewMessage()
	m.SetAddressHeader("From", viper.GetString("email.from"), viper.GetString("email.sender"))
	m.SetHeader("To", toList...)
	if len(ccList) > 0 {
		m.SetHeader("Cc", ccList...)
	}
	m.SetHeader("Subject", subject)

	Email, err := ioutil.ReadFile("./conf/email/invite_email.html")
	if err != nil {
		return err
	}
	t, err := template.New("Invite Email").Parse(string(Email))
	if err != nil {
		return err
	}
	buffer := new(bytes.Buffer)
	t.Execute(buffer, content)
	m.SetBody("text/html", buffer.String())

	d := gomail.NewDialer(viper.GetString("email.host"), viper.GetInt("email.port"), viper.GetString("email.user"), viper.GetString("email.password"))

	return d.DialAndSend(m)
}

// SendActivateMail send activate mail
func SendActivateMail(subject string, content vm.ResActivate, to, cc []string) error {
	toList := make([]string, 0, len(to))
	ccList := make([]string, 0, len(cc))

	for _, v := range to {
		v = strings.TrimSpace(v)
		if util.MatchEmail([]byte(v)) {
			exists := false
			for _, vv := range toList {
				if v == vv {
					exists = true
					break
				}
			}
			if !exists {
				toList = append(toList, v)
			}
		}
	}
	for _, v := range cc {
		v = strings.TrimSpace(v)
		if util.MatchEmail([]byte(v)) {
			exists := false
			for _, vv := range ccList {
				if v == vv {
					exists = true
					break
				}
			}
			if !exists {
				ccList = append(ccList, v)
			}
		}
	}

	m := gomail.NewMessage()
	m.SetAddressHeader("From", viper.GetString("email.from"), viper.GetString("email.sender"))
	m.SetHeader("To", toList...)
	if len(ccList) > 0 {
		m.SetHeader("Cc", ccList...)
	}
	m.SetHeader("Subject", subject)

	Email, err := ioutil.ReadFile("./conf/email/activate_mail.html")
	if err != nil {
		return err
	}
	t, err := template.New("Activate Email").Parse(string(Email))
	if err != nil {
		return err
	}
	buffer := new(bytes.Buffer)
	t.Execute(buffer, content)
	m.SetBody("text/html", buffer.String())

	d := gomail.NewDialer(viper.GetString("email.host"), viper.GetInt("email.port"), viper.GetString("email.user"), viper.GetString("email.password"))

	return d.DialAndSend(m)
}
