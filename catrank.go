package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type Cat struct {
	Name     string
	ImageURL string
	Votes    int
}

var cats = []Cat{
	{
		Name: "Susanna",
	},
	{
		Name: "Walter",
	},
	{
		Name: "Panpan",
	},
	{
		Name: "Busby",
	},
	{
		Name: "Keaton",
	},
}

var indexTemplate = template.Must(template.ParseFiles("templates/index.html"))

func main() {
	// Load the cat data from the JSON file
	err := loadCatData()
	if err != nil {
		fmt.Println("Error loading cat data:", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			catIndexStr := r.FormValue("cat")
			catIndex, err := strconv.Atoi(catIndexStr)
			if err == nil {
				cats[catIndex].Votes++

				// Save the updated cat data to the JSON file
				err = saveCatData()
				if err != nil {
					fmt.Println("Error saving cat data:", err)
				}
			}
		}

		err := indexTemplate.Execute(w, cats)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Fatal(http.ListenAndServe(":8000", nil))
}

func loadCatData() error {
	data, err := ioutil.ReadFile("cats.json")
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &cats)
	if err != nil {
		return err
	}

	return nil
}

func saveCatData() error {
	data, err := json.MarshalIndent(cats, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("cats.json", data, 0644)
	if err != nil {
		return err
	}

	return nil
}
