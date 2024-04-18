package schemas

type UserCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type User struct {
	Id           uint64
	Login        string `json:"login,omitempty"`
	PasswordHash string `json:"password,omitempty"`
	FirstName    string `json:"firstName,omitempty"`
	LastName     string `json:"lastName,omitempty"`
	Birthdate    string `json:"birthdate,omitempty"`
	Email        string `json:"email,omitempty"`
	PhoneNumber  string `json:"phoneNumber,omitempty"`
}
