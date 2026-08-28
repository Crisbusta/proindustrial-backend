package notify

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
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

// SendRejectionEmail avisa al proveedor que su solicitud no fue aprobada.
// El motivo es opcional: si el admin no escribió ninguno, el correo sale sin
// esa sección en vez de dejar un hueco vacío.
func (m *Mailer) SendRejectionEmail(to, companyName, reason string) DeliveryResult {
	subject := "Sobre tu solicitud de registro en PuntoFusión"

	reasonBlock := ""
	if r := strings.TrimSpace(reason); r != "" {
		reasonBlock = fmt.Sprintf("\n\nMotivo:\n%s", r)
	}

	body := strings.TrimSpace(fmt.Sprintf(`
Hola,

Revisamos la solicitud de registro de %s en PuntoFusión y, por ahora, no pudimos aprobarla.%s

Si crees que se trata de un error, o quieres corregir la información y volver a postular, escríbenos a contacto@puntofusion.cl y lo revisamos contigo.

Gracias por tu interés.

— El equipo de PuntoFusión
`, companyName, reasonBlock))

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

// sanitizeHeader evita la inyección de cabeceras: parte del asunto viene
// de campos que rellena el público (por ejemplo el servicio de una
// cotización), y un \r\n permitiría añadir un Bcc arbitrario.
func sanitizeHeader(s string) string {
	return strings.NewReplacer("\r", " ", "\n", " ").Replace(s)
}

// smtpDialTimeout acota la conexión SMTP. Sin él, un servidor que acepta
// la conexión y se queda colgado dejaba el request de aprobación pendiente
// indefinidamente, después de que la empresa ya se había creado.
const smtpDialTimeout = 10 * time.Second

func (m *Mailer) smtpSend(to, subject, body string) error {
	addr := net.JoinHostPort(m.cfg.SMTPHost, m.cfg.SMTPPort)
	raw := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n",
		sanitizeHeader(m.cfg.SMTPFrom), sanitizeHeader(to), sanitizeHeader(subject), body))

	if m.cfg.SMTPPort == "465" {
		dialer := &net.Dialer{Timeout: smtpDialTimeout}
		conn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: m.cfg.SMTPHost})
		if err != nil {
			return err
		}
		_ = conn.SetDeadline(time.Now().Add(30 * time.Second))
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

	// smtp.SendMail no acepta timeout, así que se abre la conexión a mano.
	conn, err := net.DialTimeout("tcp", addr, smtpDialTimeout)
	if err != nil {
		return err
	}
	_ = conn.SetDeadline(time.Now().Add(30 * time.Second))

	c, err := smtp.NewClient(conn, m.cfg.SMTPHost)
	if err != nil {
		return err
	}
	defer c.Close()

	if ok, _ := c.Extension("STARTTLS"); ok {
		if err := c.StartTLS(&tls.Config{ServerName: m.cfg.SMTPHost}); err != nil {
			return err
		}
	}
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
	if _, err := wc.Write(raw); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	return c.Quit()
}

func textToHTML(text string) string {
	escaped := strings.ReplaceAll(text, "&", "&amp;")
	escaped = strings.ReplaceAll(escaped, "<", "&lt;")
	escaped = strings.ReplaceAll(escaped, ">", "&gt;")
	escaped = strings.ReplaceAll(escaped, "\n", "<br />")
	return fmt.Sprintf("<p>%s</p>", escaped)
}
