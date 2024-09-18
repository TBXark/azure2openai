package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var BuildVersion string

type Config struct {
	EndpointFormat struct {
		ChatCompletions  string `json:"chat_completions"`
		ImageGenerations string `json:"image_generations"`
		Models           string `json:"models"`
	} `json:"endpoint_format"`
	ModelMap map[string]string `json:"model_map"`
	Address  string            `json:"address"`
}

func NewConfig(path string) (*Config, error) {
	if strings.HasPrefix(path, "http") {
		resp, err := http.Get(path)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		config := &Config{}
		err = json.NewDecoder(resp.Body).Decode(config)
		if err != nil {
			return nil, err
		}
		return config, nil
	} else {
		bytes, err := os.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}
		config := &Config{}
		err = json.Unmarshal(bytes, config)
		if err != nil {
			return nil, err
		}
		return config, nil
	}
}

type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *HTTPError) Error() string {
	return e.Message
}

func azureRedirect(endpoint string, config *Config) func(http.ResponseWriter, *http.Request) {
	handler := func(writer http.ResponseWriter, request *http.Request) error {

		var reqBody io.Reader = nil
		var uri = endpoint

		if request.Method == http.MethodPost {
			var body map[string]any
			err := json.NewDecoder(request.Body).Decode(&body)
			if err != nil {
				return &HTTPError{http.StatusBadRequest, err.Error()}
			}

			if model, ok := body["model"].(string); ok {
				if m, exist := config.ModelMap[model]; exist {
					model = m
				}
				uri = fmt.Sprintf(uri, model)
				delete(body, "model")
			}

			bodyBytes, err := json.Marshal(body)
			if err != nil {
				return err
			}
			reqBody = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequest(request.Method, uri, reqBody)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		token := request.Header.Get("Authorization")
		if token != "" && strings.HasPrefix(token, "Bearer ") {
			req.Header.Set("api-key", token[7:])
			req.Header.Del("Authorization")
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		for k, v := range resp.Header {
			for _, vv := range v {
				writer.Header().Add(k, vv)
			}
		}
		writer.WriteHeader(resp.StatusCode)
		_, err = io.Copy(writer, resp.Body)
		if err != nil {
			return err
		}
		return nil
	}
	return func(writer http.ResponseWriter, request *http.Request) {
		if err := handler(writer, request); err != nil {
			var httpErr *HTTPError
			if errors.As(err, &httpErr) {
				http.Error(writer, httpErr.Message, httpErr.Code)
			} else {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}

func main() {
	conf := flag.String("config", "config.json", "config file path")
	help := flag.Bool("help", false, "show help")
	flag.Parse()
	if *help {
		fmt.Printf("Version: %s\n", BuildVersion)
		flag.Usage()
		return
	}

	config, err := NewConfig(*conf)
	if err != nil {
		log.Fatal(err)
	}
	startServer(config)
}

func startServer(config *Config) {
	http.HandleFunc("/v1/chat/completions", azureRedirect(config.EndpointFormat.ChatCompletions, config))
	http.HandleFunc("images/generations", azureRedirect(config.EndpointFormat.ImageGenerations, config))
	http.HandleFunc("/v1/models", azureRedirect(config.EndpointFormat.Models, config))
	log.Printf("listening on %s", config.Address)
	log.Fatal(http.ListenAndServe(config.Address, nil))
}
