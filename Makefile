all:
	mkdir -p build
	#go get -u github.com/dongleiw/dssh
	go get github.com/dongleiw/dssh
	go build -o build/dssh
