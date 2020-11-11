package app

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/OB1Company/filehive/repo"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
)

var (
	jpgTestImage  = "/9j/4AAQSkZJRgABAQAAAQABAAD//gA7Q1JFQVRPUjogZ2QtanBlZyB2MS4wICh1c2luZyBJSkcgSlBFRyB2NjIpLCBxdWFsaXR5ID0gNjUK/9sAQwALCAgKCAcLCgkKDQwLDREcEhEPDxEiGRoUHCkkKyooJCcnLTJANy0wPTAnJzhMOT1DRUhJSCs2T1VORlRAR0hF/9sAQwEMDQ0RDxEhEhIhRS4nLkVFRUVFRUVFRUVFRUVFRUVFRUVFRUVFRUVFRUVFRUVFRUVFRUVFRUVFRUVFRUVFRUVF/8AAEQgAMgAyAwEiAAIRAQMRAf/EAB8AAAEFAQEBAQEBAAAAAAAAAAABAgMEBQYHCAkKC//EALUQAAIBAwMCBAMFBQQEAAABfQECAwAEEQUSITFBBhNRYQcicRQygZGhCCNCscEVUtHwJDNicoIJChYXGBkaJSYnKCkqNDU2Nzg5OkNERUZHSElKU1RVVldYWVpjZGVmZ2hpanN0dXZ3eHl6g4SFhoeIiYqSk5SVlpeYmZqio6Slpqeoqaqys7S1tre4ubrCw8TFxsfIycrS09TV1tfY2drh4uPk5ebn6Onq8fLz9PX29/j5+v/EAB8BAAMBAQEBAQEBAQEAAAAAAAABAgMEBQYHCAkKC//EALURAAIBAgQEAwQHBQQEAAECdwABAgMRBAUhMQYSQVEHYXETIjKBCBRCkaGxwQkjM1LwFWJy0QoWJDThJfEXGBkaJicoKSo1Njc4OTpDREVGR0hJSlNUVVZXWFlaY2RlZmdoaWpzdHV2d3h5eoKDhIWGh4iJipKTlJWWl5iZmqKjpKWmp6ipqrKztLW2t7i5usLDxMXGx8jJytLT1NXW19jZ2uLj5OXm5+jp6vLz9PX29/j5+v/aAAwDAQACEQMRAD8A840awhv5zFKWDYyMHrVvWtE/szynj3GJ+MnsaoWFw1ndxTr1Rskeor0+70uPXNBYQ4JkQSRH36iiXw3CO9meWxxNJIqICWY4AHeu5g8C232aMztL5pUFtpGM/lUXgPw+13qD3lwhEdscAEdX/wDrVseNddl0l4bSxcLcN8zHAOB6c1UnyJLqxRTlJ9kY83guzQcNN/30P8KwNY0W206AvufceFBPWvRtMtrw6RHLqUm+dxvOVA2j04rzjxJqAv8AUXEZ/cxHavv71M20+UcbNc3Q5/bRUu2igCVRXpfw51MXFtJp0rfPD88ee6nr+R/nXmq13fw40xpL6TUXyEiGxfcnrVwV7kSdrHo7C10ixnn2rFEu6R8cZPU15r4espfFviua/uQTbxvvbPT/AGVrX+IetMyQ6PakmSUhpAv6Cuh0DT4PC/hsGbCsE82ZvfHSs4O160umiLmtFTW73/rzMjx7rC6Zp32WFgJ7gY4/hXua8nbmtTXtWk1rVZruQnDHCL/dXtWW1TBPd7suVl7q6EeKKKKsgkt42nlSNBlmIAr1rTZINA0MAkBYU3MfU15joEsEN5588irs+6GPetfX9cW8iis7eVSjHLsDxRJ+7yx3YRV5XeyNjwlbPrviCbWL0bkjbcoPQt2H4Vf+IOsTyxpplrHIyt80rKpwfQU3R9W0vS9PitkvIBtHzHeOT3q+3ibTiP8Aj9g/77FE+R2itkEXK7m92eXm2uB1gk/74NRPDKoOY3H1U16RceI7FgcXkJ/4GKwdT1q2lgkVJ0YlSOGoco20CzONzRUe6igBq1KtFFAD6KKKAGmo26UUUAR0UUUAf//Z"
	jpgImageBytes []byte
)

func init() {
	jpgImageBytes, _ = base64.StdEncoding.DecodeString(`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`)
}

type apiTests []apiTest

type apiTest struct {
	name             string
	path             string
	method           string
	body             []byte
	statusCode       int
	setup            func(db *repo.Database) error
	expectedResponse []byte
}

func errorReturn(err error) []byte {
	return []byte(fmt.Sprintf(`{"error": "%s"}%s`, err.Error(), "\n"))
}

func runAPITests(t *testing.T, tests apiTests) {
	db, err := repo.NewDatabase("", repo.Dialect("memory"))
	if err != nil {
		t.Fatal(err)
	}

	staticDir := path.Join(os.TempDir(), "filehive_apiTests", "www")
	if err := os.MkdirAll(path.Join(staticDir, "images"), os.ModePerm); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(staticDir)

	server := &FileHiveServer{
		db:            db,
		staticFileDir: staticDir,
	}

	r := server.newV1Router()
	ts := httptest.NewServer(r)
	defer ts.Close()

	var cookies []*http.Cookie
	for _, test := range tests {
		if test.setup != nil {
			if err := test.setup(db); err != nil {
				t.Fatalf("%s: %s", test.name, err)
			}
		}

		req, err := http.NewRequest(test.method, fmt.Sprintf("%s%s", ts.URL, test.path), bytes.NewReader(test.body))
		if err != nil {
			t.Fatal(err)
		}
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != test.statusCode {
			t.Errorf("%s: Expected status code %d, got %d", test.name, test.statusCode, res.StatusCode)
			continue
		}
		if len(res.Cookies()) > 0 {
			cookies = res.Cookies()
		}

		response, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			log.Fatal(err)
		}

		if test.expectedResponse != nil && !bytes.Equal(response, test.expectedResponse) {
			t.Errorf("%s: Expected response %s, got %s", test.name, string(test.expectedResponse), string(response))
			continue
		}
	}
}
