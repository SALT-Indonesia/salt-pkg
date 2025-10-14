package create_user

type Request struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
