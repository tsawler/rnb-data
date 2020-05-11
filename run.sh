go build -o places *.go
./places -addr ":8080" -dsn "homestead:secret@/places?parseTime=true"
