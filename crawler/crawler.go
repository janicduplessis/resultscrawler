package crawler

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

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
	log.Println("start")

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
	log.Println(requestString)
	req, err := http.NewRequest("POST", requestString, nil)
	if err != nil {
		doneCh <- false
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		doneCh <- false
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		doneCh <- false
	}

	log.Println(string(body))

	doneCh <- true
}
