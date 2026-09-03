package credentials

var sampleSecretKey = []byte("password")

type Token struct {
	UserId     string `json:"userId"`
	SessionId  string `json:"sessionId"`
	Email      string `json:"email"`
	CustomerId string `json:"customerId"`
}
