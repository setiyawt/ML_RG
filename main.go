package main

import (
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
	// TODO: replace this
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

func (c *AIModelConnector) ConnectAIModel(payload interface{}, token string) (Response, error) {
	// TODO: replace this
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return Response{}, err
	}

	req, err := http.NewRequest("POST", "https://api-inference.huggingface.co/models/google/tapas-base-finetuned-wtq", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return Response{}, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}

	var hfResponse Response
	err = json.Unmarshal(body, &hfResponse)
	if err != nil {
		return Response{}, err
	}

	return hfResponse, nil
}

func main() {
	// TODO: answer here
	data, err := os.ReadFile("data-series.csv")
	if err != nil {
		fmt.Println("Error reading CSV file:", err)
		return
	}

	// Convert CSV data to a slice
	table, err := CsvToSlice(string(data))
	if err != nil {
		fmt.Println("Error converting CSV to slice:", err)
		return
	}

	// Create an instance of AIModelConnector
	connector := &AIModelConnector{
		Client: &http.Client{},
	}

	// Prompt the user for a query
	fmt.Print("Enter your query: ")
	var query string
	fmt.Scanln(&query)

	// Connect to the AI model and get the response
	payload := Inputs{
		Table: table,
		Query: query,
	}
	response, err := connector.ConnectAIModel(payload, "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")
	if err != nil {
		fmt.Println("Error connecting to AI model:", err)
		return
	}

	// Print the response
	fmt.Printf("Answer: %s\n", response.Answer)
	fmt.Printf("Coordinates: %v\n", response.Coordinates)
	fmt.Printf("Cells: %v\n", response.Cells)
	fmt.Printf("Aggregator: %s\n", response.Aggregator)
}
