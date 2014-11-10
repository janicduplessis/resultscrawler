package webserver

import "github.com/janicduplessis/resultscrawler/lib"

type Page struct {
	PageID string
	Title  string
	User   *lib.User
}

type homeModel struct {
	Page Page
}

type aboutModel struct {
	Page Page
}
