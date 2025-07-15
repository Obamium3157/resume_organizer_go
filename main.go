package main

import (
	"log"
	"os"
	"resume_organizer_go/disk"
	"resume_organizer_go/mail"

	"github.com/joho/godotenv"
)

func getenv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("%s не задан", key)
	}
	return v
}

func main() {
	godotenv.Load()
	email := getenv("EMAIL")
	password := getenv("PASSWORD")
	token := getenv("AUTHORIZATION_TOKEN")

	imapClient, err := mail.Connect(email, password)
	if err != nil {
		log.Fatalf("Ошибка подключения к почте: %v", err)
	}
	defer imapClient.Logout()

	diskSession := disk.NewSession(token)

	if err := mail.ProcessEmails(imapClient, diskSession); err != nil {
		log.Fatalf("Ошибка обработки почты: %v", err)
	}
}
