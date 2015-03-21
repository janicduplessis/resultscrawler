package mobluqam

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/janicduplessis/resultscrawler/pkg/api"
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
	log.Printf("Sending request for %s\n", class.Name)
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
	log.Printf("Parsing response for %s\n", class.Name)
	resultsResponse := &resultsResponse{}
	err = json.Unmarshal(respData, resultsResponse)
	if err != nil {
		doneCh <- crawler.RunResult{
			ClassIndex: classIndex,
			Err:        err,
		}
		return
	}
	results := parseResponse(resultsResponse)

	doneCh <- crawler.RunResult{
		ClassIndex: classIndex,
		Class:      results,
	}
}

type resultIndexes struct {
	Result       int
	Average      int
	StandardDev  int
	WResult      int
	WAverage     int
	WStandardDev int
}

func parseResponse(response *resultsResponse) *api.Class {
	class := &api.Class{}

	headersRow := response.Normal[0]
	hasWeighted := len(response.Weighted) > 0

	var indexes *resultIndexes
	switch len(headersRow) {
	case 2:
		if hasWeighted {
			indexes = &resultIndexes{1, -1, -1, 1, -1, -1}
		} else {
			indexes = &resultIndexes{1, -1, -1, -1, -1, -1}
		}
	case 3:
		if hasWeighted {
			indexes = &resultIndexes{1, 2, -1, 1, 2, -1}
		} else {
			indexes = &resultIndexes{1, -1, -1, 2, -1, -1}
		}
	case 4:
		if hasWeighted {
			indexes = &resultIndexes{1, 2, 3, 1, 2, 3}
		} else {
			indexes = &resultIndexes{1, 2, 3, -1, -1, -1}
		}
	default:
		log.Println("Invalid layout")
		return class
	}

	nResults := response.Normal[1:]
	var wResults [][]string
	if hasWeighted {
		wResults = response.Weighted[1:]
	} else {
		wResults = nResults
	}

	hasFinal := wResults[len(wResults)-1][0] == "Note:"

	var totalRow []string
	if hasFinal {
		totalRow = wResults[len(wResults)-2]
	} else {
		totalRow = wResults[len(wResults)-1]
	}

	class.Total = api.ResultInfo{
		Result:      resAt(indexes.Result, totalRow),
		Average:     resAt(indexes.Average, totalRow),
		StandardDev: resAt(indexes.StandardDev, totalRow),
	}

	// If we have the final grade row it will be after total in weighted results.
	if hasFinal {
		finalRow := wResults[len(wResults)-1]
		class.Final = finalRow[1]
		wResults = wResults[:len(wResults)-2]
	} else {
		wResults = wResults[:len(wResults)-1]
	}

	if !hasWeighted {
		nResults = wResults
	}

	results := make([]api.Result, len(nResults))
	for i := range nResults {
		nRes := nResults[i]
		wRes := wResults[i]
		results[i] = api.Result{
			Name: nRes[0],
			Normal: api.ResultInfo{
				Result:      resAt(indexes.Result, nRes),
				Average:     resAt(indexes.Average, nRes),
				StandardDev: resAt(indexes.StandardDev, nRes),
			},
			Weighted: api.ResultInfo{
				Result:      resAt(indexes.WResult, wRes),
				Average:     resAt(indexes.WAverage, wRes),
				StandardDev: resAt(indexes.WStandardDev, wRes),
			},
		}
	}

	class.Results = results
	return class
}

func resAt(index int, cols []string) string {
	if index == -1 || len(cols[index]) == 0 {
		return "N/A"
	}
	return cols[index]
}
