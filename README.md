# GoFigure

A data exploration and visualization tool built in Go with both server and terminal UI components.

This is a work-in-progress tool primarily designed for my own use, and for practice with Go.

## Overview

GoFigure provides a simple way to load, visualize, and explore datasets. It consists of:
- A server component for loading and querying data
- A TUI (Terminal User Interface) for interactive data exploration

![TUI Data Table View](img.png)

![TUI Summary View](img_1.png)

### Server Component

Start the server:

Run `go run main.go` from the `/cmd` dir

#### API Endpoints

**Load Data:**

From the `/cmd` dir 

```
curl -X POST -d '{"path":"../../test_data/iris.csv"}' http://localhost:8080/load
```
**Get Data:**

```
curl http://localhost:8080/data
```
Parameters:
- `start`: Starting row index
- `limit`: Number of rows to return

**Get Data Summary:**


```
curl http://localhost:8080/summary
```

### Terminal UI

Run the terminal UI:

from the `/tui` dir run `go run main.go`


## TODO

### tui 
- [ ] refactor into different files: i) requests; ii) modelling; iii) displaying
- [ ] add load data to server capabilities
- [ ] read from stdin + display in table
- [ ] variable column widths depending on data
- [ ] visual sugar: scrollbars, indexes
- [ ] interactivity: paging, scrolling left/right/up/down, filtering, selecting endpoint, sorting
- [ ] file browser?
- [ ] status bar
- [ ] other keyboard shortcuts (home/end, pgup/pgdown)
- [ ] row numbering
- [ ] help popup
- [ ] summary statistics like jetbrains/positron?
- [ ] change datatypes? round numerics? formatting e.g. %'s, $'s?
- [ ] add mouse functionality
- [ ] [charting?](https://github.com/NimbleMarkets/ntcharts)

### server 
- [ ] read from stdin 
- [ ] add extra formats: json, parquet, duckdb, bigquery
- [ ] replace internals with duckdb
- [ ] add filtering and paging
- [ ] add sorting
- [ ] caching or compression?
- [ ] different server ports

### both
- [ ] config file for server port
- [ ] add tests
- [ ] shared package for data modelling?