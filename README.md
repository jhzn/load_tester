# **Load test an HTTP server with N concurrent users**

## **Start HTTP server**
```sh
#Run in shell 1
go run server.go
```

## **Start HTTP load tester**
```sh
#Run in shell 2
#First argument specifies how many concurrent users to simulate
go run main.go 10000
```
```sh
#output running the above
2019/08/18 21:19:42 Load test is done. Time taken = 1.719374361s
2019/08/18 21:19:42 Program done. Wrote stats to file: request_stats.csv
```