package handler

import (
	"github.com/emersion/go-imap/client"
	"github.com/joho/godotenv"
	"log"
	"os"
	"resume_organizer_go/disk"
	"resume_organizer_go/mail"
)

func getEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("%s не задан", key)
	}
	return v
}

func Start() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка при загрузке .env файла")
	}

	email := getEnv("EMAIL")
	password := getEnv("PASSWORD")
	token := getEnv("AUTHORIZATION_TOKEN")

	imapClient, err := mail.Connect(email, password)
	if err != nil {
		log.Fatalf("Ошибка подключения к почте: %v", err)
	}
	defer func(imapClient *client.Client) {
		err := imapClient.Logout()
		if err != nil {
			log.Fatalf("Ошибка выхода из клиента: %v", err)
		}
	}(imapClient)

	diskSession := disk.NewSession(token)

	if err := mail.ProcessEmails(imapClient, diskSession); err != nil {
		log.Fatalf("Ошибка обработки почты: %v", err)
	}
}
