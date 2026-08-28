package notify

import (
	"bufio"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/crisbusta/proindustrial-backend-public/internal/config"
)

// fakeSMTP levanta un servidor SMTP mínimo y devuelve su dirección junto con
// un canal que entrega la transcripción de lo que recibió.
func fakeSMTP(t *testing.T, advertiseSTARTTLS bool) (host, port string, transcript <-chan string) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("no se pudo abrir el listener: %v", err)
	}
	t.Cleanup(func() { ln.Close() })

	out := make(chan string, 1)

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(10 * time.Second))

		var sb strings.Builder
		r := bufio.NewReader(conn)
		w := func(s string) { conn.Write([]byte(s + "\r\n")) }

		w("220 fake ESMTP")
		inData := false
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				break
			}
			sb.WriteString(line)
			cmd := strings.ToUpper(strings.TrimSpace(line))

			if inData {
				if strings.TrimSpace(line) == "." {
					inData = false
					w("250 2.0.0 Ok")
				}
				continue
			}

			switch {
			case strings.HasPrefix(cmd, "EHLO"):
				if advertiseSTARTTLS {
					w("250-fake")
					w("250 STARTTLS")
				} else {
					w("250 fake")
				}
			case strings.HasPrefix(cmd, "HELO"):
				w("250 fake")
			case strings.HasPrefix(cmd, "MAIL FROM"), strings.HasPrefix(cmd, "RCPT TO"):
				w("250 2.1.0 Ok")
			case strings.HasPrefix(cmd, "DATA"):
				inData = true
				w("354 End data with <CR><LF>.<CR><LF>")
			case strings.HasPrefix(cmd, "QUIT"):
				w("221 2.0.0 Bye")
				out <- sb.String()
				return
			default:
				w("250 2.0.0 Ok")
			}
		}
		out <- sb.String()
	}()

	h, p, _ := net.SplitHostPort(ln.Addr().String())
	return h, p, out
}

// Producción usa SMTP en el puerto 587 (Zoho), así que esta es la ruta viva.
func TestSMTPSendCompletesTransaction(t *testing.T) {
	host, port, transcript := fakeSMTP(t, false)

	m := NewMailer(config.Config{
		SMTPHost: host,
		SMTPPort: port,
		SMTPFrom: "admin@puntofusion.cl",
		// Sin SMTPUser: PlainAuth se niega a autenticar sobre una conexión
		// sin cifrar, así que el test cubre el flujo de comandos.
	})

	if err := m.smtpSend("proveedor@empresa.cl", "Asunto de prueba", "Cuerpo"); err != nil {
		t.Fatalf("smtpSend devolvió error: %v", err)
	}

	got := <-transcript
	for _, want := range []string{
		"MAIL FROM:<admin@puntofusion.cl>",
		"RCPT TO:<proveedor@empresa.cl>",
		"DATA",
		"Subject: Asunto de prueba",
		"Content-Type: text/plain; charset=UTF-8",
		"QUIT",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("la transcripción no contiene %q\n--- recibido ---\n%s", want, got)
		}
	}
}

// El asunto de las notificaciones de cotización incluye texto que escribe
// cualquiera desde el formulario público.
func TestSMTPSendStripsHeaderInjection(t *testing.T) {
	host, port, transcript := fakeSMTP(t, false)

	m := NewMailer(config.Config{
		SMTPHost: host,
		SMTPPort: port,
		SMTPFrom: "admin@puntofusion.cl",
	})

	malicious := "Termofusión\r\nBcc: victima@ejemplo.cl"
	if err := m.smtpSend("proveedor@empresa.cl", malicious, "Cuerpo"); err != nil {
		t.Fatalf("smtpSend devolvió error: %v", err)
	}

	got := <-transcript
	// \r y \n se reemplazan por un espacio cada uno, así que el intento de
	// inyección queda como texto plano dentro de la propia línea Subject.
	if !strings.Contains(got, "Subject: Termofusión  Bcc: victima@ejemplo.cl") {
		t.Errorf("el asunto no quedó aplanado en una sola línea\n--- recibido ---\n%s", got)
	}
	// Y la cabecera no debe aparecer nunca al inicio de una línea.
	for _, line := range strings.Split(got, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "Bcc:") {
			t.Errorf("Bcc quedó como cabecera propia: %q", line)
		}
	}
}

func TestSMTPSendTimesOutOnDeadServer(t *testing.T) {
	// Puerto cerrado: debe fallar rápido, no colgarse.
	m := NewMailer(config.Config{
		SMTPHost: "127.0.0.1",
		SMTPPort: "1",
		SMTPFrom: "admin@puntofusion.cl",
	})

	done := make(chan error, 1)
	go func() { done <- m.smtpSend("a@b.cl", "x", "y") }()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("se esperaba un error contra un puerto cerrado")
		}
	case <-time.After(smtpDialTimeout + 5*time.Second):
		t.Fatal("smtpSend se colgó: el timeout de dial no se aplicó")
	}
}

func TestSendRejectionEmailIncludesReason(t *testing.T) {
	host, port, transcript := fakeSMTP(t, false)

	m := NewMailer(config.Config{
		SMTPHost: host,
		SMTPPort: port,
		SMTPFrom: "admin@puntofusion.cl",
	})

	res := m.SendRejectionEmail("proveedor@empresa.cl", "Tuberías del Sur", "No opera en las categorías del directorio.")
	if res.Status != "sent" {
		t.Fatalf("estado inesperado: %q (%s)", res.Status, res.Note)
	}

	got := <-transcript
	for _, want := range []string{
		"Subject: Sobre tu solicitud de registro",
		"Tuberías del Sur",
		"Motivo:",
		"No opera en las categorías del directorio.",
		"contacto@puntofusion.cl",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("el correo no contiene %q\n--- recibido ---\n%s", want, got)
		}
	}
}

func TestSendRejectionEmailOmitsEmptyReason(t *testing.T) {
	host, port, transcript := fakeSMTP(t, false)

	m := NewMailer(config.Config{
		SMTPHost: host,
		SMTPPort: port,
		SMTPFrom: "admin@puntofusion.cl",
	})

	if res := m.SendRejectionEmail("proveedor@empresa.cl", "Tuberías del Sur", "   "); res.Status != "sent" {
		t.Fatalf("estado inesperado: %q (%s)", res.Status, res.Note)
	}

	got := <-transcript
	// Sin motivo no debe quedar un encabezado "Motivo:" colgando vacío.
	if strings.Contains(got, "Motivo:") {
		t.Errorf("se incluyó la sección de motivo estando vacía\n--- recibido ---\n%s", got)
	}
}
