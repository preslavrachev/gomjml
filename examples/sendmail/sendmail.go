package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"

	"github.com/joho/godotenv"
	"github.com/preslavrachev/gomjml/mjml"
)

func main() {
	// Load .env if present
	_ = godotenv.Load()

	var (
		mjmlFile = "sample.mjml"
		subject  = "MJML Test Email"
		smtpHost = "smtp.gmail.com"
		smtpPort = "587"
		username = os.Getenv("GMAIL_USERNAME")
		password = os.Getenv("GMAIL_PASSWORD")
		to       = os.Getenv("EMAIL_TO")
	)

	if username == "" || password == "" || to == "" {
		log.Fatal("GMAIL_USERNAME, GMAIL_PASSWORD, and EMAIL_TO must be set in .env or envvars")
	}

	// Read MJML file
	mjmlContent, err := ioutil.ReadFile(mjmlFile)
	if err != nil {
		log.Fatalf("Failed to read MJML file: %v", err)
	}

	// Render MJML to HTML
	html, err := mjml.Render(string(mjmlContent))
	if err != nil {
		log.Fatalf("MJML render error: %v", err)
	}

	// Prepare email
	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s",
		username,
		to,
		subject,
		html,
	)

	// Gmail SMTP settings
	auth := smtp.PlainAuth("", username, password, smtpHost)

	// Send email
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, username, []string{to}, []byte(msg))
	if err != nil {
		log.Fatalf("Failed to send email: %v", err)
	}

	fmt.Println("Email sent successfully!")
}
