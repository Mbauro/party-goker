BINARY_NAME=party-goker

build:
	go build -o $(BINARY_NAME)-srv ./server
	go build -o $(BINARY_NAME)-client ./client

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)-*

build-all:
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-srv-darwin-amd64 ./server
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY_NAME)-srv-darwin-arm64 ./server
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-srv-linux ./server
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME)-srv-windows.exe ./server

	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-client-darwin-amd64 ./client
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY_NAME)-client-darwin-arm64 ./client
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-client-linux ./client
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME)-client-windows.exe ./client
