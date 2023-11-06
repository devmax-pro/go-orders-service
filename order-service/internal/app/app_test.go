package app_test

import (
	"encoding/json"
	"github.com/devmax-pro/order-service/internal/adapters/cache"
	"github.com/devmax-pro/order-service/internal/adapters/http/controller"
	"github.com/devmax-pro/order-service/internal/adapters/http/router"
	"github.com/devmax-pro/order-service/internal/adapters/storage/memory"
	"github.com/devmax-pro/order-service/internal/entities"
	"github.com/devmax-pro/order-service/internal/usecases/get_order"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestGetOrder(t *testing.T) {
	createdAt := time.Now()
	order := entities.Order{
		OrderUID:    "b563feb7b2b84b6test",
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		Delivery: entities.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: entities.Payment{
			Transaction:  "b563feb7b2b84b6test",
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []entities.Item{
			{
				ChrtID:      9934930,
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				Rid:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		ShardKey:          "9",
		SmID:              99,
		DateCreated:       createdAt,
		OofShard:          "1",
	}

	repo := memory.NewOrders(map[string]*entities.Order{order.OrderUID: &order})
	csh := cache.NewMemoryCache[entities.Order]()
	getOrderHandler := get_order.New(repo, csh)
	ctrl := controller.New(getOrderHandler)
	rtr := router.New(ctrl)

	req, _ := http.NewRequest("GET", "/order/b563feb7b2b84b6test", nil)
	rr := httptest.NewRecorder()

	rtr.ServeHTTP(rr, req)

	assertStatus(t, rr.Code, http.StatusOK)
	assertContentType(t, rr, "application/json")

	resp := getOrderFromResponse(t, rr.Body)
	resp.DateCreated = createdAt // Hack for comparison of time.Time

	assertOrder(t, resp, order)
}

func getOrderFromResponse(t testing.TB, body io.Reader) (order entities.Order) {
	t.Helper()
	err := json.NewDecoder(body).Decode(&order)

	if err != nil {
		t.Fatalf("Unable to parse response from server %q into slice of Order, '%v'", body, err)
	}

	return
}

func assertOrder(t testing.TB, got, want entities.Order) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct status, got %d, want %d", got, want)
	}
}

func assertContentType(t testing.TB, response *httptest.ResponseRecorder, want string) {
	t.Helper()
	if response.Header().Get("Content-Type") != want {
		t.Errorf("response did not have content-type of %s, got %v", want, response.Result().Header)
	}
}
