package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DollarPrice struct {
	Bid string `json:"bid"`
}

type EconomiaUSDBRL struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func main() {
	http.HandleFunc("/cotacao", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	db, err := openDatabase()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()
	usdBRL, err := getEconomiaUSDBRL()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = savePriceDatabase(db, usdBRL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	dollarPrice := &DollarPrice{
		Bid: usdBRL.USDBRL.Bid,
	}
	json.NewEncoder(w).Encode(dollarPrice)
}

func getEconomiaUSDBRL() (*EconomiaUSDBRL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	url := "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var usdBRL EconomiaUSDBRL
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &usdBRL)
	if err != nil {
		return nil, err
	}
	return &usdBRL, nil
}

func openDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./sqlite.db")
	if err != nil {
		return nil, err
	}
	createTable := `
	CREATE TABLE IF NOT EXISTS prices (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT,
		codein TEXT,
		name TEXT,
		high TEXT,
		low TEXT,
		varBid TEXT,
		pctChange TEXT,
		bid TEXT,
		ask TEXT,
		timestamp TEXT,
		create_date TEXT
	);`
	_, err = db.Exec(createTable)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func savePriceDatabase(db *sql.DB, usdBRL *EconomiaUSDBRL) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	insertInto := `
	INSERT INTO prices (
		code, codein, name, high, low, varBid,
		pctChange, bid, ask, timestamp, create_date
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	_, err := db.ExecContext(ctx, insertInto,
		usdBRL.USDBRL.Code,
		usdBRL.USDBRL.Codein,
		usdBRL.USDBRL.Name,
		usdBRL.USDBRL.High,
		usdBRL.USDBRL.Low,
		usdBRL.USDBRL.VarBid,
		usdBRL.USDBRL.PctChange,
		usdBRL.USDBRL.Bid,
		usdBRL.USDBRL.Ask,
		usdBRL.USDBRL.Timestamp,
		usdBRL.USDBRL.CreateDate,
	)
	if err != nil {
		return err
	}
	return nil
}
