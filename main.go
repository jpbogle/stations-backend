package main

import (
	"log"
	"net/http"
	"os"
	"stations/controllers"
	"github.com/kabukky/httpscerts"
	h "stations/handlers"
    "stations/utils"
	"sync"
    // "github.com/fvbock/endless"
)

var IS_DEBUG = false

func main() {
	defer controllers.Close()
	log.SetFlags(0)

	parseArgs()

	var wg sync.WaitGroup
	wg.Add(1)

	// createCerts()  //Create certs for https
    // redirect every http request to https
    // go http.ListenAndServe(":8080", http.HandlerFunc(redirect))
    // serve index (and anything else) as https
    // mux := http.NewServeMux()
    // mux.HandleFunc("/", h.RootHandler.ServeHTTP)
    go http.ListenAndServe(":8080", h.RootHandler)

	// go endless.ListenAndServe(":8080", h.RootHandler)
	log.Print("\n> Now listening on localhost:8080\n\n")
	wg.Wait()
}

func parseArgs() {
	for _, v := range os.Args {
		if v == "--debug" {
			utils.SetDebug(true)
		}
		if v == "--drop" && IS_DEBUG {
			controllers.DropTables()
			log.Println("Successfully dropped tables")
			os.Exit(0)
		}
	}
}

func createCerts() {
    err := httpscerts.Check("cert.pem", "key.pem")
    // If they are not available, generate new ones.
    if err != nil {
        err = httpscerts.Generate("cert.pem", "key.pem", "127.0.0.1:8081")
        if err != nil {
            log.Fatal("Error: Couldn't create https certs.")
        }
    }
}

func redirect(w http.ResponseWriter, req *http.Request) {
    // remove/add not default ports from req.Host
    target := "https://" + req.Host + req.URL.Path
    if len(req.URL.RawQuery) > 0 {
        target += "?" + req.URL.RawQuery
    }
    log.Printf("redirect to: %s", target)
    http.Redirect(w, req, target, http.StatusMovedPermanently)
}
