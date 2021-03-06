package main

import (
    "time"
    "encoding/hex"
    "crypto/sha1"
    "regexp"
    "log"
    "net/http"
    "encoding/json"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "github.com/davecgh/go-spew/spew"
    "github.com/dgrijalva/jwt-go"
)

type NewUser struct {
  Username string
  Password string
  Email string
  Country string
  Birthday string
}

type User struct {
  Username string
  Password string
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type Session struct {
  Performance float32
  Level int
}

type SessionWithDate struct {
  Session
  Date string
}

var database *sql.DB
// JWT's key
var jwtKey = []byte("key")

func setup() {
  // Open the database
  var err error
  database, err = sql.Open("sqlite3", "./bogo.db")
  if err != nil {
    panic("Error in openning the database!")
  }
  // Prepare and execute the statement that will create the tables
  // if they don't exist
  query := `CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, username VARCHAR UNIQUE, password VARCHAR, email VARCHAR UNIQUE, country VARCHAR, birthday DATETIME, creationDate DATETIME);
            CREATE TABLE IF NOT EXISTS scores (id INTEGER PRIMARY KEY, userid INTEGER, performance FLOAT, date DATETIME, level TINYINT, FOREIGN KEY(userid) REFERENCES users(id));
            CREATE TABLE IF NOT EXISTS performance (id INTEGER PRIMARY KEY, userid INTEGER, average FLOAT, date DATETIME, FOREIGN KEY(userid) REFERENCES users(id));`
  statement, _ := database.Prepare(query)
  statement.Exec()
}

func CreateNewUser(w http.ResponseWriter, r *http.Request) {
  var u NewUser
  // Decodes the JSON Data from the front-end
  err := json.NewDecoder(r.Body).Decode(&u)
  if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  // Validates username field
  validation, _ := regexp.Match("^[A-Za-z0-9_]*$", []byte(u.Username))
  if len(u.Username) < 6 || len(u.Username) > 60  || !validation {
    http.Error(w, "Invalid username.", http.StatusBadRequest)
    return
  }
  // Checks if the username is being used
  var valid string
  _ = database.QueryRow("SELECT username FROM users WHERE username = ?", u.Username).Scan(&valid)
  if valid != "" {
    http.Error(w, "Username already in use.", http.StatusBadRequest)
    return
  }

  // Validates password field
  if len(u.Password) < 6 || len(u.Password) > 60 {
    http.Error(w, "Invalid password.", http.StatusBadRequest)
    return
  }

  // Creates a hash out of the password
  h := sha1.New()
  h.Write([]byte(u.Password))
  hashedPassword := hex.EncodeToString(h.Sum(nil))

  // Validates email field
  validation, _ = regexp.Match("^[a-zA-Z0-9\\-\\._]+@[a-zA-Z0-9\\-\\._]+\\.[a-zA-Z]+$", []byte(u.Email))
  if len(u.Email) > 256 || !validation {
    http.Error(w, "Invalid e-mail.", http.StatusBadRequest)
    return
  }
  // Checks if the e-mail already exists in the database
  valid = ""
  _ = database.QueryRow("SELECT email FROM users WHERE email = ?", u.Email).Scan(&valid)
  if valid != "" {
    http.Error(w, "E-mail already in use.", http.StatusBadRequest)
    return
  }
  // Validates the country
  if len(u.Country) >= 150 {
    http.Error(w, "Invalid country.", http.StatusBadRequest)
    return
  }
  // Validates the birthday
  validation, _ = regexp.Match("^\\d{2,2}-\\d{2,2}-\\d{4,4}$", []byte(u.Birthday))
  if len(u.Birthday) != 10 || !validation {
    http.Error(w, "Invalid birthday.", http.StatusBadRequest)
    return
  }
  // Parses birthday to time.time format
  birthday, _ := time.Parse("02-02-2006", u.Birthday)
  // Defines account's creation time
  creationTime := time.Now()
  // Inserts into the database the newuser
  statement, _ := database.Prepare("INSERT INTO users(username, password, email, country, birthday, creationDate) VALUES (?, ?, ?, ?, ?, ?)")
  _, err = statement.Exec(u.Username, hashedPassword, u.Email, u.Country, birthday, creationTime)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
		return
  }
}

func Login(w http.ResponseWriter, r *http.Request) {
  var u User
  // Decodes the JSON Data from the front-end
  err := json.NewDecoder(r.Body).Decode(&u)
  if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }
  // Check if the username is valid
  validation, _ := regexp.Match("^[A-Za-z0-9_]*$", []byte(u.Username))
  if len(u.Username) < 6 || len(u.Username) > 60  || !validation {
    http.Error(w, "Invalid username.", http.StatusBadRequest)
    return
  }
  // Checks if the username exists
  var valid string
  _ = database.QueryRow("SELECT username FROM users WHERE username = ?", u.Username).Scan(&valid)
  if valid == "" {
    http.Error(w, "Username doesn't exist.", http.StatusBadRequest)
    return
  }

  // Make a hash from the incoming password
  h := sha1.New()
  h.Write([]byte(u.Password))
  hashedPassword := hex.EncodeToString(h.Sum(nil))
  // Query the user's password from the database
  _ = database.QueryRow("SELECT password FROM users WHERE username = ?", u.Username).Scan(&valid)
  // Compares the hashes to see if they match
  if valid != hashedPassword {
    http.Error(w, "Wrong password.", http.StatusBadRequest)
    return
  }
  // Defines the JWT lifetime
  expirationTime := time.Now().Add(31 * time.Second)
  // ??
  claims := &Claims{
		Username: u.Username,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}
  // Creates JWT with HS256 & Claims
  token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
  // Creates JWT string
	tokenString, err := token.SignedString(jwtKey)
  if err != nil {
		// If there is an error in creating the JWT return an internal server error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
  // Sets the cookie
  http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

  spew.Dump(u)
}

func RefreshToken(w http.ResponseWriter, r *http.Request) {
  // Obatins the token from the request
  c, err := r.Cookie("token")
  if err != nil {
    if err == http.ErrNoCookie {
    // If the cookie is not set, return an unauthorized status
      w.WriteHeader(http.StatusUnauthorized)
      return
    }
  // For any other type of error, return a bad request status
    w.WriteHeader(http.StatusBadRequest)
  return
  }

  // Get the JWT string from the cookie
	tknStr := c.Value

  // Initialize a new instance of `Claims`
  claims := &Claims{}

  // Parse the token
  tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
  // If the token expiration time lasts more than 30 seconds
  // Sends an bad request Error
  if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 15 * time.Second {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

  // Now, create a new token for the current use, with a renewed expiration time
	expirationTime := time.Now().Add(15 * time.Minute)
	claims.ExpiresAt = expirationTime.Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the new token as the users `token` cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})
}

func AddSession(w http.ResponseWriter, r *http.Request) {

  // Obtains the token from the request
  c, err := r.Cookie("token")
  if err != nil {
    if err == http.ErrNoCookie {
    // If the cookie is not set, return an unauthorized status
      w.WriteHeader(http.StatusUnauthorized)
      return
    }
  // For any other type of error, return a bad request status
    w.WriteHeader(http.StatusBadRequest)
  return
  }

  // Get the JWT string from the cookie
  tknStr := c.Value

  // Initialize a new instance of `Claims`
  claims := &Claims{}

  // Parse the token
  tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
    return jwtKey, nil
  })
  if err != nil {
    if err == jwt.ErrSignatureInvalid {
      w.WriteHeader(http.StatusUnauthorized)
      return
    }
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  if !tkn.Valid {
    w.WriteHeader(http.StatusUnauthorized)
    return
  }

  var session Session

  // Decodes the JSON Data from the front-end
  err = json.NewDecoder(r.Body).Decode(&session)
  if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  var (
    count int
    lastDate string
    performance float32
    userID int
  )
  // Checks how many rows there are in the database
  err = database.QueryRow("SELECT COUNT(*) FROM scores JOIN users ON users.id=scores.userid WHERE users.username = ?", claims.Username).Scan(&count)
  // Retrives the userID
  database.QueryRow(`SELECT ID FROM users WHERE username = ?`, claims.Username).Scan(&userID)
  // Checks if there is an entry for today's performance
  database.QueryRow(`SELECT average, date FROM performance WHERE userid=? ORDER BY date DESC LIMIT 1`, userID).Scan(&performance, &lastDate)

  switch {
    // It is not the first entry of the day
    // So calculate the average and store it
  case lastDate != "" && lastDate[:10] == time.Now().String()[:10]:
      var c int
      database.QueryRow(`SELECT COUNT(*) FROM performance WHERE userid=?`, userID).Scan(&c)
      if newAverage := (performance + (session.Performance / 100) + float32(session.Level)) / 2; c >= 30 {
          var id int
          // Gets the entry's ID for UPDATING
          database.QueryRow(`SELECT id FROM performance WHERE userID=? ORDER BY date LIMIT 1`, userID).Scan(&id)
          // Updates the oldest
          statement, _ := database.Prepare(`UPDATE performance SET average = ?, date = ? WHERE id = ?`)
          statement.Exec(newAverage, time.Now())
      } else {
        // If there is less than 30 entries, creates a new one
        statement, _ := database.Prepare(`INSERT INTO performance (userid, average, date) VALUES (?, ?, ?)`)
        statement.Exec(userID, newAverage, time.Now())
      }
    // If it is the first entry of the day
  case lastDate == "" || lastDate[:10] != time.Now().String()[:10]:
      var c int
      database.QueryRow(`SELECT COUNT(*) FROM performance WHERE userid=?`, userID).Scan(&c)
      // If the database holds more than 30 entries then update it
      if c >= 30 {
        var id int
        database.QueryRow(`SELECT id FROM performance WHERE userid=? ORDER BY date DESC LIMIT 1`, userID).Scan(&id)
        statement, _ := database.Prepare(`UPDATE performance SET average = ?, date = ? WHERE id = ?`)
        statement.Exec(session.Performance / 100 + float32(session.Level), time.Now())
      } else {
        // Otherwise creates a new entry
        statement, _ := database.Prepare(`INSERT INTO performance (userid, average, date) VALUES (?, ?, ?)`)
        statement.Exec(userID, session.Performance / 100 + float32(session.Level), time.Now())
      }
  }

  switch {
    case err != nil:
      w.WriteHeader(http.StatusInternalServerError)
      return
    case count < 10:
      // If there are less than 10 entries, add more
      statement, _ := database.Prepare(`INSERT INTO scores (userID, performance, date, level) VALUES (?, ?, ?, ?)`)
      statement.Exec(userID, session.Performance, time.Now(), session.Level)
    case count >= 10:
      // If there are more or equal than ten entries, update the oldest one
      statement, _ := database.Prepare(`UPDATE scores SET performance = ?, date = ?, level = ? WHERE id=(SELECT Id FROM scores WHERE userid=? ORDER BY date LIMIT 1)`)
      statement.Exec(session.Performance, time.Now(), session.Level, userID)
  }

}

func RetrieveGraph(w http.ResponseWriter, r *http.Request) {
  // Obtains the token from the request
  c, err := r.Cookie("token")
  if err != nil {
    if err == http.ErrNoCookie {
    // If the cookie is not set, return an unauthorized status
      w.WriteHeader(http.StatusUnauthorized)
      return
    }
  // For any other type of error, return a bad request status
    w.WriteHeader(http.StatusBadRequest)
  return
  }

  // Get the JWT string from the cookie
	tknStr := c.Value

  // Initialize a new instance of `Claims`
  claims := &Claims{}

  // Parse the token
  tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
  // Creates a slice of an anonymous struct which will be converted
  // To a JSON
  graph := []struct {
    average float32
    date string
  }{}

  var userID int
  database.QueryRow(`SELECT id FROM users WHERE username = ?`, claims.Username).Scan(&userID)
  rows, _ := database.Query(`SELECT average, date FROM performance WHERE userID = ? ORDER BY date`, userID)

  for rows.Next() {
    g := struct {
      average float32
      date string
    }{}
    // Map the values to the struct
    rows.Scan(&g.average, &g.date)
    // Append the value to the graph slice
    graph = append(graph, g)
  }
  jsonData, _ := json.Marshal(graph)
  w.Write(jsonData)
}

func RetrieveSessions(w http.ResponseWriter, r *http.Request) {
  // Obtains the token from the request
  c, err := r.Cookie("token")
  if err != nil {
    if err == http.ErrNoCookie {
    // If the cookie is not set, return an unauthorized status
      w.WriteHeader(http.StatusUnauthorized)
      return
    }
  // For any other type of error, return a bad request status
    w.WriteHeader(http.StatusBadRequest)
  return
  }

  // Get the JWT string from the cookie
	tknStr := c.Value

  // Initialize a new instance of `Claims`
  claims := &Claims{}

  // Parse the token
  tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

  // Query the user's sessions data from the database
  rows, err := database.Query(`SELECT scores.performance, scores.date, scores.level FROM scores JOIN users ON users.id=scores.userid WHERE users.username = ? ORDER BY (date)`, claims.Username)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
  defer rows.Close()
  // Declare a slice of struct(Session)
  var sessions []SessionWithDate
  // Iterates over the rows
  for rows.Next() {
    var s SessionWithDate
    if err := rows.Scan(&s.Performance, &s.Date, &s.Level); err != nil {
      w.WriteHeader(http.StatusInternalServerError)
      return
    }
    sessions = append(sessions, s)
  }

  // Converts struct into json
  jsonData, _ := json.Marshal(sessions)
  w.Write(jsonData)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
  // Obtains the token from the request
  c, err := r.Cookie("token")
  if err != nil {
    if err == http.ErrNoCookie {
    // If the cookie is not set, return an unauthorized status
      w.WriteHeader(http.StatusUnauthorized)
      return
    }
  // For any other type of error, return a bad request status
    w.WriteHeader(http.StatusBadRequest)
  return
  }
  // Get the JWT string from the cookie
  tknStr := c.Value
  // Initialize a new instance of `Claims`
  claims := &Claims{}

  // Parse the token
  tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
    return jwtKey, nil
  })
  if err != nil {
    if err == jwt.ErrSignatureInvalid {
      w.WriteHeader(http.StatusUnauthorized)
      return
    }
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  if !tkn.Valid {
    w.WriteHeader(http.StatusUnauthorized)
    return
  }
  var userID int
  database.QueryRow(`SELECT id FROM users WHERE username = ?`, claims.Username).Scan(&userID)
  // Deletes from the users table
  statement, _ := database.Prepare(`DELETE FROM users WHERE id = ?`)
  statement.Exec(userID)
  // Deletes from the scores table
  statement, _ = database.Prepare(`DELETE FROM scores WHERE userid = ?`)
  statement.Exec(userID)
  // Deletes from the performance table
  statement, _ = database.Prepare(`DELETE FROM performance WHERE userID = ?`)
  statement.Exec(userID)
}

func main() {
  // Connects to the database
  setup()
  defer database.Close()

  mux := http.NewServeMux()
  mux.HandleFunc("/user/create", CreateNewUser)
  mux.HandleFunc("/user/delete", DeleteUser)
  mux.HandleFunc("/user/login", Login)
  mux.HandleFunc("/user/refresh", RefreshToken)
  mux.HandleFunc("/data/sessions", RetrieveSessions)
  mux.HandleFunc("/data/addsession", AddSession)
  mux.HandleFunc("/data/graph", RetrieveGraph)

  err := http.ListenAndServe(":4000", mux)
  if err != nil {
    log.Fatal(err)
  }
}
