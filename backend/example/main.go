// backend/example/main.go
package main

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/NordeN37/tbank-integration" // Импорт нашего пакета
	"github.com/gorilla/mux"
)

//go:embed templates/*
var templateFS embed.FS

// Данные для передачи в шаблон
type PageData struct {
	TerminalKey string
	BackendUrl  string
	Title       string
}

func main() {
	// 1. Инициализация T-Bank клиента (используем ваш пакет)
	config := tbank.Config{
		TerminalKey: "1673598390848DEMO",
		Password:    "cfzmnx00dfoystwf",
		BaseURL:     "https://securepayments.tbank.ru/eacq/v2",
		CallbackURL: "https://al-home-test.ru/api/payment/callback",
		SuccessURL:  "https://al-home-test.ru/payment/success",
		FailURL:     "https://al-home-test.ru/payment/fail",
	}

	client := tbank.NewClient(config)

	// 2. Создание роутера
	r := mux.NewRouter()

	// 3. Загрузка HTML шаблона
	tmpl, err := template.ParseFS(templateFS, "templates/example.html")
	if err != nil {
		log.Fatal("Failed to parse template:", err)
	}

	// 4. Маршрут для главной страницы
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			TerminalKey: config.TerminalKey,
			BackendUrl:  "http://localhost:8080/api/payment/initiate",
			Title:       "T-Bank Payment Integration Demo",
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, data)
	})

	// 5. API эндпоинты (используем обработчики из нашего пакета)
	r.HandleFunc("/api/payment/initiate", client.PaymentHandler()).Methods("POST")
	r.HandleFunc("/api/payment/callback", func(w http.ResponseWriter, r *http.Request) {
		// Обработка вебхука от T-Bank
		body, _ := io.ReadAll(r.Body)
		notification, err := client.HandleNotification(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("Payment notification: Order=%s, Status=%s",
			notification.OrderID, notification.Status)
		w.WriteHeader(http.StatusOK)
	}).Methods("POST")

	// 6. Статические файлы (JS, CSS)
	// Встраиваем JS файл из нашей библиотеки
	r.HandleFunc("/static/include-tbank.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/js")
		// Здесь можно встроить содержимое include-tbank.js
		http.ServeFile(w, r, "../include-tbank.js") // Путь относительно executable
	})

	// 7. Старт сервера
	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("Payment page: http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", r))
}
