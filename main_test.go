package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/willmadison/resourceful/repository"
	"github.com/willmadison/resourceful/resourceful"
)

func setupHandler() http.HandlerFunc {
	return setupHandlerWith(repository.NewInMemory())
}

func setupHandlerWith(repo resourceful.Repository) http.HandlerFunc {
	h := registerEndpoints(repo)
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		h.ServeHTTP(response, request)
	})
}

func TestInvalidIncomingSlashCommand(t *testing.T) {
	os.Setenv("SIGNING_SECRET", "8f742231b10e8888abcd99yyyzzz85a5") // Example Signing Secret from https://api.slack.com/docs/verifying-requests-from-slack
	os.Setenv("VERIFY_TIMESTAMPS", "false")
	os.Setenv("WITH_VALIDATION", "true")

	handler := setupHandler()

	body := strings.NewReader(`token=xyzz0WbapA4vBCDEFasx0q6G&team_id=T1DC2JH3J&team_domain=testteamnow&channel_id=G8PSS9T3V&channel_name=foobar&user_id=U2CERLKJA&user_name=roadrunner&command=%2Fwebhook-collect&text=&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2FT1DC2JH3J%2F397700885554%2F96rGlfmibIGlgcZRskXaIFfN&trigger_id=398738663015.47445629121.803a0bc887a14d10d2c447fce8b6703c`)

	request := httptest.NewRequest(http.MethodPost, "http://example.com/inbox", body)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("X-Slack-Signature", "v0=a2114d57b48eac39b9ad189dd8316235a7b4a8d21a10bd27519666489c69b503")
	request.Header.Set("X-Slack-Request-Timestamp", strconv.Itoa(int(time.Now().Unix()))) // correct timestamp => 1531420618
	w := httptest.NewRecorder()

	handler(w, request)

	response := w.Result()

	assert.Equal(t, 400, response.StatusCode)
}

func TestValidIncomingSlashCommand(t *testing.T) {
	os.Setenv("SIGNING_SECRET", "8f742231b10e8888abcd99yyyzzz85a5") // Example Signing Secret from https://api.slack.com/docs/verifying-requests-from-slack
	os.Setenv("VERIFY_TIMESTAMPS", "false")
	os.Setenv("WITH_VALIDATION", "true")

	handler := setupHandler()

	body := strings.NewReader(`token=xyzz0WbapA4vBCDEFasx0q6G&team_id=T1DC2JH3J&team_domain=testteamnow&channel_id=G8PSS9T3V&channel_name=foobar&user_id=U2CERLKJA&user_name=roadrunner&command=%2Fwebhook-collect&text=&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2FT1DC2JH3J%2F397700885554%2F96rGlfmibIGlgcZRskXaIFfN&trigger_id=398738663015.47445629121.803a0bc887a14d10d2c447fce8b6703c`)

	request := httptest.NewRequest(http.MethodPost, "http://example.com/inbox", body)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("X-Slack-Signature", "v0=a2114d57b48eac39b9ad189dd8316235a7b4a8d21a10bd27519666489c69b503")
	request.Header.Set("X-Slack-Request-Timestamp", "1531420618")
	w := httptest.NewRecorder()

	handler(w, request)

	response := w.Result()

	assert.Equal(t, 200, response.StatusCode)
}

func TestTimeStampValidationOnSlashCommands(t *testing.T) {
	os.Setenv("SIGNING_SECRET", "8f742231b10e8888abcd99yyyzzz85a5") // Example Signing Secret from https://api.slack.com/docs/verifying-requests-from-slack
	os.Setenv("VERIFY_TIMESTAMPS", "true")
	os.Setenv("WITH_VALIDATION", "true")

	handler := setupHandler()

	body := strings.NewReader(`token=xyzz0WbapA4vBCDEFasx0q6G&team_id=T1DC2JH3J&team_domain=testteamnow&channel_id=G8PSS9T3V&channel_name=foobar&user_id=U2CERLKJA&user_name=roadrunner&command=%2Fwebhook-collect&text=&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2FT1DC2JH3J%2F397700885554%2F96rGlfmibIGlgcZRskXaIFfN&trigger_id=398738663015.47445629121.803a0bc887a14d10d2c447fce8b6703c`)

	request := httptest.NewRequest(http.MethodPost, "http://example.com/inbox", body)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("X-Slack-Signature", "v0=a2114d57b48eac39b9ad189dd8316235a7b4a8d21a10bd27519666489c69b503")
	request.Header.Set("X-Slack-Request-Timestamp", "1531420618")
	w := httptest.NewRecorder()

	handler(w, request)

	response := w.Result()

	assert.Equal(t, 400, response.StatusCode)
}

func TestFullSlashCommands(t *testing.T) {
	os.Setenv("WITH_VALIDATION", "false")

	var callbackInvoked bool

	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callbackInvoked = true
		fmt.Fprintf(w, `{"message": "success"}`)
	}))

	repo := repository.NewInMemory()
	handler := setupHandlerWith(repo)

	body := strings.NewReader(`token=xyzz0WbapA4vBCDEFasx0q6G
	&team_id=T1DC2JH3J
	&team_domain=testteamnow
	&channel_id=G8PSS9T3V
	&channel_name=foobar
	&user_id=U2CERLKJA
	&user_name=roadrunner
	&command=%2Fresources
	&text=add+interestings+stack+overflow+http%3A%2F%2Fstackoverflow.com
	&response_url=` + url.QueryEscape(fakeServer.URL))

	request := httptest.NewRequest(http.MethodPost, "http://example.com/inbox", body)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()

	handler(w, request)

	response := w.Result()

	assert.Equal(t, 200, response.StatusCode)
	assert.Equal(t, true, callbackInvoked)

	u, _ := url.Parse("http://stackoverflow.com")

	expected := resourceful.Resource{
		Type:  "interestings",
		Title: "stack overflow",
		URL:   *u,
	}

	actual, err := repo.Fetch(*u)
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)

	actualResponseRaw, _ := ioutil.ReadAll(response.Body)
	response.Body.Close()

	type slashCommandResponse struct {
		Type        string `json:"response_type"`
		Text        string `json:"text"`
		Attachments []struct {
			Text string `json:"text"`
		} `json:"attachments"`
	}

	var actualResponse slashCommandResponse

	json.Unmarshal(actualResponseRaw, &actualResponse)

	expectedResponse := slashCommandResponse{
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
	assert.Equal(t, expectedResponse, actualResponse)
}
