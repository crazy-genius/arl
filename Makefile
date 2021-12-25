vet:
	go vet ./...
list:
	go list ./...
lint:
	golint ./...
race_mac:
	GODEBUG=netdns=go go run --race cmd/arl/main.go
race:
	go run --race cmd/arl/main.go