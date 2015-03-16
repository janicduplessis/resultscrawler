package mobluqam

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/janicduplessis/resultscrawler/pkg/crawler"
)

const (
	urlResultats = "https://mobile.uqam.ca/portail_etudiant/proxy_resultat.php"

	headerUserAgent = "Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.36 Safari/537.36"

	fieldCode  = "code_perm"
	fieldNip   = "nip"
	fieldYear  = "annee"
	fieldClass = "sigle"
	fieldGroup = "groupe"
)

// Crawler for getting all grades of a user using the webservice on mobile.uqam.ca
type Crawler struct {
	Client crawler.ResultGetterClient
}

type resultsResponse struct {
	Normal   [][]string `json:"0"`
	Weighted [][]string `json:"1"`
}

// NewCrawler creates a new crawler object
func NewCrawler() *Crawler {
	return &Crawler{
		&http.Client{},
	}
}

// Run returns the results of all classes for the user
func (c *Crawler) Run(user *crawler.User) []crawler.RunResult {
	log.Println(fmt.Sprintf("Start looking for results for user %s. User has %v classes.",
		user.Email, len(user.Classes)))

	// Request results
	doneCh := make(chan crawler.RunResult)
	for i := range user.Classes {
		go c.runClass(user, i, doneCh)
	}

	// Wait for all results to be done
	results := make([]crawler.RunResult, len(user.Classes))
	for range user.Classes {
		result := <-doneCh
		if result.Err != nil {
			log.Println(result.Err.Error())
		}
		results[result.ClassIndex] = result
	}

	log.Printf("Done looking for results for user %s.\n", user.Email)

	return results
}

func (c *Crawler) runClass(user *crawler.User, classIndex int, doneCh chan crawler.RunResult) {
	class := user.Classes[classIndex]

	params := url.Values{
		fieldCode:  {user.Code},
		fieldNip:   {user.Nip},
		fieldClass: {class.Name},
		fieldYear:  {class.Year},
		fieldGroup: {class.Group},
	}

	req, err := http.NewRequest("POST", urlResultats, strings.NewReader(params.Encode()))
	if err != nil {
		doneCh <- crawler.RunResult{
			ClassIndex: classIndex,
			Err:        err,
		}
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", headerUserAgent)
	req.Header.Add("Origin", "https://mobile.uqam.ca")
	req.Header.Add("Referer", "https://mobile.uqam.ca/portail_etudiant/")

	resp, err := c.Client.Do(req)
	if err != nil {
		doneCh <- crawler.RunResult{
			ClassIndex: classIndex,
			Err:        err,
		}
		return
	}
	defer resp.Body.Close()

	respData, err := ioutil.ReadAll(resp.Body)
	// Remove the while(1); before the actual json.
	respData = respData[9:]
	if err != nil {
		doneCh <- crawler.RunResult{
			ClassIndex: classIndex,
			Err:        err,
		}
		return
	}

	results := &resultsResponse{}
	err = json.Unmarshal(respData, results)
	if err != nil {
		doneCh <- crawler.RunResult{
			ClassIndex: classIndex,
			Err:        err,
		}
		return
	}
	log.Printf("%+v", *results)
}
