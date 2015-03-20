package resuqam

import (
	"io"
	"net/http"
	"os"
	"testing"

	"labix.org/v2/mgo/bson"

	"github.com/janicduplessis/resultscrawler/pkg/api"
	"github.com/janicduplessis/resultscrawler/pkg/crawler"
)

type FakeClient struct {
	Data io.ReadCloser
}

func (c *FakeClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		Body: c.Data,
	}, nil
}

func TestCrawlerNewResults(t *testing.T) {
	crawler := getCrawler(t, "test/results.html")
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
	crawler := getCrawler(t, "test/no_results.html")
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
	crawler := getCrawler(t, "test/invalid_code_or_nip.html")
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
	crawler := getCrawler(t, "test/invalid_class_or_group.html")
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
	crawler := getCrawler(t, "test/not_registered_for_class.html")
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

func getTestUser() *crawler.User {
	return &crawler.User{
		ID:    bson.NewObjectId().Hex(),
		Code:  "aaaaaa",
		Nip:   "zzzzzzz",
		Email: "test@test.com",
		Classes: []api.Class{
			api.Class{
				Name:    "Class1",
				Group:   "20",
				Year:    "2014",
				Results: []api.Result{},
			},
		},
	}
}

func getCrawler(t *testing.T, fileToCrawl string) *Crawler {
	data, err := os.Open(fileToCrawl)
	if err != nil {
		t.Errorf("Error opening test file %s. Err: %s", fileToCrawl, err)
	}
	client := &FakeClient{
		Data: data,
	}
	crawler := NewCrawler()
	crawler.Client = client
	return crawler
}

func init() {
	// Working directory is different in test so we have to fix the path of
	// the template file.
	crawler.MsgTemplatePath = "../../../crawler/msgtemplate.html"
}
