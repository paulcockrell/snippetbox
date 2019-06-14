package main

import (
        "fmt"
	/* "crypto/tls" */
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/paulcockrell/snippetbox/pkg/models"
	"github.com/paulcockrell/snippetbox/pkg/models/mysql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golangcollege/sessions"
)

type contextKey string

var contextKeyUser = contextKey("user")

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	session  *sessions.Session
	snippets interface {
		Insert(string, string, string) (int, error)
		Get(int) (*models.Snippet, error)
		Latest() ([]*models.Snippet, error)
	}
	templateCache map[string]*template.Template
	users         interface {
		Insert(string, string, string) error
		Authenticate(string, string) (int, error)
		Get(int) (*models.User, error)
	}
}

func getEnv(key, fallback string) string {
  if value, ok := os.LookupEnv(key); ok {
    return value
  }
  return fallback
}

func main() {
        // Command line address argument was from tutorial, now converted to use PORT address
        // for Heroku deployment
	// addr := flag.String("addr", ":4000", "HTTP network address")
        port := getEnv("PORT", "4000")
        addr := fmt.Sprintf(":%s", port)

        // Command line dsn argument was from tutorial, now converted to use Heroku mysql
	// dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "mySQL data source name")
        //dsn := getEnv("JAWSDB_URL", "web:pass@/snippetbox?parseTime=true")
        dsn := "fqv6f91u1qxbzrk0:zpq1z8g72wzyjd1j@tcp(b8rg15mwxwynuk9q.chr7pe7iynqr.eu-west-1.rds.amazonaws.com:3306)/hazz8ovy0wr03hvu"

	secret := flag.String("secret", "s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge", "Secret key")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB(dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	templateCache, err := newTemplateCache("./ui/html")
	if err != nil {
		errorLog.Fatal(err)
	}

	session := sessions.New([]byte(*secret))
	session.Lifetime = 12 * time.Hour
	session.Secure = true
	session.SameSite = http.SameSiteStrictMode

	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		session:       session,
		snippets:      &mysql.SnippetModel{DB: db},
		users:         &mysql.UserModel{DB: db},
		templateCache: templateCache,
	}

        /*
	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
	}
        */

	srv := &http.Server{
		Addr:         addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		/* TLSConfig:    tlsConfig, */
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Starting server on %s", addr)
        // This was to run locally with SSL but not needed as Heroku provides this
        // for us, might want a way to switch between this for local use and without
        // for prod deploy
        //
	/* srvErr := srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem") */

        srvErr := srv.ListenAndServe()
	errorLog.Fatal(srvErr)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
