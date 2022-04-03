all:
	mkdir -p build
	#go build -o build/dssh
	#GOARCH=amd64 GOOS=linux go build -o build/dssh
	GOARCH=amd64 GOOS=windows go build -o build/dssh
