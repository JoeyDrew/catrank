package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"syscall"
	"time"
)

type Cat struct {
	Name     string
	ImageURL string
	Votes    int
}

var (
	//go:embed cats.json
	initialCatsData []byte

	indexTemplate = template.Must(template.ParseFiles("templates/index.html"))
)

func main() {
	log.Print("info starting")

	var cats []Cat
	var err error

	catsDataPath, ok := os.LookupEnv("CATS_DATA_PATH")
	if !ok {
		catsDataPath = "/data/cats.json"
	}

	// Load the cat data from the JSON file or exit
	if cats, err = initCatData(catsDataPath); err != nil {
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

			// Sort cats prior to saving
			sort.Slice(cats, func(i, j int) bool {
				return cats[i].Votes > cats[j].Votes
			})

			// Save the updated cat data to the JSON file
			if err = saveCatData(cats, catsDataPath); err != nil {
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
		log.Print("serving on :8080")
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

func initCatData(catsDataPath string) ([]Cat, error) {
	cats, err := loadCatData(catsDataPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err = os.WriteFile(catsDataPath, initialCatsData, 0644); err != nil {
				return nil, fmt.Errorf("failed to initialize cats data: %w", err)
			}

			// Try to load from the file system again. If it fails, there's probably something weird going
			// on with the file system (e.g., we have write permission but not read permission)
			if cats, err = loadCatData(catsDataPath); err != nil {
				return nil, fmt.Errorf("failed to load cats data after initialization: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to load initial cats data: %w", err)
		}
	}

	// Make sure we can actually write to the file as well
	file, err := os.OpenFile(catsDataPath, os.O_WRONLY, 0)
	if err != nil {
		if os.IsPermission(err) {
			return nil, fmt.Errorf("no permission to write to cats data path \"%s\": %w", catsDataPath, err)
		}
		return nil, fmt.Errorf("failed to check write permission to cats data path \"%s\": %w", catsDataPath, err)
	}
	file.Close()

	return cats, nil
}

func loadCatData(catsDataPath string) ([]Cat, error) {
	var cats []Cat

	data, err := os.ReadFile(catsDataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cats json file: %w", err)
	}

	if err = json.Unmarshal(data, &cats); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cats data: %w", err)
	}

	return cats, nil
}

func saveCatData(cats []Cat, catsDataPath string) error {
	data, err := json.MarshalIndent(cats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cats json data: %w", err)
	}

	if err = os.WriteFile(catsDataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to save cats json file: %w", err)
	}

	return nil
}
