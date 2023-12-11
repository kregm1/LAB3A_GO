package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

const fixerAPIKey = "65ecfb4ff508284f9bfa456d79c721d8"
const apiUrl = "http://data.fixer.io/api/latest?access_key=%s&symbols=%s"

type ExchangeRateResponse struct {
	Success bool
	Rates   map[string]float64
}

func getRate(baseCurrency, targetCurrency string) (float64, error) {
	url := fmt.Sprintf(apiUrl, fixerAPIKey, targetCurrency)
	if baseCurrency != "EUR" {
		url = fmt.Sprintf("%s,%s", url, baseCurrency)
	}

	resp, err := http.Get(url)
	if err != nil {
		return 0.0, err
	}
	defer resp.Body.Close()

	var data ExchangeRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0.0, err
	}

	if !data.Success {
		return 0.0, fmt.Errorf("error fetching exchange rate")
	}

	if baseCurrency == "EUR" {
		return data.Rates[targetCurrency], nil
	}

	targetRate := data.Rates[targetCurrency]
	baseRate := data.Rates[baseCurrency]
	return targetRate / baseRate, nil
}

func homePage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))
	tmpl.Execute(w, nil)
}

func convertCurrency(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	baseCurrency := r.FormValue("base")
	targetCurrency := r.FormValue("target")
	amountStr := r.FormValue("amount")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	rate, err := getRate(baseCurrency, targetCurrency)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting exchange rate: %v", err), http.StatusInternalServerError)
		return
	}

	convertedAmount := amount * rate

	fmt.Fprintf(w, "Converted amount: %.2f %s", convertedAmount, targetCurrency)
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", homePage)
	r.HandleFunc("/convert", convertCurrency).Methods("POST")

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
