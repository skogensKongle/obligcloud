package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

var url = "https://api.fixer.io/latest"

type fromFixer struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float32 `json:"rates"`
}

var getShit fromFixer

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("listening...")
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}

func handler(res http.ResponseWriter, req *http.Request) {
	//Fetch response form url1

	repo, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer repo.Body.Close()

	//Grab body from Response

	body, err := ioutil.ReadAll(repo.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(body, &getShit)
	if err != nil {
		panic(err)
	}

	json.NewEncoder(res).Encode(getShit)
}
