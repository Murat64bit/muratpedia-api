build:
	GOARCH=amd64 GOOS=linux go build -o bin/main main.go

deploy_prod: build 
	serverless deploy --stage prod --aws-profile Murat64Bit

run:
	go run main.go