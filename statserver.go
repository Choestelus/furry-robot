package corgis

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
)

type StatusCtn struct {
	Status string
}

func HttpServe() {
	r := mux.NewRouter()
	r.HandleFunc("/", StatusHandler).Methods("POST")
	http.Handle("/", r)

	n := negroni.Classic()
	n.UseHandler(r)

	log.Fatal(http.ListenAndServe(":44005", r))
}

func StatusHandler(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var status StatusCtn
	err := decoder.Decode(&status)
	if err != nil {
		log.Printf("Status Handler Error: %v\n", err)
	}
	fmt.Fprintf(rw, "POSTed %v\n", status)
}
