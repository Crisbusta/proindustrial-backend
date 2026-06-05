package notify

import (
	"bytes"
	"crypto/tls"
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
	subject := "Tu empresa fue aprobada en PuntoFusión"
	body := strings.TrimSpace(fmt.Sprintf(`
Hola,

Tu empresa %s fue aprobada en PuntoFusión.

Ya puedes ingresar al panel en:
%s/panel/login

Usuario: %s
Contraseña inicial: %s

Por seguridad, al ingresar deberás cambiar tu contraseña inmediatamente.
`, companyName, strings.TrimRight(m.cfg.AppBaseURL, "/"), to, initialPassword))

	return m.send(to, subject, body)
}

func (m *Mailer) SendQuoteReply(to, requesterName, companyName, service, replyText string) DeliveryResult {
	subject := fmt.Sprintf("Respuesta a tu solicitud: %s", service)
	body := strings.TrimSpace(fmt.Sprintf(`
Hola %s,

%s ha respondido a tu solicitud de cotización para "%s":

%s

Si tienes dudas adicionales puedes responder directamente a este correo.

— El equipo de PuntoFusión
`, requesterName, companyName, service, replyText))

	return m.send(to, subject, body)
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

func (m *Mailer) SendNewQuoteNotification(to, companyName, requesterName, requesterCompany, service, description string) DeliveryResult {
	subject := fmt.Sprintf("Nueva solicitud de cotización — %s", service)
	requesterLine := requesterName
	if requesterCompany != "" {
		requesterLine = fmt.Sprintf("%s (%s)", requesterName, requesterCompany)
	}
	descLine := ""
	if description != "" {
		descLine = fmt.Sprintf("\nDetalle:\n%s\n", description)
	}
	body := strings.TrimSpace(fmt.Sprintf(`
Hola %s,

%s ha solicitado una cotización para "%s".
%s
Ingresa a tu panel para ver el detalle y responder:
%s/panel/solicitudes

— PuntoFusión
`, companyName, requesterLine, service, descLine, strings.TrimRight(m.cfg.AppBaseURL, "/")))

	return m.send(to, subject, body)
}

func (m *Mailer) SendQuoteSubmitConfirmation(to, requesterName, companyName, service string) DeliveryResult {
	subject := fmt.Sprintf("Tu solicitud fue enviada a %s", companyName)
	body := strings.TrimSpace(fmt.Sprintf(`
Hola %s,

Tu solicitud de cotización para "%s" fue enviada exitosamente a %s.

Te contactarán a la brevedad a este correo.

— PuntoFusión
`, requesterName, service, companyName))

	return m.send(to, subject, body)
}

func (m *Mailer) send(to, subject, body string) DeliveryResult {
	if m.cfg.ResendAPIKey != "" && m.cfg.ResendFrom != "" {
		return m.sendViaResend(to, subject, body)
	}
	if m.cfg.SMTPHost == "" || m.cfg.SMTPFrom == "" {
		log.Printf("email not sent (no provider). to=%s subject=%q", to, subject)
		return DeliveryResult{Status: "logged", Note: "No hay proveedor de correo configurado."}
	}
	if err := m.smtpSend(to, subject, body); err != nil {
		log.Printf("email send failed: to=%s err=%v", to, err)
		return DeliveryResult{Status: "failed", Note: fmt.Sprintf("No se pudo enviar: %v", err)}
	}
	return DeliveryResult{Status: "sent", Note: fmt.Sprintf("Correo enviado a %s.", to)}
}

func (m *Mailer) smtpSend(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", m.cfg.SMTPHost, m.cfg.SMTPPort)
	raw := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s\r\n",
		m.cfg.SMTPFrom, to, subject, body))

	if m.cfg.SMTPPort == "465" {
		conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: m.cfg.SMTPHost})
		if err != nil {
			return err
		}
		c, err := smtp.NewClient(conn, m.cfg.SMTPHost)
		if err != nil {
			return err
		}
		defer c.Close()
		if m.cfg.SMTPUser != "" {
			if err := c.Auth(smtp.PlainAuth("", m.cfg.SMTPUser, m.cfg.SMTPPass, m.cfg.SMTPHost)); err != nil {
				return err
			}
		}
		if err := c.Mail(m.cfg.SMTPFrom); err != nil {
			return err
		}
		if err := c.Rcpt(to); err != nil {
			return err
		}
		wc, err := c.Data()
		if err != nil {
			return err
		}
		if _, err = wc.Write(raw); err != nil {
			return err
		}
		if err = wc.Close(); err != nil {
			return err
		}
		return c.Quit()
	}

	var auth smtp.Auth
	if m.cfg.SMTPUser != "" {
		auth = smtp.PlainAuth("", m.cfg.SMTPUser, m.cfg.SMTPPass, m.cfg.SMTPHost)
	}
	return smtp.SendMail(addr, auth, m.cfg.SMTPFrom, []string{to}, raw)
}

func textToHTML(text string) string {
	escaped := strings.ReplaceAll(text, "&", "&amp;")
	escaped = strings.ReplaceAll(escaped, "<", "&lt;")
	escaped = strings.ReplaceAll(escaped, ">", "&gt;")
	escaped = strings.ReplaceAll(escaped, "\n", "<br />")
	return fmt.Sprintf("<p>%s</p>", escaped)
}
