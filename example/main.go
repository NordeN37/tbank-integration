package main

import (
	"context"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/NordeN37/tbank-integration/tbank"
)

var tmpl = template.Must(template.ParseGlob("example/templates/*.html"))

func main() {
	successURL := os.Getenv("SUCCESS_URL")
	failURL := os.Getenv("FAIL_URL")
	terminalKey := os.Getenv("TERMINAL_KEY")
	secretKey := os.Getenv("SECRET_KEY")
	callbackURL := os.Getenv("CALLBACK_URL")

	tcfg := tbank.Config{
		TerminalKey: terminalKey,
		SecretKey:   secretKey,
		SuccessURL:  &successURL,
		FailURL:     &failURL,
		CallbackURL: callbackURL,
	}
	var s = Server{
		tcfg,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "payment_page.html", nil)
	})
	mux.HandleFunc("/pay", s.PaymentHandler)
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		var res interface{}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			return
		}
		defer r.Body.Close()

		if err := json.Unmarshal(body, &res); err != nil {
			log.Println(err)
			return
		}
		log.Println(res)
	})

	http.ListenAndServe(":8080", mux)

}

type Server struct {
	tConf tbank.Config
}

func (s *Server) PaymentHandler(w http.ResponseWriter, r *http.Request) {
	tb := tbank.New(s.tConf)
	orderTest := RandomInt(5)
	resp, err := tb.Init(context.Background(), tbank.InitRequest{
		OrderId:     orderTest,
		Amount:      2000,
		Description: "Оплата заказа #123",
		Receipt: tbank.InitRequestReceipt{
			Email:    "test@test.ru",
			Phone:    "+79997309293",
			Taxation: "usn_income",
			Items: []tbank.InitRequestReceiptItems{
				tbank.InitRequestReceiptItems{
					Name:     "test",
					Price:    1000,
					Quantity: 2,
					Amount:   2000,
					Tax:      "vat5",
				},
			},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	w.WriteHeader(http.StatusOK)
}

func RandomInt(n int) string {
	// Инициализируем генератор случайных чисел
	rand.Seed(time.Now().UnixNano())

	// Определяем цифры, которые можно использовать
	digits := "0123456789"

	// Создаем слайс байт нужной длины
	result := make([]byte, n)

	// Заполняем случайными цифрами
	for i := 0; i < n; i++ {
		result[i] = digits[rand.Intn(len(digits))]
	}

	return string(result)
}
