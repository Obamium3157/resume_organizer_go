package disk

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"
)

const baseUrl string = "https://cloud-api.yandex.net/v1/disk/"
const appPath string = "app:/"

type Session struct {
	Token string
}

func NewSession(token string) *Session {
	return &Session{Token: token}
}

func FindFile(folder, part string, s *Session) (string, error) {
	path := appPath
	if folder != "" {
		path += folder
	}

	req, _ := http.NewRequest("GET", baseUrl+"resources", nil)
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
	base := baseUrl + "resources"
	req, _ := http.NewRequest("PUT", base, nil)
	req.Header.Set("Authorization", "OAuth "+token)
	q := req.URL.Query()
	q.Add("path", appPath+folder)
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
	base := baseUrl + "resources/move"
	req, _ := http.NewRequest("POST", base, nil)
	req.Header.Set("Authorization", "OAuth "+token)
	q := req.URL.Query()
	q.Add("from", appPath+from)
	q.Add("path", appPath+to)
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
