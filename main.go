package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	c := flag.String("config", "config.json", "config file path")
	flag.Parse()

	file, err := os.ReadFile(*c)
	if err != nil {
		log.Panicf("read config file error: %v", err)
	}
	var config struct {
		EndpointFormat struct {
			ChatCompletions  string `json:"chat_completions"`
			ImageGenerations string `json:"image_generations"`
		} `json:"endpoint_format"`
		ModelMap map[string]string `json:"model_map"`
		Address  string            `json:"address"`
	}
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Panicf("parse config file error: %v", err)
	}

	handler := func(urlFormat string, writer http.ResponseWriter, request *http.Request) {
		var body map[string]any
		e := json.NewDecoder(request.Body).Decode(&body)
		if e != nil {
			http.Error(writer, "invalid request body", http.StatusBadRequest)
			return
		}
		model, ok := body["model"].(string)
		if !ok {
			http.Error(writer, "invalid model", http.StatusBadRequest)
			return
		}
		if m, exist := config.ModelMap[model]; exist {
			model = m
		}
		u := fmt.Sprintf(urlFormat, model)
		delete(body, "model")
		bodyBytes, e := json.Marshal(body)
		if e != nil {
			http.Error(writer, e.Error(), http.StatusInternalServerError)
			return
		}
		req, e := http.NewRequest(http.MethodPost, u, bytes.NewReader(bodyBytes))
		if e != nil {
			http.Error(writer, e.Error(), http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		token := request.Header.Get("Authorization")
		if token != "" && strings.HasPrefix(token, "Bearer ") {
			req.Header.Set("api-key", token[7:])
		}
		resp, e := http.DefaultClient.Do(req)
		if e != nil {
			http.Error(writer, e.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		for k, v := range resp.Header {
			for _, vv := range v {
				writer.Header().Add(k, vv)
			}
		}
		writer.WriteHeader(resp.StatusCode)
		_, e = io.Copy(writer, resp.Body)
		if e != nil {
			http.Error(writer, e.Error(), http.StatusInternalServerError)
			return
		}
	}
	http.HandleFunc("/v1/chat/completions", func(writer http.ResponseWriter, request *http.Request) {
		handler(config.EndpointFormat.ChatCompletions, writer, request)
	})
	http.HandleFunc("images/generations", func(writer http.ResponseWriter, request *http.Request) {
		handler(config.EndpointFormat.ImageGenerations, writer, request)
	})
	log.Printf("listening on %s", config.Address)
	log.Fatal(http.ListenAndServe(config.Address, nil))
}
