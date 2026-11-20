Cara Run & Build
```go
// build the plugins first

go run . coordinator /var/tmp/mr-sock pg-*.txt
go run . worker ./wc.so /var/tmp/mr-sock
go run . sequential ./wc.so pg-*.txt
```
