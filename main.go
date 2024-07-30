package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type AIModelConnector struct {
	Client *http.Client
}

type Inputs struct {
	Table map[string][]string `json:"table"`
	Query string              `json:"query"`
}

type Response struct {
	Answer      string   `json:"answer"`
	Coordinates [][]int  `json:"coordinates"`
	Cells       []string `json:"cells"`
	Aggregator  string   `json:"aggregator"`
}

func CsvToSlice(data string) (map[string][]string, error) {
	reader := csv.NewReader(strings.NewReader(data))
	reader.FieldsPerRecord = -1 // Allow inconsistent number of fields

	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}

	result := make(map[string][]string, len(headers))
	for _, header := range headers {
		result[header] = []string{}
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		for i, value := range record {
			result[headers[i]] = append(result[headers[i]], value)
		}
	}

	return result, nil
}

func (c *AIModelConnector) ConnectAIModel(payload interface{}, token string) (Response, *http.Response, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return Response{}, nil, err
	}

	req, err := http.NewRequest("POST", "https://api-inference.huggingface.co/models/google/tapas-base-finetuned-wtq", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return Response{}, nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return Response{}, nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Response{}, nil, err
	}

	fmt.Printf("Status Code: %d\n", resp.StatusCode)
	fmt.Printf("Response Body: %s\n", body)

	var hfResponse Response
	err = json.Unmarshal(body, &hfResponse)
	if err != nil {
		return Response{}, resp, err
	}

	return hfResponse, resp, nil
}

func main() {
	data, err := os.ReadFile("data-series.csv")
	if err != nil {
		fmt.Println("Error reading CSV file:", err)
		return
	}

	table, err := CsvToSlice(string(data))
	if err != nil {
		fmt.Println("Error converting CSV to slice:", err)
		return
	}

	connector := &AIModelConnector{
		Client: &http.Client{},
	}

	fmt.Print("Enter your query: ")
	reader := bufio.NewReader(os.Stdin)
	query, _ := reader.ReadString('\n')
	query = strings.TrimSpace(query)

	payload := Inputs{
		Table: table,
		Query: query,
	}

	fmt.Printf("Payload: %+v\n", payload)

	response, httpResp, err := connector.ConnectAIModel(payload, "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")
	if err != nil {
		fmt.Println("Error connecting to AI model:", err)
		return
	}

	fmt.Printf("Status Code: %d\n", httpResp.StatusCode)

	if httpResp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(httpResp.Body)
		fmt.Printf("Error Response: %s\n", string(body))
	} else {
		fmt.Printf("Response: %+v\n", response)
	}
}
