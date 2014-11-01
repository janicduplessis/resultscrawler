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
	fieldCode    = "owa_cd_perm"
	fieldNip     = "owa_cpa"
	fieldClass   = "owa_sigle"
	fieldGroup   = "owa_groupe"
	fieldYear    = "owa_annee"
)

type Crawler struct {
}

func (c *Crawler) Run(user *lib.User) error {
	log.Println("Start looking for results")

	doneCh := make(chan bool)
	for _, class := range user.Classes {
		c.runClass(user, &class, doneCh)
	}

	for i := 0; i < len(user.Classes); i++ {
		<-doneCh
	}

	log.Println("done")

	return nil
}

func (c *Crawler) runClass(user *lib.User, class *lib.Class, doneCh chan bool) {
	client := &http.Client{}
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
		doneCh <- false
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		doneCh <- false
	}

	_, err = c.parseResponse(resp.Body)
	if err != nil {
		doneCh <- false
	}

	doneCh <- true
}

func (c *Crawler) parseResponse(resp io.Reader) ([]lib.Result, error) {
	var results []lib.Result
	doc, err := html.Parse(resp)
	if err != nil {
		return nil, err
	}
	var f func(n *html.Node)
	done := false
	f = func(n *html.Node) {
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
		for c := n.FirstChild; c != nil && !done; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	if !done {
		log.Println("Could not find any results")
		return nil, errors.New("No results")
	}

	for i, res := range results {
		log.Println(fmt.Sprintf("Result %v", i))
		log.Println(fmt.Sprintf("  Name: %v", res.Name))
		log.Println(fmt.Sprintf("  Result: %v", res.Result))
		log.Println(fmt.Sprintf("  Average: %v", res.Average))
	}

	return results, nil
}

func (c *Crawler) parseResultsTable(node *html.Node) []lib.Result {
	// Get all rows from the table
	rows := make([]*html.Node, 0)
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
	cols := make([]*html.Node, 0)
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
