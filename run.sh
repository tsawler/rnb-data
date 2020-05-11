rm -f places || true
rm -f places-linux || true
go build -o places *.go
env GOOS=linux GOARCH=amd64 go build -o places-linux *.go
./places -addr ":8080" -dsn "homestead:secret@/places?parseTime=true"
