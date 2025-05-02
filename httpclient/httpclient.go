/*
(C) Copyright Hewlett Packard Enterprise Development LP
*/

// Package httpclient implements HTTP Methods to communicate with Fabric Manager's APIs
package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.hpe.com/hpe/sshot-net-operator/models"
)

// Client describe the BaseURL for Fabric Manager
type Client struct {
	BaseURL string
}

const (
	//ContextDeadline is the context deadline
	ContextDeadline = 30 * time.Second
)

// NewClient returns a client
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
	}
}

// SendRequest method implements logic to communicate with Fabric Manager's APIs
func (c *Client) SendRequest(ctx context.Context, method string, path string, data interface{}) ([]byte, error) {
	var err error
	req := &http.Request{}
	tlsConfig := &tls.Config{}

	//create context with timeout
	ctx, cancel := context.WithTimeout(ctx, ContextDeadline)
	defer cancel()

	if strings.Contains(path, "token") {
		formData, ok := data.(map[string]string)
		if !ok {
			return nil, fmt.Errorf("invalid data type for token request")
		}

		formdata := url.Values{}
		formdata.Set("grant_type", formData["grant_type"])
		formdata.Set("client_id", formData["client_id"])
		formdata.Set("client_secret", formData["client_secret"])
		formdata.Set("scope", formData["scope"])

		req, err = http.NewRequestWithContext(ctx, method, c.BaseURL+path, strings.NewReader(formdata.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")
	} else {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequestWithContext(ctx, method, c.BaseURL+path, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+models.AccessToken)
	}

	if models.SkipTLSVerify == "true" {
		tlsConfig.InsecureSkipVerify = true
	}

	if models.SkipTLSVerify == "false" {
		// Path to the CA certificate
		caCertPath := models.CAPublicKey

		// Read the CA certificate
		caCert, err := os.ReadFile(caCertPath)
		if err != nil {
			fmt.Errorf("failed to read CA certificate: %v", err)
		}

		// Create a new CA certificate pool and add the CA certificate to it
		caCertPool := x509.NewCertPool()
		if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
			return nil, fmt.Errorf("failed to append CA certificate")
		}

		// Create a new TLS configuration with the CA certificate pool
		tlsConfig.RootCAs = caCertPool
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client := &http.Client{
		Transport: transport,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 202 {
		var errorResponse models.ErrorResponse
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(responseBody, &errorResponse)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("could not complete request %+v", errorResponse.Message)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return responseBody, nil
}
