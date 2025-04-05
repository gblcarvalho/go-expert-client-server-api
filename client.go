package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"
)

type DollarPriceRcv struct {
	Bid string `json:"bid"`
}

func main() {
	f, err := os.Create("cotacao.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 300 * time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	var price DollarPriceRcv
	err = json.Unmarshal(body, &price)
	if err != nil {
		panic(err)
	}
	f.Write([]byte("Dólar: " + price.Bid))
}

