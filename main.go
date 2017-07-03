package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	plus "google.golang.org/api/plus/v1"
	"google.golang.org/api/plusdomains/v1"
)

var (
	oauthConfig = &oauth2.Config{
		ClientID:     "137566482663-pf3gl293a569tiqao8hfearldgcpcfv7.apps.googleusercontent.com",
		ClientSecret: "f0ZYHC_IOr26LM6RrIBNeVgn",
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost:9091/googleAuth",
		Scopes: []string{
			plusdomains.PlusMeScope,
			plusdomains.UserinfoProfileScope,
			plusdomains.UserinfoEmailScope,
			plusdomains.PlusCirclesReadScope,
			"https://www.googleapis.com/auth/contacts.readonly",
			"https://www.googleapis.com/auth/plus.circles.read",
		},
	}
	oauthStateString  = "hgvgvjubvgtcr"
	googleUserInfoURL = "https://www.googleapis.com/plus/v1/people/me?access_token="
	store             = sessions.NewCookieStore([]byte("SECRET"))
)

// home handles rendering of home page
func home(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-name")
	if err == nil && session.Values["access-token"] != nil {
		http.Redirect(w, r, "/info", http.StatusTemporaryRedirect)
		return
	}

	t, err := template.New("index.html").ParseFiles("./index.html")
	if err != nil {
		log.Println("An error occured: ", err)
	}
	t.Execute(w, nil)
	// http.ServeFile(w, r, r.URL.Path[1:]) -- another method of rendering html files
	// fmt.Fprintf(w, "welcome home") //data sent to client side
}

// googleLogin handles redirection to google service login
func googleLogin(w http.ResponseWriter, r *http.Request) {
	url := oauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// hadleGoogleCallback handles the authentication data received from google
func handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Println("An error occured when fetching the session ", err)
	}
	// set the state
	session.Values["state"] = oauthStateString
	session.Save(r, w)

	state := r.FormValue("state")
	if state != session.Values["state"].(string) {
		log.Printf("invalid oauth state, expected '%s', got '%s'\n", session.Values["state"], state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	// get authorization code and exchange it with access token
	token, err := oauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Printf("Error while exchanging code %v", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// check if the user is already authenticated
	storedToken := session.Values["access-token"]
	log.Println("token", session.Values["access-token"])

	if storedToken != nil {
		log.Println("user already authenticated")
	}

	// save for use later
	session.Values["access-token"] = token.AccessToken
	session.Save(r, w)
	http.Redirect(w, r, "/info", http.StatusTemporaryRedirect)
}

// HandleInfoDisplay makes calls to the google api to fetch the necessary info and display it
func handleInfoDisplay(w http.ResponseWriter, r *http.Request) {
	type Info struct {
		Activities     *plus.ActivityFeed
		Circles        *plusdomains.CircleFeed
		People         *plusdomains.PeopleFeed
		UserInfo       plusdomains.Person
		PeopleInCircle map[string][]*plusdomains.PeopleFeed
	}
	session, err := store.Get(r, "session-name")
	if err != nil || session.Values["access-token"] == nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	token := session.Values["access-token"].(string)
	log.Println("token", token)

	client := oauthConfig.Client(oauth2.NoContext, &oauth2.Token{AccessToken: token})

	// get user info
	response, err := http.Get(googleUserInfoURL + token)
	if err != nil {
		log.Println("Error occured when getting user info", err)
	}
	defer response.Body.Close()

	plusDomainService, _ := plusdomains.New(client)
	plusService, _ := plus.New(client)
	data := Info{}

	// get a list of activities done by the user
	data.Activities, err = plusService.Activities.List("me", "public").Do()
	if err != nil {
		log.Println("Error occured while fetching activities", err)
	}

	// get a list of circles assicoated with the user
	data.Circles, err = plusDomainService.Circles.List("me").Do()
	if err != nil {
		log.Println("Error occured while fetching circles", err)
	}

	data.PeopleInCircle = make(map[string][]*plusdomains.PeopleFeed)
	var peopleSlice []*plusdomains.PeopleFeed

	// get users per circle based on the circle ID provided
	for _, circle := range data.Circles.Items {
		data.People, err = plusDomainService.People.ListByCircle(circle.Id).Do()
		if err != nil {
			log.Println("Error occured while fetching people in circles", err)
		}
		data.PeopleInCircle[circle.DisplayName] = append(peopleSlice, data.People)

	}

	// decode the user info response to json
	err = json.NewDecoder(response.Body).Decode(&data.UserInfo)
	if err != nil {
		log.Println("Error occured while unmarshalling", err)
	}

	t, err := template.New("info.html").ParseFiles("info.html")
	if err != nil {
		log.Println("Error occured when parsing file")
	}

	log.Println("Data: ", data)

	err = t.Execute(w, data)

	log.Printf("Errror occurred %v", err)

}
func main() {
	http.HandleFunc("/", home) // setting the router to home handler
	http.HandleFunc("/googleLogin", googleLogin)
	http.HandleFunc("/googleAuth", handleGoogleCallback)
	http.HandleFunc("/info", handleInfoDisplay)

	err := http.ListenAndServe(":9091", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
