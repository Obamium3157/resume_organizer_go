package mail

import (
	"fmt"
	"io"
	"log"
	"net/mail"
	"regexp"
	"resume_organizer_go/disk"
	"strconv"
	"strings"

	"github.com/emersion/go-imap"
	imapClient "github.com/emersion/go-imap/client"
)

func Connect(email, password string) (*imapClient.Client, error) {
	c, err := imapClient.DialTLS("imap.yandex.ru:993", nil)
	if err != nil {
		return nil, fmt.Errorf("IMAP dial: %w", err)
	}

	if err := c.Login(email, password); err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}
	log.Println("IMAP подключение успешно")
	return c, nil
}

func ProcessEmails(c *imapClient.Client, diskSession *disk.Session) error {
	_, err := c.Select("INBOX", false)
	if err != nil {
		return fmt.Errorf("выбор INBOX: %w", err)
	}

	seqset := new(imap.SeqSet)
	seqset.Add("1:*")

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchBodyStructure, section.FetchItem()}

	messages := make(chan *imap.Message, 10)
	go func() {
		err := c.Fetch(seqset, items, messages)
		if err != nil {
			log.Fatalf("Fetch error: %v", err)
		}
	}()

	for msg := range messages {
		if msg.Envelope == nil {
			continue
		}

		from := msg.Envelope.From[0].MailboxName + "@" + msg.Envelope.From[0].HostName
		if !strings.Contains(from, "tilda.ws") {
			continue
		}

		body := msg.GetBody(section)
		if body == nil {
			continue
		}

		m, err := mail.ReadMessage(body)
		if err != nil {
			log.Printf("Ошибка чтения MIME: %v", err)
			continue
		}

		bodyBytes, _ := io.ReadAll(m.Body)
		bodyStr := string(bodyBytes)

		fields := ParseFields(bodyStr)
		emailVal := fields["email"]
		if emailVal == "" {
			continue
		}
		year := msg.Envelope.Date.Year()

		path, err := disk.FindFile("", emailVal, diskSession)
		if err != nil || path == "" {
			log.Printf("Файл не найден для: %s", emailVal)
			continue
		}

		yearFolder := strconv.Itoa(year)
		if err := disk.CreateFolder(yearFolder, diskSession.Token); err != nil {
			log.Printf("Не удалось создать папку: %v", err)
		}

		newPath := fmt.Sprintf("%s/%s", yearFolder, path)
		if err := disk.MoveFile(path, newPath, diskSession.Token); err != nil {
			log.Printf("Ошибка при перемещении файла: %v", err)
		} else {
			log.Printf("Файл успешно перемещён: %s -> %s", path, newPath)
		}
	}

	return nil
}

func ParseFields(body string) map[string]string {
	re := regexp.MustCompile(`(?i)(email|name|phone|comments):\s*(.+?)(?:<br>|$)`)
	matches := re.FindAllStringSubmatch(body, -1)
	fields := make(map[string]string)
	for _, m := range matches {
		fields[strings.ToLower(m[1])] = strings.TrimSpace(m[2])
	}
	return fields
}
