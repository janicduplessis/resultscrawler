package webserver

type (
	userModel struct {
		Email     string `json:"email"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}

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

	loginResponse struct {
		Status int        `json:"status"`
		User   *userModel `json:"user"`
	}

	registerResponse struct {
		Status int        `json:"status"`
		User   *userModel `json:"user"`
	}

	crawlerConfigClassModel struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Year  string `json:"year"`
		Group string `json:"group"`
	}

	crawlerConfigModel struct {
		Status            bool   `json:"status"`
		Code              string `json:"code"`
		Nip               string `json:"nip"`
		NotificationEmail string `json:"notificationEmail"`
	}
)
