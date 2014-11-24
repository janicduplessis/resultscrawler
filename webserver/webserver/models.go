package webserver

type (
	userModel struct {
		Email     string `json:"email"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}

	loginRequest struct {
		Email    string `json:"userName"`
		Password string `json:"password"`
	}

	registerRequest struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}

	loginResponse struct {
		Status int        `json:"status"`
		User   *userModel `json:"user"`
	}

	registerResponse struct {
		Status int        `json:"status"`
		User   *userModel `json:"user"`
	}
)
