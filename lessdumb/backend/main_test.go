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

var user User

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
  if query != u.Username {
    t.Errorf("The username was not created: got %v want %v", query, u.Username)
  }
}
