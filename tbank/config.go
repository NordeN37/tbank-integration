package tbank

type Config struct {
	TerminalKey string
	Password    string

	SuccessURL  string
	FailURL     string
	CallbackURL string
}
