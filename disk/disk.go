package disk

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"
)

type Session struct {
	Token string
}

func NewSession(token string) *Session {
	return &Session{Token: token}
}

func FindFile(folder, part string, s *Session) (string, error) {
	path := "app:/"
	if folder != "" {
		path += folder
	}

	req, _ := http.NewRequest("GET", "https://cloud-api.yandex.net/v1/disk/resources", nil)
	req.Header.Set("Authorization", "OAuth "+s.Token)
	q := req.URL.Query()
	q.Add("path", path)
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка запроса к Яндекс.Диску: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	items := gjson.GetBytes(body, "_embedded.items").Array()
	for _, item := range items {
		if item.Get("type").String() == "file" && strings.Contains(item.Get("name").String(), part) {
			return extractFilename(item.Get("path").String()), nil
		}
	}
	return "", nil
}

func extractFilename(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

func CreateFolder(folder, token string) error {
	base := "https://cloud-api.yandex.net/v1/disk/resources"
	req, _ := http.NewRequest("PUT", base, nil)
	req.Header.Set("Authorization", "OAuth "+token)
	q := req.URL.Query()
	q.Add("path", "app:/"+folder)
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 201 || resp.StatusCode == 409 {
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("create folder failed: %s", string(body))
}

func MoveFile(from, to, token string) error {
	base := "https://cloud-api.yandex.net/v1/disk/resources/move"
	req, _ := http.NewRequest("POST", base, nil)
	req.Header.Set("Authorization", "OAuth "+token)
	q := req.URL.Query()
	q.Add("from", "app:/"+from)
	q.Add("path", "app:/"+to)
	q.Add("overwrite", "true")
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("move failed: %s", string(body))
}
