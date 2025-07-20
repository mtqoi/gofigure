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
	dfMutex = &sync.RWMutex{}
)

// loadHandler loads a CSV file into the global dataframe.
// expects a JSON body like: {"path": "/path/to/my/file.csv"}
// TODO: ensure we can use relative filepaths to where the CLI user is, not just to this main.go file
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
	_, err = w.Write([]byte("Load successful"))
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

// NOTE:  this is the function that I want to replace with duckdb internals for pagination and limiting
func dataHandler(w http.ResponseWriter, r *http.Request) {
	dfMutex.RLock()
	defer dfMutex.RUnlock()

	if df.Nrow() == 0 {
		http.Error(w, "No data loaded", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// remember to handle cases with fewer than 100 rows
	rowCount := df.Nrow()
	if rowCount > 100 {
		rowCount = 100
	}
	indices := make([]int, rowCount)
	for i := 0; i < rowCount; i++ {
		indices[i] = i
	}

	subset := df.Subset(series.Ints(indices))
	records := subset.Records()

	if err := json.NewEncoder(w).Encode(records); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		log.Printf("Error encoding dataframe: %v", err)
	}

	log.Println("Successfully returned data")
}

// TODO: replace with output of duckdb describe() on the df
// have a think about where best to do column-level descriptive statistics
func summaryHandler(w http.ResponseWriter, r *http.Request) {
	dfMutex.RLock()
	defer dfMutex.RUnlock()

	if df.Nrow() == 0 {
		http.Error(w, "No data loaded", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	summary := df.Describe()
	records := summary.Records()

	if err := json.NewEncoder(w).Encode(records); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		log.Printf("Error encoding summary: %v", err)
		return
	}

	log.Println("Successfully returned summary")

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
