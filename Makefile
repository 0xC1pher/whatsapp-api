all:
	swag init
	swag fmt
	go fmt ./...
	go build