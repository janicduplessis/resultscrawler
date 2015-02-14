package api

import "time"

type (
	// The Crawler interface exposes the public crawler api.
	Crawler interface {
		Refresh(userID string) error
	}

	// CrawlerConfig contains info about the crawler configuration.
	CrawlerConfig struct {
		UserID            string `json:"userId"`
		Status            bool   `json:"status"`
		Code              string `json:"code"`
		Nip               string `json:"nip"`
		NotificationEmail string `json:"notificationEmail"`
	}

	// User contains info about users.
	User struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}

	// Results contains all results for a user organized by class.
	Results struct {
		UserID     string    `json:"userId"`
		LastUpdate time.Time `json:"lastUpdate"`
		Classes    []Class   `json:"classes"`
	}

	// Class is an entity for a class.
	Class struct {
		ID      string     `json:"id"`
		Name    string     `json:"name"`
		Group   string     `json:"group"`
		Year    string     `json:"year"`
		Results []Result   `json:"results"`
		Total   ResultInfo `json:"total"`
		Final   string     `json:"final"`
	}

	// Result is an entity for storing a result
	Result struct {
		Name     string     `json:"name"`
		Normal   ResultInfo `json:"normal"`
		Weighted ResultInfo `json:"weighted"`
	}

	// ResultInfo contains normal or weighted grades.
	ResultInfo struct {
		Result      string `json:"result"`
		Average     string `json:"average"`
		StandardDev string `json:"standardDev"`
	}
)
