Load data: 
```aiignore
curl -X POST -d '{"path":"../../test_data/iris.csv"}' http://localhost:8080/load
```

Get data:
```aiignore
curl http://localhost:8080/data
```
& has parameters: `start`, `limit`

Describe data:
```aiignore
curl http://localhost:8080/summary
```