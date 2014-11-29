package crawler

import (
	"os"
	"testing"

	"labix.org/v2/mgo/bson"

	"github.com/janicduplessis/resultscrawler/pkg/crawler/test"
	"github.com/janicduplessis/resultscrawler/pkg/logger"
	"github.com/janicduplessis/resultscrawler/pkg/store"
)

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

func getTestUser() *crawlerUser {
	return &crawlerUser{
		ID:    bson.NewObjectId(),
		Code:  "aaaaaa",
		Nip:   "zzzzzzz",
		Email: "test@test.com",
		Classes: []store.Class{
			store.Class{
				Name:    "Class1",
				Group:   "20",
				Year:    "2014",
				Results: []store.Result{},
			},
		},
	}
}

func getCrawler(t *testing.T, fileToCrawl string) *Crawler {
	data, err := os.Open(fileToCrawl)
	if err != nil {
		t.Errorf("Error opening test file %s. Err: %s", fileToCrawl, err)
	}
	client := &test.FakeClient{
		Data: data,
	}
	logger := &logger.ConsoleLogger{}
	return NewCrawler(client, logger)
}

func init() {
	// Working directory is different in test so we have to fix the path of
	// the template file.
	msgTemplatePath = "../../crawler/msgtemplate.html"
}
