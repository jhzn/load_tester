package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func dummyJson(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		type person struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		kalle := &person{
			Name: "Kalle",
			Age:  30,
		}
		j, err := json.Marshal(kalle)
		if err != nil {
			log.Fatal(err)
		}
		_, err = w.Write(j)
		if err != nil {
			log.Fatal(err)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Invalid http method")
	}

	r.Close = true
}

func echo(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "POST":
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		w.Write(bodyBytes)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Invalid http method")
	}
}

func main() {
	http.HandleFunc("/json", dummyJson)
	http.HandleFunc("/echo", echo)

	log.Println("Server started!")
	http.ListenAndServe(":8080", nil)
}
