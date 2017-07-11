export GOPATH="$PWD"
go fmt ciel
go vet ciel
go fmt ciel-driver
go vet ciel-driver
go build ciel
