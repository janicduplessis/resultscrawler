package crawler

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"code.google.com/p/go.net/html"

	"github.com/janicduplessis/resultscrawler/lib"
)

const (
	userAgent    = "Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.36 Safari/537.36"
	urlResultats = "https://www-s.websysinfo.uqam.ca/etudiant/drew00da"

	fieldCode  = "owa_cd_perm"
	fieldNip   = "owa_cpa"
	fieldClass = "owa_sigle"
	fieldGroup = "owa_groupe"
	fieldYear  = "owa_annee"

	warningString      = "ATTENTION"
	noResultsString    = "ne sont pas disponibles via"
	invalidClassString = "Session/sigle/groupe inexistant"
	notListedString    = "pas inscrit"
	invalidInfoString  = "Code permanent ou NIP non valide"
)

type runResult struct {
	ClassIndex int
	Results    []lib.Result
	Err        error
}

// Crawler for getting all grades of a user on resultats uqam
type Crawler struct {
	Crypto lib.Crypto
}

// NewCrawler creates a new crawler object
func NewCrawler(crypto lib.Crypto) *Crawler {
	return &Crawler{
		Crypto: crypto,
	}
}

// Run returns the results of all classes for the user
func (c *Crawler) Run(user *lib.User) ([]lib.Class, error) {
	log.Println(fmt.Sprintf("Start looking for results for user %s. User has %v classes.",
		user.UserName, len(user.Classes)))

	// Request results
	doneCh := make(chan runResult)
	for i := range user.Classes {
		go c.runClass(user, i, doneCh)
	}

	// Wait for all results to be done
	classes := make([]lib.Class, len(user.Classes))
	copy(classes, user.Classes)
	for _ = range user.Classes {
		result := <-doneCh
		if result.Err != nil {
			log.Println(result.Err.Error())
		}
		classes[result.ClassIndex].Results = result.Results
	}

	log.Printf("Done looking for results for user %s.\n", user.UserName)

	return classes, nil
}

func (c *Crawler) runClass(user *lib.User, classIndex int, doneCh chan runResult) {
	client := &http.Client{}
	class := user.Classes[classIndex]
	// Decrypt the user code and nip
	data, err := c.Crypto.AESDecrypt(user.Code)
	if err != nil {
		doneCh <- runResult{
			Err: err,
		}
		return
	}
	userCode := string(data)
	data, err = c.Crypto.AESDecrypt(user.Nip)
	if err != nil {
		doneCh <- runResult{
			Err: err,
		}
		return
	}
	userNip := string(data)
	params := url.Values{
		fieldCode:  {userCode},
		fieldNip:   {userNip},
		fieldClass: {class.Name},
		fieldGroup: {class.Group},
		fieldYear:  {class.Year},
	}
	requestString := fmt.Sprintf("%s?%s", urlResultats, params.Encode())
	req, err := http.NewRequest("POST", requestString, nil)
	if err != nil {
		doneCh <- runResult{
			Err: err,
		}
		return
	}

	req.Header.Set("User-Agent", userAgent)

	log.Printf("Sending request for %s\n", class.Name)
	resp, err := client.Do(req)
	if err != nil {
		doneCh <- runResult{
			Err: err,
		}
		return
	}

	log.Printf("Parsing response for %s\n", class.Name)
	results, err := c.parseResponse(resp.Body)
	if err != nil {
		doneCh <- runResult{
			Err: err,
		}
		return
	}
	doneCh <- runResult{
		ClassIndex: classIndex,
		Results:    results,
	}
}

func (c *Crawler) parseResponse(resp io.Reader) ([]lib.Result, error) {
	var results []lib.Result
	doc, err := html.Parse(resp)
	if err != nil {
		return nil, err
	}
	var f func(n *html.Node)
	done := false
	hasWarning := false
	f = func(n *html.Node) {
		// If there is no error yet
		if !hasWarning {
			// Check if there is an error
			if n.Type == html.TextNode && strings.Contains(n.Data, warningString) {
				log.Println("Found warning")
				hasWarning = true
			}
			// Looking for the results table, it has a 'name' attribute
			// with the value 'form'
			if n.Type == html.ElementNode && n.Data == "table" {
				for _, attr := range n.Attr {
					if attr.Key == "name" && attr.Val == "form" {
						log.Println("Found results table")
						results = c.parseResultsTable(n)
						done = true
					}
				}
			}
		} else {
			// If there is an error try to find what it is.
			if n.Type == html.TextNode {
				if strings.Contains(strings.ToUpper(n.Data), strings.ToUpper(noResultsString)) {
					err = errors.New("No results for this class")
					done = true
				} else if strings.Contains(n.Data, invalidClassString) {
					err = errors.New("Invalid year/class/group")
					done = true
				} else if strings.Contains(n.Data, invalidInfoString) {
					err = errors.New("Invalid code or nip")
					done = true
				} else if strings.Contains(n.Data, notListedString) {
					err = errors.New("Not listed for this class")
					done = true
				}
			}
		}
		for c := n.FirstChild; c != nil && !done; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	if !done {
		return nil, errors.New("Unknown error")
	}

	return results, err
}

func (c *Crawler) parseResultsTable(node *html.Node) []lib.Result {
	// Get all rows from the table
	var rows []*html.Node
	tBody := node.FirstChild.NextSibling
	for c := tBody.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "tr" {
			rows = append(rows, c)
		}
	}
	// Remove title rows: first, 2nd and last
	rows = rows[2 : len(rows)-1]
	log.Println(fmt.Sprintf("Found %v results", len(rows)))

	// Parse rows
	results := make([]lib.Result, len(rows))
	for i, row := range rows {
		results[i] = c.parseResultRow(row)
	}
	return results
}

func (c *Crawler) parseResultRow(node *html.Node) lib.Result {
	var cols []*html.Node
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "td" {
			cols = append(cols, c)
		}
	}
	return lib.Result{
		Name:    strings.TrimSpace(cols[0].FirstChild.Data),
		Result:  strings.TrimSpace(cols[1].FirstChild.Data),
		Average: strings.TrimSpace(cols[2].FirstChild.Data),
	}
}
