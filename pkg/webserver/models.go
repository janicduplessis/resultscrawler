package webserver

import (
	"fmt"
	"time"

	"github.com/janicduplessis/resultscrawler/pkg/api"
)

type (
	jsonTime time.Time

	// requests
	loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	registerRequest struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}

	// responses
	loginResponse struct {
		Status int        `json:"status"`
		Token  string     `json:"token"`
		User   *userModel `json:"user"`
	}

	registerResponse struct {
		Status int        `json:"status"`
		Token  string     `json:"token"`
		User   *userModel `json:"user"`
	}

	resultsResponse struct {
		Year       string      `json:"year"`
		Classes    []api.Class `json:"classes"`
		LastUpdate jsonTime    `json:"lastUpdate"`
	}

	// models
	userModel struct {
		Email     string `json:"email"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}

	crawlerConfigClassModel struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Year  string `json:"year"`
		Group string `json:"group"`
	}
)

func (j jsonTime) MarshalJSON() ([]byte, error) {
	format := time.Time(j).Format("Jan 2, 2006 at 15:04")
	return []byte(fmt.Sprintf("\"%s\"", format)), nil
}
