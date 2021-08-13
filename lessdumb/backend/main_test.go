package main

import (
  "math/rand"
  "time"
  "encoding/json"
  "net/http/httptest"
  "net/http"
  "testing"
  "bytes"

)

var (
  user User
  cookie *http.Cookie
)

const letterBytes = "1234567890_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
    b := make([]byte, n)
    for i := range b {
        b[i] = letterBytes[rand.Intn(len(letterBytes))]
    }
    return string(b)
}

func TestCreateNewUser(t *testing.T) {
  setup()
  defer database.Close()
  // Sets the seed of rand to time now
  rand.Seed(time.Now().UnixNano())
  var u NewUser
  // Fills the NewUser's structure with the random data
  u.Username = RandStringBytes(rand.Intn(60 - 6) + 6)
  u.Password = RandStringBytes(rand.Intn(60 - 6) + 6)
  u.Email = RandStringBytes(rand.Intn(60 - 3) + 3) + "@lessdumb.org"
  u.Country = "Brazil"
  u.Birthday = "02-02-2006"
  t.Log("\nUser: ", u.Username,
              "\nPassword: ", u.Password,
              "\nEmail: ", u.Email,
              "\nCountry: ", u.Country,
              "\nBirthday: ", u.Birthday,
  )
  jsonStr, err := json.Marshal(&u)

  // Creates the request
  req, err := http.NewRequest("POST", "http://localhost:4000/user/create", bytes.NewBuffer(jsonStr))
  if err != nil {
    t.Fatal(err)
  }
  // Sets the header
  req.Header.Set("Content-Type", "application/json")
  // Creates a recorder
  rr := httptest.NewRecorder()
  // ??
  handler := http.HandlerFunc(CreateNewUser)
  // Hits the API's endpoint
  handler.ServeHTTP(rr, req)
  // Checks if the httprequest's status is OK
  if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
  // Assigns the variables which will be used to login
  user.Username = u.Username
  user.Password = u.Password
  // Access the database and checks if the user was created
  var query string
  err = database.QueryRow(`SELECT username FROM users WHERE username = ?`, u.Username).Scan(&query)
  // Checks if the username retrived from the DB
  // is the same that the generated
  if query != u.Username {
    t.Errorf("The username was not created: got %v want %v", query, u.Username)
  }
}

func TestLogin(t *testing.T) {
  setup()
  defer database.Close()
  // Creates a JSON str to do the login
  jsonStr, err := json.Marshal(&user)
  if err != nil {
    t.Errorf("Error creating JSON from user struct.")
  }

  req, err := http.NewRequest("POST", "http://localhost:4000/user/login", bytes.NewBuffer(jsonStr))
  // Sets the header
  req.Header.Set("Content-Type", "application/json")
  // Creates a recorder
  rr := httptest.NewRecorder()
  // ??
  handler := http.HandlerFunc(Login)
  // Hits the API's endpoint
  handler.ServeHTTP(rr, req)
  // Checks if the httprequest's status is OK
  if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
  // Gets the http response from http request
  httpResponse := rr.Result()
  // Extract the cookies from
  cookies := httpResponse.Cookies()
  // Saves the cookie for further testing
  cookie = cookies[0]
}

func TestRefreshToken(t *testing.T) {
  // Creates the POST request
  req, err := http.NewRequest("POST", "http://localhost:4000/user/refresh", nil)
  if err != nil {
    panic("Error creating /user/refresh request.")
  }
  // Adds the cookie to the req
  req.AddCookie(cookie)
  // Creates a recorder
  rr := httptest.NewRecorder()
  // ??
  handler := http.HandlerFunc(RefreshToken)
  // Puts the execution to wait
  time.Sleep(30 * time.Second)
  // Hits the API's endpoint
  handler.ServeHTTP(rr, req)
  if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
  // Updates cookie = refreshed cookie
  httpResponse := rr.Result()
  cookies := httpResponse.Cookies()
  if cookies == nil {
    t.Errorf("New cookie was not received.")
  }
  cookie = cookies[0]
}

func TestAddSession(t *testing.T) {
  setup()
  defer database.Close()
  for i := 0; i < 35; i++ {
    if i == 10 {
      // Wait for 5 seconds and test the update function
      time.Sleep(5 * time.Second)
    }
    var session Session
    // Permeates session's structure with random data
    session.Performance = 1 + rand.Float32() * 100
    session.Level = rand.Intn(99) + 1
    // Converts to JSON
    jsonStr, _ := json.Marshal(&session)
    // Creates the request
    req, err := http.NewRequest("POST", "http://localhost:4000/data/addsession", bytes.NewBuffer(jsonStr))
    if err != nil {
      panic("Error creating /data/addsession request.")
    }
    // Sets the header
    req.Header.Set("Content-Type", "application/json")
    // Adds cookie
    req.AddCookie(cookie)
    // Creates a recorder
    rr := httptest.NewRecorder()
    // ??
    handler := http.HandlerFunc(AddSession)
    // Hits the API's endpoint
    handler.ServeHTTP(rr, req)
    // Checks if the httprequest's status is OK
    if status := rr.Code; status != http.StatusOK {
  		t.Errorf("Handler returned wrong status code: got %v want %v",
  			status, http.StatusOK)
  	}
  }
}

func TestRetrieveSessions(t *testing.T) {
  setup()
  defer database.Close()
  // Creates the request
  req, err := http.NewRequest("POST", "http://localhost:4000/data/sessions", nil)
  if err != nil {
    panic("Error creating /data/sessions request.")
  }
  // Adds cookie
  req.AddCookie(cookie)
  // Creates a recorder
  rr := httptest.NewRecorder()
  // ??
  handler := http.HandlerFunc(RetrieveSessions)
  // Hits the API's endpoint
  handler.ServeHTTP(rr, req)
  // Checks if the httprequest's status is OK
  if status := rr.Code; status != http.StatusOK {
    t.Errorf("Handler returned wrong status code: got %v want %v",
      status, http.StatusOK)
  }
  // Checks if the JSON was received
  if rr.Body == nil {
    t.Errorf("JSON wasn't received.")
  }
}

func TestRetrieveGraph(t *testing.T) {
  setup()
  defer database.Close()
  req, err := http.NewRequest("POST", "http://localhost:4000/data/sessions", nil)
  if err != nil {
    panic("Error creating /data/graph request.")
  }
  // Adds cookie
  req.AddCookie(cookie)
  // Creates a recorder
  rr := httptest.NewRecorder()
  // ??
  handler := http.HandlerFunc(RetrieveGraph)
  // Hits the API's endpoint
  handler.ServeHTTP(rr, req)
  // Checks if the httprequest's status is OK
  if status := rr.Code; status != http.StatusOK {
    t.Errorf("Handler returned wrong status code: got %v want %v",
      status, http.StatusOK)
  }
  // Checks if the JSON was received
  if rr.Body == nil {
    t.Errorf("JSON wasn't received.")
  }

}

func TestDeleteUser(t *testing.T) {
  setup()
  defer database.Close()
  // Creates the request
  req, err := http.NewRequest("POST", "http://localhost:4000/data/sessions", nil)
  if err != nil {
    panic("Error creating /data/sessions request.")
  }
  // Adds cookie
  req.AddCookie(cookie)
  // Creates a recorder
  rr := httptest.NewRecorder()
  // ??
  handler := http.HandlerFunc(DeleteUser)
  // Hits the API's endpoint
  handler.ServeHTTP(rr, req)
  // Checks if the httprequest's status is OK
  if status := rr.Code; status != http.StatusOK {
    t.Errorf("Handler returned wrong status code: got %v want %v",
      status, http.StatusOK)
  }
}
