all:
	GOOS=darwin go build -o bender-darwin main.go
	GOOS=linux go build -o bender-linux main.go
