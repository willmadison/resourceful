package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/willmadison/resourceful/repository"
	"github.com/willmadison/resourceful/resourceful"
)

func main() {
	router := registerEndpoints(repository.NewInMemory())
	http.Handle("/", router)

	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}

func registerEndpoints(repo resourceful.Repository) *mux.Router {
	r := mux.NewRouter()

	includeValidation := true

	validationFlag := os.Getenv("WITH_VALIDATION")
	if validationFlag != "" {
		includeValidation, _ = strconv.ParseBool(validationFlag)
	}

	handler := func(response http.ResponseWriter, request *http.Request) {
		type slashCommandResponse struct {
			Type        string `json:"response_type"`
			Text        string `json:"text"`
			Attachments []struct {
				Text string `json:"text"`
			} `json:"attachments"`
		}

		request.ParseForm()

		responseURL := request.FormValue("response_url")
		http.Post(responseURL, "text/plain", strings.NewReader("Hello from resourceful!"))

		command := request.FormValue("text")
		commandParts := strings.Fields(command)

		if len(commandParts) > 3 {
			resourceType := commandParts[1]
			uri := commandParts[len(commandParts)-1]
			url, _ := url.Parse(uri)
			title := strings.Join(commandParts[2:len(commandParts)-1], " ")

			resource := resourceful.Resource{
				Type:  resourceType,
				Title: title,
				URL:   *url,
			}

			err := repo.Add(resource)
			if err != nil {
				http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			cmdResponse := slashCommandResponse{
				Type: "in_channel",
				Text: "I got you fam. Adding that resource right now.",
				Attachments: []struct {
					Text string `json:"text"`
				}{
					{
						Text: "Resource added. Please find it here: URL_PLACEHOLDER",
					},
				},
			}

			commandResponse, err := json.Marshal(cmdResponse)
			if err != nil {
				http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			response.Write(commandResponse)
		}
	}

	if includeValidation {
		handler = withSlackVerification(handler)
	}

	r.HandleFunc("/inbox", handler).Methods("POST")

	return r
}

func withSlackVerification(f http.HandlerFunc) http.HandlerFunc {
	validateSlackRequest := func(request *http.Request) error {
		var b bytes.Buffer

		version := os.Getenv("API_VERSION")
		if version == "" {
			version = "v0"
		}

		b.WriteString(version)
		b.WriteString(":")

		rawTimestamp := request.Header.Get("X-Slack-Request-Timestamp")
		if rawTimestamp == "" {
			return errors.New("bad request")
		}

		var err error
		verifyTimestamps := true

		verificationFlag := os.Getenv("VERIFY_TIMESTAMPS")
		if verificationFlag != "" {
			verifyTimestamps, err = strconv.ParseBool(verificationFlag)
		}

		if verifyTimestamps {
			ts, err := strconv.Atoi(rawTimestamp)
			if err != nil {
				return errors.New("bad request")
			}
			timestamp := time.Unix(int64(ts), 0)

			if timestamp.Before(time.Now().Add(-5 * time.Minute)) {
				return errors.New("bad request")
			}
		}

		b.WriteString(rawTimestamp)
		b.WriteString(":")

		body, err := ioutil.ReadAll(request.Body)
		defer request.Body.Close()
		if err != nil {
			return err
		}

		request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		b.Write(body)

		signature := request.Header.Get("X-Slack-Signature")
		if signature == "" {
			return errors.New("bad request")
		}

		secret := os.Getenv("SIGNING_SECRET")

		h := hmac.New(sha256.New, []byte(secret))

		h.Write(b.Bytes())

		sha := hex.EncodeToString(h.Sum(nil))

		calculatedSignature := fmt.Sprintf("%s=%s", version, sha)

		if calculatedSignature != signature {
			return errors.New("bad request")
		}

		return nil
	}

	return func(response http.ResponseWriter, request *http.Request) {
		err := validateSlackRequest(request)
		if err != nil {
			http.Error(response, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		f(response, request)
	}
}
