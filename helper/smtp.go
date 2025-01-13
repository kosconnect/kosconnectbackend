package helper

import (
    "net/smtp"
	"time"

	// "gopkg.in/gomail.v2"
	// "fmt"
)

func SendVerificationEmail(email, verificationLink, fullName string) error {
    // Template body email
    body := `
    <!DOCTYPE html>
    <html>
    <head>
        <style>
            body {
                font-family: Arial, sans-serif;
                line-height: 1.6;
                color: #333333;
            }
            .header {
                text-align: center;
                padding: 20px;
                background-color: #f8f9fa;
                border-bottom: 1px solid #e0e0e0;
            }
            .header img {
                max-height: 50px;
                vertical-align: middle;
            }
            .header .title {
                display: inline-block;
                font-family: 'Merriweather', serif;
                font-size: 24px;
                color: #005f73;
                vertical-align: middle;
                margin-left: 10px;
            }
            .content {
                padding: 20px;
            }
            .button {
                display: inline-block;
                padding: 10px 20px;
                color: white;
                background-color: #007BFF;
                text-decoration: none;
                border-radius: 5px;
                font-size: 16px;
            }
            .footer {
                margin-top: 20px;
                font-size: 12px;
                color: #666666;
                text-align: center;
            }
        </style>
    </head>
    <body>
        <div class="header">
            <img src="https://kosconnect-server.vercel.app/images/logokos.png" alt="KosConnect Logo">
            <span class="title">KosConnect</span>
        </div>
        <div class="content">
            <h2>Halo, ` + fullName + `</h2>
            <p>Terima kasih telah mendaftar di KosConnect!</p>
            <p>Untuk menyelesaikan pendaftaran Anda, silakan verifikasi alamat email Anda dengan mengklik tombol di bawah ini:</p>
            <a href="` + verificationLink + `" class="button">
                Verifikasi Email
            </a>
            <p>Jika Anda tidak meminta ini, silakan abaikan email ini.</p>
            <p>Terima kasih,<br>Tim KosConnect</p>
        </div>
        <div class="footer">
            &copy; ` + time.Now().Format("2006") + ` KosConnect. Semua Hak Dilindungi.
        </div>
    </body>
    </html>
    `

    // Konfigurasi SMTP
    auth := smtp.PlainAuth("", "kosconnect2@gmail.com", "email_password", "smtp.gmail.com")
    to := []string{email}

    // Header email
    subject := "Subject: Verifikasi Email Anda di KosConnect\r\n"
    contentType := "MIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n"
    msg := []byte(subject + contentType + "\r\n" + body)

    // Kirim email
    return smtp.SendMail("smtp.gmail.com:587", auth, "kosconnect2@gmail.com", to, msg)
}



// func SendBookingNotification(ownerEmail, boardingHouseName, customerName, bookingDetails string) error {
// 	mailer := gomail.NewMessage()
// 	mailer.SetHeader("From", "kosconnect2@gmail.com")
// 	mailer.SetHeader("To", ownerEmail)
// 	mailer.SetHeader("Subject", fmt.Sprintf("Booking Baru di %s", boardingHouseName))
// 	mailer.SetBody("text/plain", fmt.Sprintf(
// 		"Halo,\n\nAda booking baru di kos \"%s\" oleh pelanggan \"%s\".\n\nDetail Booking:\n%s\n",
// 		boardingHouseName, customerName, bookingDetails,
// 	))

// 	dialer := gomail.NewDialer("smtp.example.com", 587, "kosconnect2@example.com", "your-password")
// 	if err := dialer.DialAndSend(mailer); err != nil {
// 		fmt.Printf("Gagal mengirim email: %v", err)
// 		return err
// 	}

// 	fmt.Println("Email berhasil dikirim!")
// 	return nil
// }
