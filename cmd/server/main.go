package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
)

// global state to hold our dataframe, protected by a mutex
// for safe concurrent access from multiple api calls
var (
	df      dataframe.DataFrame
	dfMutex = &sync.Mutex{}
)

// loadHandler loads a CSV file into the global dataframe.
// expects a JSON body like: {"path": "/path/to/my/file.csv"}

func loadHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad request body", http.StatusBadRequest)
		return
	}
	file, err := os.Open(payload.Path)
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}

	loadedDf := dataframe.ReadCSV(file)
	if loadedDf.Error() != nil {
		http.Error(w, "Failed to read CSV: "+loadedDf.Error().Error(), http.StatusInternalServerError)
		return
	}

	// lock the mutex for writing
	dfMutex.Lock()
	df = loadedDf
	dfMutex.Unlock()

	log.Printf("Successfully loaded dataframe from %s. Dimensions: %dx%d", payload.Path, df.Nrow(), df.Ncol())
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Load successful"))
}

// dataHandler returns the first 100 rows of the dataframe.
func dataHandler(w http.ResponseWriter, r *http.Request) {
	dfMutex.RLock() // lock for reading
	defer dfMutex.RUnlock()

	if df.Nrow() == 0 {
		http.Error(w, "No data loaded", http.StatusNotFound)
		return
	}

	// for simplicity, send first 100 rows. A real app would use pagination. We will add this later
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(df.Subset(series.Ints(0, 99)))
}

// summaryHandler returns the output of Gota's Describe() function.

func summaryHandler(w http.ResponseWriter, r *http.Request) {
	dfMutex.Rlock()
	defer dfMutex.RUnlock()

	if df.Nrow() == 0 {
		http.Error(w, "No data loaded", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(df.Describe())
}

func main() {
	http.HandleFunc("/load", loadHandler)
	http.HandleFunc("/data", dataHandler)
	http.HandleFunc("/summary", summaryHandler)

	log.Println("Starting data engine server on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
