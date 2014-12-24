package webserver

import "github.com/janicduplessis/resultscrawler/pkg/api"

type (
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
		User   *userModel `json:"user"`
	}

	registerResponse struct {
		Status int        `json:"status"`
		User   *userModel `json:"user"`
	}

	resultsResponse struct {
		Year    string      `json:"year"`
		Classes []api.Class `json:"classes"`
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
