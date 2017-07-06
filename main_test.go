package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHomeHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	// record the response
	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(home)

	handler.ServeHTTP(responseRecorder, req)

	// check content type is text/html
	if contentType := responseRecorder.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Errorf("wrong Content-Type: expected %v but got %v", "text/html; charset=utf-8", contentType)
	}

	// check response body is what we expected
	expected := "Log in with Google"
	if !strings.Contains(responseRecorder.Body.String(), expected) {
		t.Errorf("wrong response: expected %v but got %v", expected, responseRecorder.Body.String())
	}

}

func TestGoogleLoginHandlerRedirectsToGoogleOauth(t *testing.T) {
	req, err := http.NewRequest("GET", "googleLogin", nil)
	if err != nil {
		t.Fatal(err)
	}
	// record the response
	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(googleLogin)

	handler.ServeHTTP(responseRecorder, req)

	//check the status code is 307
	if status := responseRecorder.Code; status != http.StatusTemporaryRedirect {
		t.Errorf("handler returned wrong status code: expected %v but got %v", http.StatusTemporaryRedirect, status)
	}

	// check the redirect url contains https://accounts.google.com/o/oauth2/auth
	expected := "https://accounts.google.com/o/oauth2/auth"
	if !strings.Contains(responseRecorder.Body.String(), expected) {
		t.Errorf(" redirect to googleOauth failed: expected %v but got %v", expected, responseRecorder.Body.String())
	}
}

func TestGoogleCallbackHandlerRedirectsForMisMatchState(t *testing.T) {
	req, err := http.NewRequest("GET", "googleAuth", nil)
	if err != nil {
		t.Fatal(err)
	}
	// record the response
	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(handleGoogleCallback)

	handler.ServeHTTP(responseRecorder, req)

	// check redirection to homepage
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("handler returned wrong status code: expected %v but got %v", http.StatusTemporaryRedirect, responseRecorder.Code)
	}
}

func TestGoogleCallbackHandlerRedirectsOnTokenExchangeError(t *testing.T) {
	req, err := http.NewRequest("GET", "/googleAuth?state=gfdrswqrfcAWED", nil)
	if err != nil {
		t.Fatal(err)
	}
	// record the response
	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(handleGoogleCallback)

	handler.ServeHTTP(responseRecorder, req)

	// check redirection to homepage
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("handler returned wrong status code: expected %v but got %v", http.StatusTemporaryRedirect, responseRecorder.Code)
	}
	log.Println(responseRecorder.Body.String())

}

func TestGoogleCallbackHandlerRedirectsToInfoPageOnSuccess(t *testing.T) {}

func TestHomeHandlerRedirectsToInfoPage(t *testing.T) {

}
func TestHandleInfoDisplayRedirectsForEmptySession(t *testing.T) {}

func TestHandleInfoDisplay(t *testing.T) {}
