package corgis

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/fatih/color"
	"github.com/gorilla/mux"
)

var mutex sync.Mutex

type StatusCtn struct {
	Status string
	From   string
}

func HttpServe() {
	r := mux.NewRouter()
	r.HandleFunc("/status", StatusHandler).Methods("POST")
	http.Handle("/", r)

	//n := negroni.Classic()
	lightBlue := color.New(color.FgCyan).Add(color.Bold).SprintfFunc()
	// colorLogger := &negroni.Logger{log.New(os.Stdout, "[negroni] ", 0)}
	colorLogger := &negroni.Logger{log.New(os.Stdout, lightBlue("%v [negroni] ", time.Now().String()), 0)}

	n := negroni.New(negroni.NewRecovery(), colorLogger)
	n.UseHandler(r)

	log.Println(lightBlue("Listening on port 44005"))
	log.Fatal(http.ListenAndServe(":44005", n))
}

func StatusHandler(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var status StatusCtn
	err := decoder.Decode(&status)
	if err != nil {
		log.Printf("Status Handler Error: %v\n", err)
	}
	fmt.Fprintf(rw, "got POSTed %v\n", status)
	if status.Status == "ok" {
		if status.From == "HDD" {
			mutex.Lock()
			updatingFlagHDD = false
			mutex.Unlock()
		} else if status.From == "SSD" {
			mutex.Lock()
			updatingFlagSSD = false
			mutex.Unlock()
		}
	}
	log.Printf("----> clearflag: uFHDD == %v uFSSD == %v\n", updatingFlagHDD, updatingFlagSSD)
}

func CallPostmark(where string) {
	monSSDurl := "http://192.168.122.11:44005/start"
	monHDDurl := "http://192.168.122.252:44005/start"
	var url string
	var body []byte
	if where == "HDD" {
		url = monHDDurl
	} else if where == "SSD" {
		url = monSSDurl
	}
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("calling postmark %v error: %v\n", where, err)
		goto EndFunc
	}

	body, _ = ioutil.ReadAll(resp.Body)
	fmt.Printf("status response:%v %v\nHeader: %v\nbody: [%v]\n", resp.StatusCode, resp.Status, resp.Header, body)
	if where == "HDD" {
		mutex.Lock()
		updatingFlagHDD = true
		mutex.Unlock()
	} else if where == "SSD" {
		mutex.Lock()
		updatingFlagSSD = true
		mutex.Unlock()
	}
	log.Printf("----> setflag: uFHDD == %v uFSSD == %v\n", updatingFlagHDD, updatingFlagSSD)
	defer resp.Body.Close()
EndFunc:
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Recovered in CallPostmark: %v\n", err)
		}
	}()
}
