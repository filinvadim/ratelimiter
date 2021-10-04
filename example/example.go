package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/filinvadim/ratelimiter"
	"io"
	"log"
	"net/http"
	"time"
)

type errorResponse struct {
	ErrorMessage string `json:"error_message"`
	ErrorCode    string `json:"error_code"`
}

func main()  {
	// Rate Limit: 1 request per 5 seconds
	url := "https://www.okex.com/api/system/v3/status"
	window := (5 * time.Second) + (time.Millisecond * 150) // 150ms gap to align with okex server time
	l := ratelimiter.NewLimiter(context.TODO(), 1, window)
	defer l.Close()

	for {
		l.Limit(1, func() {
			resp, err := http.Get(url)
			if err != nil {
				log.Fatalln(err)
			}
			bt, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			var errResp errorResponse
			json.Unmarshal(bt, &errResp)
			if errResp.ErrorCode == "30014" { // too many requests
				log.Fatalln(errResp.ErrorMessage)
			}
		})
		fmt.Println(time.Now().Second())
	}
}