package jsonrpc

import (
	"bytes"
	"encoding/json"
	models "eth_bal/internal/models"
	"fmt"
	"net/http"
)

func SendJSONRPCRequest(client *http.Client, apiKey string, request models.JSONRPCRequest, response *models.JSONRPCResponse) error {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", "https://go.getblock.io/"+apiKey+"/", bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return err
	}
	if response.Error != nil {
		return fmt.Errorf("JSON-RPC Error: %s", response.Error.Message)
	}
	return nil
}

func SendBatchJSONRPCRequest(client *http.Client, apiKey string, requests []models.JSONRPCRequest, responses *[]models.JSONRPCResponse) error {
	jsonData, err := json.Marshal(requests)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", "https://go.getblock.io/"+apiKey+"/", bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(responses)
	if err != nil {
		return err
	}
	return nil
}
