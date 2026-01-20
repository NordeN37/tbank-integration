package tbank

type Config struct {
	TerminalKey string
	SecretKey   string

	SuccessURL  *string
	FailURL     *string
	CallbackURL string

	Debug bool // Логировтаь запросы в банк, логиует в виде курла
}
