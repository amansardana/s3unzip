.PHONY: build clean deploy

build:
	dep ensure -v
	cd unzip && env GOOS=linux go build -ldflags="-s -w" -o ../bin/unzip . && cd ..
	cd s3unzip && env GOOS=linux go build -ldflags="-s -w" -o ../bin/s3unzip . && cd ..

clean:
	rm -rf ./bin ./vendor Gopkg.lock

deploy: clean build
	sls deploy --verbose
