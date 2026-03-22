package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"time"

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

	if m.cfg.ResendAPIKey != "" && m.cfg.ResendFrom != "" {
		return m.sendViaResend(to, subject, body)
	}

	if m.cfg.SMTPHost == "" || m.cfg.SMTPFrom == "" {
		log.Printf("approval email not sent; logging instead. to=%s subject=%q body=%q", to, subject, body)
		return DeliveryResult{
			Status: "logged",
			Note:   "No hay proveedor de correo configurado; el mensaje fue escrito en logs.",
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

func (m *Mailer) sendViaResend(to, subject, textBody string) DeliveryResult {
	payload := map[string]any{
		"from":    m.cfg.ResendFrom,
		"to":      []string{to},
		"subject": subject,
		"html":    textToHTML(textBody),
		"text":    textBody,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return DeliveryResult{
			Status: "failed",
			Note:   fmt.Sprintf("No se pudo serializar el correo: %v", err),
		}
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return DeliveryResult{
			Status: "failed",
			Note:   fmt.Sprintf("No se pudo crear la solicitud a Resend: %v", err),
		}
	}
	req.Header.Set("Authorization", "Bearer "+m.cfg.ResendAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("approval email send failed via resend: to=%s err=%v", to, err)
		return DeliveryResult{
			Status: "failed",
			Note:   fmt.Sprintf("No se pudo enviar el correo con Resend: %v", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("approval email send failed via resend: to=%s status=%d", to, resp.StatusCode)
		return DeliveryResult{
			Status: "failed",
			Note:   fmt.Sprintf("Resend respondió con status %d.", resp.StatusCode),
		}
	}

	return DeliveryResult{
		Status: "sent",
		Note:   fmt.Sprintf("Correo enviado a %s vía Resend.", to),
	}
}

func textToHTML(text string) string {
	escaped := strings.ReplaceAll(text, "&", "&amp;")
	escaped = strings.ReplaceAll(escaped, "<", "&lt;")
	escaped = strings.ReplaceAll(escaped, ">", "&gt;")
	escaped = strings.ReplaceAll(escaped, "\n", "<br />")
	return fmt.Sprintf("<p>%s</p>", escaped)
}
