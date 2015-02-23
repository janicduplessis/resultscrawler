package webserver

import (
	"time"

	"github.com/janicduplessis/resultscrawler/pkg/api"
)

type (
	jsonTime time.Time

	// requests
	loginRequest struct {
		Email      string `json:"email"`
		Password   string `json:"password"`
		DeviceType int    `json:"deviceType"`
	}

	registerRequest struct {
		Email             string `json:"email"`
		Password          string `json:"password"`
		FirstName         string `json:"firstName"`
		LastName          string `json:"lastName"`
		NotificationToken string `json:"notificationToken"`
		DeviceType        int    `json:"deviceType"`
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
		LastUpdate time.Time   `json:"lastUpdate"`
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
