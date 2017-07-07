export GOPATH="$PWD"
go fmt ciel-cli
go vet ciel-cli
go fmt ciel-driver
go vet ciel-driver
go build ciel-cli
