package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
)

type otpEmailData struct {
	OTP          string
	AppName      string
	ValidMinutes int
}

const otpEmailTemplate = `<!DOCTYPE html>
<html lang="id">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
<title>Kode OTP - {{.AppName}}</title>
</head>
<body style="margin:0;padding:0;background:#f4f5f7;font-family:'Segoe UI',Arial,sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="background:#f4f5f7;padding:40px 16px;">
  <tr>
    <td align="center">
      <table width="520" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:12px;overflow:hidden;border:1px solid #e2e5ea;">

        <!-- Header -->
        <tr>
          <td style="background:#0C447C;padding:32px 40px 24px;">
            <table cellpadding="0" cellspacing="0">
              <tr>
                <td>
                  <table cellpadding="0" cellspacing="0">
                    <tr>
                      <td style="background:rgba(255,255,255,0.15);border-radius:8px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                        <span style="color:white;font-size:16px;line-height:32px;">&#128274;</span>
                      </td>
                      <td style="padding-left:10px;color:white;font-size:15px;font-weight:600;letter-spacing:0.02em;">{{.AppName}}</td>
                    </tr>
                  </table>
                </td>
              </tr>
              <tr>
                <td style="padding-top:24px;">
                  <p style="margin:0;color:rgba(255,255,255,0.65);font-size:11px;letter-spacing:0.1em;text-transform:uppercase;">Keamanan Akun</p>
                  <h1 style="margin:4px 0 0;color:#ffffff;font-size:22px;font-weight:600;line-height:1.3;">Kode Verifikasi OTP</h1>
                </td>
              </tr>
            </table>
          </td>
        </tr>

        <!-- Body -->
        <tr>
          <td style="padding:32px 40px;">
            <p style="margin:0 0 24px;font-size:14px;color:#555e6d;line-height:1.75;">
              Halo, kami menerima permintaan untuk mereset password akun kamu.
              Gunakan kode berikut untuk melanjutkan proses verifikasi.
            </p>

            <!-- OTP Box -->
            <table width="100%" cellpadding="0" cellspacing="0" style="background:#f8f9fb;border-radius:12px;border:1px solid #e2e5ea;margin-bottom:24px;">
              <tr>
                <td style="padding:24px;text-align:center;">
                  <p style="margin:0 0 8px;font-size:11px;color:#9aa1ad;letter-spacing:0.1em;text-transform:uppercase;">Kode OTP kamu</p>
                  <p style="margin:0;font-family:'Courier New',monospace;font-size:38px;font-weight:700;letter-spacing:0.25em;color:#0C447C;">{{.OTP}}</p>
                  <p style="margin:12px 0 0;font-size:12px;color:#9aa1ad;">
                    &#128336; Berlaku selama <strong style="color:#555e6d;">{{.ValidMinutes}} menit</strong>
                  </p>
                </td>
              </tr>
            </table>

            <!-- Warning Box -->
            <table width="100%" cellpadding="0" cellspacing="0" style="background:#E6F1FB;border-radius:8px;border:1px solid #B5D4F4;margin-bottom:24px;">
              <tr>
                <td style="padding:12px 16px;font-size:13px;color:#185FA5;line-height:1.65;">
                  &#9432;&nbsp; Jika kamu tidak meminta ini, abaikan email ini. Jangan bagikan kode ini kepada siapapun, termasuk tim support kami.
                </td>
              </tr>
            </table>

            <p style="margin:0;font-size:13px;color:#555e6d;line-height:1.75;">
              Salam,<br>
              <strong style="color:#1a1f27;">Tim {{.AppName}}</strong>
            </p>
          </td>
        </tr>

        <!-- Footer -->
        <tr>
          <td style="border-top:1px solid #e2e5ea;background:#f8f9fb;padding:16px 40px;text-align:center;">
            <p style="margin:0;font-size:11px;color:#9aa1ad;line-height:1.7;">
              Email ini dikirim secara otomatis. Mohon jangan membalas email ini.<br>
              &copy; 2026 {{.AppName}}. Semua hak dilindungi.
            </p>
          </td>
        </tr>

      </table>
    </td>
  </tr>
</table>
</body>
</html>`

func SendOTPEmail(toEmail, otp string) error {
	from := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASS")
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Render HTML template
	tmpl, err := template.New("otp").Parse(otpEmailTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	var body bytes.Buffer
	err = tmpl.Execute(&body, otpEmailData{
		OTP:          otp,
		AppName:      "YBG",
		ValidMinutes: 15,
	})
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Build MIME message with HTML support
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := "Subject: Kode OTP Reset Password YBG\n"
	to := "To: " + toEmail + "\n"
	from_header := "From: YBG <" + from + ">\n"

	message := []byte(subject + to + from_header + mime + body.String())

	auth := smtp.PlainAuth("", from, password, smtpHost)
	return smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{toEmail}, message)
}
