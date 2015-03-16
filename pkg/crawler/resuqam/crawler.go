package resuqam

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/janicduplessis/resultscrawler/pkg/api"
	"github.com/janicduplessis/resultscrawler/pkg/crawler"

	"code.google.com/p/go.net/html"
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

// Crawler for getting all grades of a user on Resultats UQAM website.
type Crawler struct {
	Client crawler.ResultGetterClient
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
		// Only update current classes.
		//if class.Year == getCurrentSession() {
		go c.runClass(user, i, doneCh)
		//}
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
		fieldGroup: {class.Group},
		fieldYear:  {class.Year},
	}
	requestString := fmt.Sprintf("%s?%s", urlResultats, params.Encode())
	req, err := http.NewRequest("POST", requestString, nil)
	if err != nil {
		doneCh <- crawler.RunResult{
			ClassIndex: classIndex,
			Err:        err,
		}
		return
	}

	req.Header.Set("User-Agent", userAgent)

	log.Printf("Sending request for %s\n", class.Name)
	resp, err := c.Client.Do(req)
	if err != nil {
		doneCh <- crawler.RunResult{
			ClassIndex: classIndex,
			Err:        err,
		}
		return
	}
	defer resp.Body.Close()

	log.Printf("Parsing response for %s\n", class.Name)
	results, err := parseResponse(resp.Body)
	if err != nil {
		doneCh <- crawler.RunResult{
			ClassIndex: classIndex,
			Err:        err,
		}
		return
	}
	doneCh <- crawler.RunResult{
		ClassIndex: classIndex,
		Class:      results,
	}
}

func getCurrentSession() string {
	now := time.Now()
	session := 1
	switch now.Month() {
	case time.January, time.February, time.March, time.April, time.May:
		session = 1
	case time.June, time.July, time.August:
		session = 2
	case time.September, time.October, time.November, time.December:
		session = 3
	}
	return fmt.Sprintf("%d%d", now.Year(), session)
}

func parseResponse(resp io.Reader) (*api.Class, error) {
	var results *api.Class
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
				if getAttribute(n, "name") == "form" {
					log.Println("Found results table")
					results, err = parseResultsTable(n)
					done = true
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

type resultIndexes struct {
	Result       int
	Average      int
	StandardDev  int
	WResult      int
	WAverage     int
	WStandardDev int
}

func parseResultsTable(node *html.Node) (*api.Class, error) {
	// Get all rows from the table
	var resRows []*html.Node
	var otherRows []*html.Node
	tBody := node.FirstChild.NextSibling
	for c := tBody.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "tr" {
			// Result rows have no bgcolor attribute, other rows have.
			if hasAttribute(c, "bgcolor") {
				otherRows = append(otherRows, c)
			} else {
				resRows = append(resRows, c)
			}
		}
	}

	// Get the indexes based on the number of columns.
	// Could probably do something prettier but this works well.
	headersRow := parseRow(otherRows[1])
	var indexes *resultIndexes
	switch len(headersRow) {
	case 3:
		indexes = &resultIndexes{1, -1, -1, 2, -1, -1}
	case 5:
		indexes = &resultIndexes{1, 2, -1, 3, 4, -1}
	case 7:
		indexes = &resultIndexes{1, 2, 3, 4, 5, 6}
	case 9:
		indexes = &resultIndexes{1, 2, 3, 5, 6, 7}
	default:
		return nil, errors.New("Unknown table layout")
	}

	class := &api.Class{}

	// Parse other rows
	// First 2 rows are titles then 3rd row is totals and if there
	// is a 4th row it is final grade.

	// Total
	totalRow := parseRow(otherRows[2])
	class.Total = api.ResultInfo{
		Result:      resAt(indexes.Result, totalRow),
		Average:     resAt(indexes.Average, totalRow),
		StandardDev: resAt(indexes.StandardDev, totalRow),
	}

	// Final grade if available
	if len(otherRows) > 3 {
		finalRow := parseRow(otherRows[3])
		class.Final = finalRow[1]
	}

	log.Println(fmt.Sprintf("Found %v results", len(resRows)))

	// Parse result rows
	for _, row := range resRows {
		res := parseResultRow(row, indexes)
		class.Results = append(class.Results, res)
	}
	return class, nil
}

func parseResultRow(node *html.Node, indexes *resultIndexes) api.Result {
	cols := parseRow(node)
	return api.Result{
		Name: cols[0],
		Normal: api.ResultInfo{
			Result:      resAt(indexes.Result, cols),
			Average:     resAt(indexes.Average, cols),
			StandardDev: resAt(indexes.StandardDev, cols),
		},
		Weighted: api.ResultInfo{
			Result:      resAt(indexes.WResult, cols),
			Average:     resAt(indexes.WAverage, cols),
			StandardDev: resAt(indexes.WStandardDev, cols),
		},
	}
}

func parseRow(node *html.Node) (cols []string) {
	// Get all data from text nodes inside the specified node.
	var f func(*html.Node) string
	f = func(n *html.Node) string {
		str := ""
		if n.Type == html.TextNode {
			str += n.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			str += f(c)
		}
		return str
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "td" {

			cols = append(cols, strings.TrimSpace(f(c)))
		}
	}

	return cols
}

func getAttribute(node *html.Node, attribute string) string {
	for _, attr := range node.Attr {
		if attr.Key == attribute {
			return attr.Val
		}
	}

	return ""
}

func hasAttribute(node *html.Node, attribute string) bool {
	return len(getAttribute(node, attribute)) > 0
}

func resAt(index int, cols []string) string {
	if index == -1 {
		return "N/A"
	}
	return cols[index]
}
