package main

import (
	"embed"
	"fmt"
	"github.com/devmax-pro/order-service-frontend/internal/publisher"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
)

const orderServiceUrl = "http://localhost:8080"

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "test.page.gohtml")
	})

	NatsURL := os.Getenv("NATS_URL")
	pub, err := publisher.New(
		NatsURL,
		"orders-streaming",
		"order-client-id",
	)
	if err != nil {
		log.Fatal("Error occurred while init publisher", err)
	}

	log.Println("Nuts Publisher started")
	http.HandleFunc("/send-order", func(w http.ResponseWriter, r *http.Request) {

		payload := []byte(fmt.Sprintf(`{
		  "order_uid": "%s",
		  "track_number": "WBILMTESTTRACK",
		  "entry": "WBIL",
		  "delivery": {
			"name": "Test Testov",
			"phone": "+9720000000",
			"zip": "2639809",
			"city": "Kiryat Mozkin",
			"address": "Ploshad Mira 15",
			"region": "Kraiot",
			"email": "test@gmail.com"
		  },
		  "payment": {
			"transaction": "b563feb7b2b84b6test",
			"request_id": "",
			"currency": "USD",
			"provider": "wbpay",
			"amount": 1817,
			"payment_dt": 1637907727,
			"bank": "alpha",
			"delivery_cost": 1500,
			"goods_total": 317,
			"custom_fee": 0
		  },
		  "items": [
			{
			  "chrt_id": 9934930,
			  "track_number": "WBILMTESTTRACK",
			  "price": 453,
			  "rid": "ab4219087a764ae0btest",
			  "name": "Mascaras",
			  "sale": 30,
			  "size": "0",
			  "total_price": 317,
			  "nm_id": 2389212,
			  "brand": "Vivienne Sabo",
			  "status": 202
			}
		  ],
		  "locale": "en",
		  "internal_signature": "",
		  "customer_id": "test",
		  "delivery_service": "meest",
		  "shardkey": "9",
		  "sm_id": 99,
		  "date_created": "2021-11-26T06:22:19Z",
		  "oof_shard": "1"
		}`, randString(19)))

		err = pub.Publish("orders-channel", payload)
		if err != nil {
			log.Println("Message publish failed")
		}

		_, err = w.Write(payload)
		if err != nil {
			log.Println("Response write failed")
			return
		}
	})

	fmt.Println("Starting front end service on port 8081")
	err = http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatalln(err)
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

//go:embed templates
var templateFS embed.FS

func render(w http.ResponseWriter, t string) {

	partials := []string{
		"templates/base.layout.gohtml",
	}

	var templateSlice []string
	templateSlice = append(templateSlice, fmt.Sprintf("templates/%s", t))

	for _, x := range partials {
		templateSlice = append(templateSlice, x)
	}

	tmpl, err := template.ParseFS(templateFS, templateSlice...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var data struct {
		OrderServiceURL string
	}

	data.OrderServiceURL = orderServiceUrl

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
