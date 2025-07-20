package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
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

// dataHandler serves paginated data from the dataframe.
// Query parameters:
//   - start: The starting row index (default: 0)
//   - limit: Maximum number of rows to return (default: 100, max: 1000)
//
// Returns JSON-encoded records from the specified range.
func dataHandler(w http.ResponseWriter, r *http.Request) {
	dfMutex.RLock()
	defer dfMutex.RUnlock()

	if df.Nrow() == 0 {
		http.Error(w, "No data loaded", http.StatusNotFound)
		return
	}

	// Parse and validate parameters

	// default 0
	start, err := strconv.Atoi(r.URL.Query().Get("start"))
	if err != nil || start < 0 {
		start = 0
	}

	// default 100
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit <= 0 {
		limit = 100
	}

	// limit max
	if limit > 1000 {
		limit = 1000
	}

	end := start + limit
	if end > df.Nrow() {
		end = df.Nrow()
	}

	// Create subset
	if start >= end {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"records": []interface{}{},
			"message": "No data in specified range",
			"total":   df.Nrow(),
		})
		return
	}

	indices := make([]int, end-start)
	for i := 0; i < len(indices); i++ {
		indices[i] = start + i
	}

	subset := df.Subset(series.Ints(indices))
	records := subset.Records()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"records": records,
		"total":   df.Nrow(),
		"start":   start,
		"limit":   limit,
	}); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		log.Printf("Error encoding dataframe: %v", err)
		return
	}

	log.Printf("Successfully returned %d rows of data", len(records))
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
