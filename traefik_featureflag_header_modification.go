// Package featureflag_header-modfication a plugin for traefik which adds request header.
package traefik_featureflag_header_modification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// Config the plugin configuration.
type Config struct {
	Headers       []string
	FliptEndpoint string
	FlagKey       string
	HeaderResult  string
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Headers:       []string{"id"},
		FliptEndpoint: "",
		FlagKey:       "",
	}
}

// XRequestStart a traefik plugin.
type FeatureflagHeaderModification struct {
	next          http.Handler
	headers       []string
	fliptEndpoint string
	flagKey       string
	headerResult  string
	name          string
}

type FliptEvaluateResponse struct {
	RequestID             string    `json:"requestId"`
	EntityID              string    `json:"entityId"`
	Match                 bool      `json:"match"`
	FlagKey               string    `json:"flagKey"`
	SegmentKey            string    `json:"segmentKey"`
	Timestamp             time.Time `json:"timestamp"`
	Value                 string    `json:"value"`
	RequestDurationMillis float64   `json:"requestDurationMillis"`
	Attachment            string    `json:"attachment"`
	Reason                string    `json:"reason"`
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if len(config.Headers) == 0 {
		return nil, fmt.Errorf("headers cannot be empty")
	}
	if len(config.FlagKey) == 0 {
		return nil, fmt.Errorf("FlagKeyss cannot be empty")
	}
	if len(config.FliptEndpoint) == 0 {
		return nil, fmt.Errorf("FliptEndpoint cannot be empty")
	}
	if len(config.HeaderResult) == 0 {
		return nil, fmt.Errorf("HeaderResult cannot be empty")
	}

	return &FeatureflagHeaderModification{
		headers:       config.Headers,
		fliptEndpoint: config.FliptEndpoint,
		flagKey:       config.FlagKey,
		headerResult:  config.HeaderResult,
		next:          next,
		name:          name,
	}, nil
}

func (config *FeatureflagHeaderModification) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var org string
	var orgHeaderName string
	for i := 0; i < len(config.headers); i++ {
		org = req.Header.Get(config.headers[i])
		if len(org) != 0 {
			orgHeaderName = config.headers[i]
			break
		}
	}
	os.Stdout.WriteString("traefik featuer flag header modification" + org)
	if len(org) == 0 {
		config.next.ServeHTTP(rw, req)
		return
	}

	payload := []byte(`{"entityId":"` + org + `","flagKey":"` + config.flagKey + `","context":{"` + orgHeaderName + `":"` + org + `"}}`)

	resp, err := http.Post(config.fliptEndpoint, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		os.Stderr.WriteString("Error when getting feature flag:" + err.Error())
		config.next.ServeHTTP(rw, req)
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		os.Stderr.WriteString("Error while reading response:" + err.Error())
		config.next.ServeHTTP(rw, req)
		return
	}

	var fliptEvaluateResponse FliptEvaluateResponse
	err = json.Unmarshal(body, &fliptEvaluateResponse)
	if err != nil {
		os.Stderr.WriteString("Error while parsing response:" + err.Error())
		config.next.ServeHTTP(rw, req)
		return
	}
	if fliptEvaluateResponse.Match == true {
		req.Header.Set(config.headerResult, fliptEvaluateResponse.Value)
	}

	config.next.ServeHTTP(rw, req)
}
