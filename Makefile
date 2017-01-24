build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 govendor build -a -installsuffix cgo +local

test:
	govendor test -v +local

clean:
	rm rancher-service-updater
