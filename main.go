package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	TRACE_URL = "https://www.cloudflare.com/cdn-cgi/trace"
	TTL       = 3600
	PROXY     = true
)

var (
	IP_ADDRESS string
)

type DNSDetailsResponse struct {
	Result []Result   `json:"result"`
	Info   ResultInfo `json:"result_info"`
}

type Result struct {
	ID string `json:"id"`
	IP string `json:"content"`
}

type ResultInfo struct {
	Count int `json:"count"`
}

type DNSUpdateRequestBody struct {
	Content string `json:"content"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Proxied bool   `json:"proxied"`
	TTL     int    `json:"ttl"`
}

type DNSUpdateResponseBody struct {
	Success bool `json:"success"`
}

func main() {
	CLOUDFLARE_API_TOKEN := os.Getenv("CLOUDFLARE_API_TOKEN")
	if CLOUDFLARE_API_TOKEN == "" {
		log.Fatal("CLOUDFLARE_API_TOKEN not found")
	}

	ZONE_ID := os.Getenv("ZONE_ID")
	if ZONE_ID == "" {
		log.Fatal("ZONE_ID not found")
	}

	RECORD_NAME := os.Getenv("RECORD_NAME")
	if RECORD_NAME == "" {
		log.Fatal("RECORD_NAME not found")
	}

	CLOUDFLARE_API_URL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", ZONE_ID)

	for {
		time.Sleep(1 * time.Minute)

		// get ip address
		trace_resp, err := http.Get(TRACE_URL)
		if err != nil {
			log.Print(err)
			continue
		}

		body, err := io.ReadAll(trace_resp.Body)
		if err != nil {
			log.Print(err)
			continue
		}

		ipString := strings.Split(string(body), "\n")[2]
		ip := strings.Split(ipString, "=")[1]

		// check ip address
		if ip == IP_ADDRESS {
			continue
		}

		// check if record exists
		req, err := http.NewRequest(http.MethodGet, CLOUDFLARE_API_URL+"?type=A&name="+RECORD_NAME, nil)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		req.Header.Add("Authorization", "Bearer "+CLOUDFLARE_API_TOKEN)
		req.Header.Add("Content-Type", "application/json")

		dnsRecordsResp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatal(err.Error())
		}

		dnsRecords, err := io.ReadAll(dnsRecordsResp.Body)
		if err != nil {
			log.Fatal(err.Error())
		}

		var dnsDetailsResponse DNSDetailsResponse
		err = json.Unmarshal([]byte(dnsRecords), &dnsDetailsResponse)
		if err != nil {
			log.Fatal(err.Error())
		}

		if dnsDetailsResponse.Info.Count == 0 || len(dnsDetailsResponse.Result) == 0 {
			log.Fatal("no A records exist for " + RECORD_NAME)
		}

		if ip == dnsDetailsResponse.Result[0].IP {
			IP_ADDRESS = ip
			continue
		}

		RECORD_ID := dnsDetailsResponse.Result[0].ID

		// update dns record
		updateBody := DNSUpdateRequestBody{
			Content: ip,
			Name:    RECORD_NAME,
			Type:    "A",
			Proxied: PROXY,
			TTL:     TTL,
		}

		updateBodyJSON, err := json.Marshal(updateBody)
		if err != nil {
			log.Print(err.Error())
			continue
		}

		updateRequest, err := http.NewRequest(http.MethodPut, CLOUDFLARE_API_URL+"/"+RECORD_ID, bytes.NewReader(updateBodyJSON))
		if err != nil {
			log.Print(err.Error())
			continue
		}

		updateRequest.Header.Add("Authorization", "Bearer "+CLOUDFLARE_API_TOKEN)
		updateRequest.Header.Add("Content-Type", "application/json")

		updateResponse, err := http.DefaultClient.Do(updateRequest)
		if err != nil {
			log.Print(err.Error())
		}

		updateResponseBody, err := io.ReadAll(updateResponse.Body)
		if err != nil {
			log.Print(err.Error())
			continue
		}

		var dnsUpdateResponseBody DNSUpdateResponseBody
		err = json.Unmarshal(updateResponseBody, &dnsUpdateResponseBody)
		if err != nil {
			log.Print(err.Error())
			continue
		}

		if dnsUpdateResponseBody.Success {
			IP_ADDRESS = ip
		} else {
			log.Print(string(updateResponseBody))
		}
	}
}
