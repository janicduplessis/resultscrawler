// Package crawler is a crawler that runs periodically for each
// user and updates the database if it finds new results. It can also
// warn the user by email when there is new results.
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

	"github.com/janicduplessis/resultscrawler/pkg/logger"
	"github.com/janicduplessis/resultscrawler/pkg/store"
)

const (
	userAgent    = "Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.36 Safari/537.36"
	urlResultats = "https://www-s.websysinfo.uqam.ca/etudiant/drew00da"

	// POST fields
	fieldCode  = "owa_cd_perm"
	fieldNip   = "owa_cpa"
	fieldClass = "owa_sigle"
	fieldGroup = "owa_groupe"
	fieldYear  = "owa_annee"

	// Error/warning messages.
	warningString      = "ATTENTION"
	noResultsString    = "ne sont pas disponibles via"
	noResultsString2   = "valuation n'est diffu"
	invalidClassString = "Session/sigle/groupe inexistant"
	notListedString    = "pas inscrit"
	invalidInfoString  = "Code permanent inexistant ou NIP non valide"
)

var (
	// ErrNoResults happens when the crawler cannot find any results.
	ErrNoResults = errors.New("No results for this class")
	// ErrInvalidGroupClass happens when the group, class or year is invalid.
	ErrInvalidGroupClass = errors.New("Invalid year/class/group")
	// ErrInvalidCodeNip happens when the user code or nip is invalid.
	ErrInvalidCodeNip = errors.New("Invalid code or nip")
	// ErrNotRegistered happens when the user isnt registered for the specified class.
	ErrNotRegistered = errors.New("Not listed for this class")
)

type (
	// Client interface for sending a request.
	Client interface {
		Do(*http.Request) (*http.Response, error)
	}

	// Crawler for getting all grades of a user on resultats uqam
	Crawler struct {
		Client Client
		Logger logger.Logger
	}

	runResult struct {
		ClassIndex int
		Results    []store.Result
		Err        error
	}
)

// NewCrawler creates a new crawler object
func NewCrawler(client Client, logger logger.Logger) *Crawler {
	return &Crawler{
		client,
		logger,
	}
}

// Run returns the results of all classes for the user
func (c *Crawler) Run(user *crawlerUser) []runResult {
	log.Println(fmt.Sprintf("Start looking for results for user %s. User has %v classes.",
		user.Email, len(user.Classes)))

	// Request results
	doneCh := make(chan runResult)
	for i := range user.Classes {
		go c.runClass(user, i, doneCh)
	}

	// Wait for all results to be done
	results := make([]runResult, len(user.Classes))
	for _ = range user.Classes {
		result := <-doneCh
		if result.Err != nil {
			log.Println(result.Err.Error())
		}
		results[result.ClassIndex] = result
	}

	log.Printf("Done looking for results for user %s.\n", user.Email)

	return results
}

func (c *Crawler) runClass(user *crawlerUser, classIndex int, doneCh chan runResult) {
	class := user.Classes[classIndex]
	params := url.Values{
		fieldCode:  {user.Code},
		fieldNip:   {user.Nip},
		fieldClass: {class.Name},
		fieldGroup: {class.Group},
		fieldYear:  {class.Year},
	}
	requestString := fmt.Sprintf("%s?%s", urlResultats, params.Encode())
	req, err := http.NewRequest("POST", requestString, nil)
	if err != nil {
		doneCh <- runResult{
			ClassIndex: classIndex,
			Err:        err,
		}
		return
	}

	req.Header.Set("User-Agent", userAgent)

	log.Printf("Sending request for %s\n", class.Name)
	resp, err := c.Client.Do(req)
	if err != nil {
		doneCh <- runResult{
			ClassIndex: classIndex,
			Err:        err,
		}
		return
	}
	defer resp.Body.Close()

	log.Printf("Parsing response for %s\n", class.Name)
	results, err := parseResponse(resp.Body)
	if err != nil {
		doneCh <- runResult{
			ClassIndex: classIndex,
			Err:        err,
		}
		return
	}
	doneCh <- runResult{
		ClassIndex: classIndex,
		Results:    results,
	}
}

func parseResponse(resp io.Reader) ([]store.Result, error) {
	var results []store.Result
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
			if n.Type == html.TextNode {
				if strings.Contains(n.Data, warningString) {
					log.Println("Found warning")
					hasWarning = true
				}
				// There is 2 different pages for no results. This one has
				// no warning header so we will look for it here.
				if strings.Contains(n.Data, noResultsString2) {
					err = ErrNoResults
					done = true
				}
			}

			// Looking for the results table, it has a 'name' attribute
			// with the value 'form'
			if n.Type == html.ElementNode && n.Data == "table" {
				for _, attr := range n.Attr {
					if attr.Key == "name" && attr.Val == "form" {
						log.Println("Found results table")
						results = parseResultsTable(n)
						done = true
					}
				}
			}
		} else {
			// If there is an error try to find what it is.
			if n.Type == html.TextNode {
				if strings.Contains(n.Data, noResultsString) {
					err = ErrNoResults
					done = true
				} else if strings.Contains(n.Data, invalidClassString) {
					err = ErrInvalidGroupClass
					done = true
				} else if strings.Contains(n.Data, invalidInfoString) {
					err = ErrInvalidCodeNip
					done = true
				} else if strings.Contains(n.Data, notListedString) {
					err = ErrNotRegistered
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

func parseResultsTable(node *html.Node) []store.Result {
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
	results := make([]store.Result, len(rows))
	for i, row := range rows {
		results[i] = parseResultRow(row)
	}
	return results
}

func parseResultRow(node *html.Node) store.Result {
	var cols []*html.Node
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "td" {
			cols = append(cols, c)
		}
	}
	return store.Result{
		Name:    strings.TrimSpace(cols[0].FirstChild.Data),
		Result:  strings.TrimSpace(cols[1].FirstChild.Data),
		Average: strings.TrimSpace(cols[2].FirstChild.Data),
	}
}
