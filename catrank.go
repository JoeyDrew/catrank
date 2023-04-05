package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
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
		Name: "Pan-Pan",
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
	log.Print("info starting")
	// Load the cat data from the JSON file or exit
	if err := loadCatData(); err != nil {
		log.Fatalf("error loading cat data: %v", err)
	}
	log.Print("info cat data loaded")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Print("debug handling request")
		defer log.Print("debug request processed")
		switch r.Method {
		case "GET":
			if err := indexTemplate.Execute(w, cats); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case "POST":
			catIndexStr := r.FormValue("cat")
			catIndex, err := strconv.Atoi(catIndexStr)
			if err != nil {
				log.Printf("warn parsing cat index: %s, err: %v", catIndexStr, err)
				http.Error(w, "invalid index", http.StatusBadRequest)
				return
			}

			cats[catIndex].Votes++
			log.Printf("info vote for %s", cats[catIndex].Name)

			// Save the updated cat data to the JSON file
			if err = saveCatData(); err != nil {
				log.Printf("error saving cat data: %v", err)
			}
			log.Print("info cat data saved")

			// Redirect to a regular GET request to display results and prevent browser from duplicating votes
			http.Redirect(w, r, r.URL.String(), http.StatusSeeOther)
		default:
			// Received an unsupported http method. Let them know
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	// Serve CSS file
	http.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "style.css")
	})

	// Handle SIGINT and SIGTERM signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start http server
	srv := &http.Server{Addr: ":8080"}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("error listen: %v", err)
		}
	}()

	<-sigCh

	// Received signal to shut down. Let's clean up a little
	log.Print("info stopping server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("error shutting down server: %v", err)
	}
	log.Print("info done")
}

func loadCatData() error {
	data, err := os.ReadFile("cats.json")
	if err != nil {
		return fmt.Errorf("failed to read cats json file: %w", err)
	}

	if err = json.Unmarshal(data, &cats); err != nil {
		return fmt.Errorf("failed to unmarshal cats data: %w", err)
	}

	return nil
}

func saveCatData() error {
	data, err := json.MarshalIndent(cats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cats json data: %w", err)
	}

	if err = os.WriteFile("cats.json", data, 0644); err != nil {
		return fmt.Errorf("failed to save cats json file: %w", err)
	}

	return nil
}
