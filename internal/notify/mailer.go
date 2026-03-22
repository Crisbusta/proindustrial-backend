package notify

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"

	"github.com/crisbusta/proindustrial-backend-public/internal/config"
)

type DeliveryResult struct {
	Status string `json:"status"`
	Note   string `json:"note"`
}

type Mailer struct {
	cfg config.Config
}

func NewMailer(cfg config.Config) *Mailer {
	return &Mailer{cfg: cfg}
}

func (m *Mailer) SendApprovalEmail(to, companyName, initialPassword string) DeliveryResult {
	subject := "Tu empresa fue aprobada en ProIndustrial"
	body := strings.TrimSpace(fmt.Sprintf(`
Hola,

Tu empresa %s fue aprobada en ProIndustrial.

Ya puedes ingresar al panel en:
%s/panel/login

Usuario: %s
Contraseña inicial: %s

Por seguridad, al ingresar deberás cambiar tu contraseña inmediatamente.
`, companyName, strings.TrimRight(m.cfg.AppBaseURL, "/"), to, initialPassword))

	if m.cfg.SMTPHost == "" || m.cfg.SMTPFrom == "" {
		log.Printf("approval email not sent via SMTP; logging instead. to=%s subject=%q body=%q", to, subject, body)
		return DeliveryResult{
			Status: "logged",
			Note:   "SMTP no configurado; el mensaje fue escrito en logs.",
		}
	}

	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s\r\n", m.cfg.SMTPFrom, to, subject, body))
	addr := fmt.Sprintf("%s:%s", m.cfg.SMTPHost, m.cfg.SMTPPort)

	var auth smtp.Auth
	if m.cfg.SMTPUser != "" {
		auth = smtp.PlainAuth("", m.cfg.SMTPUser, m.cfg.SMTPPass, m.cfg.SMTPHost)
	}

	if err := smtp.SendMail(addr, auth, m.cfg.SMTPFrom, []string{to}, msg); err != nil {
		log.Printf("approval email send failed: to=%s err=%v", to, err)
		return DeliveryResult{
			Status: "failed",
			Note:   fmt.Sprintf("No se pudo enviar el correo: %v", err),
		}
	}

	return DeliveryResult{
		Status: "sent",
		Note:   fmt.Sprintf("Correo enviado a %s.", to),
	}
}
