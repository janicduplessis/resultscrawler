package crawler

import (
	"os"
	"testing"

	"labix.org/v2/mgo/bson"

	"github.com/janicduplessis/resultscrawler/crawler/crawler/test"
	"github.com/janicduplessis/resultscrawler/lib"
)

func TestCrawlerNewResults(t *testing.T) {
	data, err := os.Open("test/results.html")
	if err != nil {
		t.Errorf("Error opening test file. Err: %s", err)
	}
	client := &test.FakeClient{
		Data: data,
	}
	crypto := &test.FakeCrypto{}
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

func TestCrawlerErrorNoResults(t *testing.T) {
	data, err := os.Open("test/no_results.html")
	if err != nil {
		t.Errorf("Error opening test file. Err: %s", err)
	}
	client := &test.FakeClient{
		Data: data,
	}
	crypto := &test.FakeCrypto{}
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
	data, err := os.Open("test/invalid_code_or_nip.html")
	if err != nil {
		t.Errorf("Error opening test file. Err: %s", err)
	}
	client := &test.FakeClient{
		Data: data,
	}
	crypto := &test.FakeCrypto{}
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
	data, err := os.Open("test/invalid_class_or_group.html")
	if err != nil {
		t.Errorf("Error opening test file. Err: %s", err)
	}
	client := &test.FakeClient{
		Data: data,
	}
	crypto := &test.FakeCrypto{}
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
	data, err := os.Open("test/not_registered_for_class.html")
	if err != nil {
		t.Errorf("Error opening test file. Err: %s", err)
	}
	client := &test.FakeClient{
		Data: data,
	}
	crypto := &test.FakeCrypto{}
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

func init() {
	// Working directory is different in test so we have to fix the path of
	// the template file.
	msgTemplatePath = "msgtemplate.html"
}
