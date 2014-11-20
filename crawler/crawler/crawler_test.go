package crawler

import (
	"net/http"
	"os"
	"testing"

	"labix.org/v2/mgo/bson"

	"github.com/janicduplessis/resultscrawler/lib"
)

type FakeClient struct {
	FilePath string
}

type FakeCrypto struct {
}

func (c *FakeClient) Do(req *http.Request) (*http.Response, error) {
	file, err := os.Open(c.FilePath)
	if err != nil {
		return nil, err
	}
	return &http.Response{
		Body: file,
	}, nil
}

func (c *FakeCrypto) AESEncrypt(data []byte) ([]byte, error) {
	return data, nil
}

func (c *FakeCrypto) AESDecrypt(data []byte) ([]byte, error) {
	return data, nil
}

func (c *FakeCrypto) GenerateRandomKey(strength int) []byte {
	return []byte("1234")
}

func TestCrawlerNewResults(t *testing.T) {
	client := &FakeClient{
		"crawler/test/results.html",
	}
	crypto := &FakeCrypto{}
	crawler := NewCrawler(client, crypto)
	results := crawler.Run(getTestUser())
	if len(results) <= 0 {
		t.Error("Found no results. Expected results")
	}
	for _, res := range results {
		if res.Err != nil {
			t.Errorf("Expected no errors. Found error: %s", res.Err)
		}
	}
}

func TestCrawlerNoResults(t *testing.T) {
	client := &FakeClient{
		"crawler/test/no_results.html",
	}
	crypto := &FakeCrypto{}
	crawler := NewCrawler(client, crypto)
	results := crawler.Run(getTestUser())
	if len(results) <= 0 {
		t.Error("Found no results. Expected results")
	}
	for _, res := range results {
		if res.Err != ErrNoResults {
			t.Errorf("Expected ErrNoResults. Found: %s", res.Err)
		}
	}
}

func TestCrawlerErrorNoResults(t *testing.T) {
	client := &FakeClient{
		"crawler/test/no_results.html",
	}
	crypto := &FakeCrypto{}
	crawler := NewCrawler(client, crypto)
	results := crawler.Run(getTestUser())
	if len(results) <= 0 {
		t.Error("Found no results. Expected results")
	}
	for _, res := range results {
		if res.Err != ErrNoResults {
			t.Errorf("Expected ErrNoResults. Found: %s", res.Err)
		}
	}
}

func TestCrawlerErrorInvalidCodeNip(t *testing.T) {
	client := &FakeClient{
		"crawler/test/invalid_code_or_nip.html",
	}
	crypto := &FakeCrypto{}
	crawler := NewCrawler(client, crypto)
	results := crawler.Run(getTestUser())
	if len(results) <= 0 {
		t.Error("Found no results. Expected results")
	}
	for _, res := range results {
		if res.Err != ErrInvalidCodeNip {
			t.Errorf("Expected ErrInvalidCodeNip. Found: %s", res.Err)
		}
	}
}

func TestCrawlerErrorInvalidClassGroup(t *testing.T) {
	client := &FakeClient{
		"crawler/test/invalid_class_or_group.html",
	}
	crypto := &FakeCrypto{}
	crawler := NewCrawler(client, crypto)
	results := crawler.Run(getTestUser())
	if len(results) <= 0 {
		t.Error("Found no results. Expected results")
	}
	for _, res := range results {
		if res.Err != ErrInvalidGroupClass {
			t.Errorf("Expected ErrInvalidGroupClass. Found: %s", res.Err)
		}
	}
}

func TestCrawlerErrorNotRegistered(t *testing.T) {
	client := &FakeClient{
		"crawler/test/not_registered_for_class.html",
	}
	crypto := &FakeCrypto{}
	crawler := NewCrawler(client, crypto)
	results := crawler.Run(getTestUser())
	if len(results) <= 0 {
		t.Error("Found no results. Expected results")
	}
	for _, res := range results {
		if res.Err != ErrNotRegistered {
			t.Errorf("Expected ErrNotRegistered. Found: %s", res.Err)
		}
	}
}

func getTestUser() *lib.User {
	return &lib.User{
		ID:       bson.NewObjectId(),
		UserName: "Test",
		Email:    "test@test.com",
		Code:     []byte("aaaaaa"),
		Nip:      []byte("zzzzzzz"),
		Classes: []lib.Class{
			lib.Class{
				Name:    "Class1",
				Group:   "20",
				Year:    "2014",
				Results: []lib.Result{},
			},
		},
	}
}
