PREFIX  ?= /usr/local

qcal: 
	GOOS=linux GOARCH=amd64 go build -o qcal -ldflags="-s -w"

linux-arm:
	GOOS=linux GOARCH=arm go build -o qcal -ldflags="-s -w"

macos:	
	GOOS=darwin GOARCH=amd64 go build -o app-amd64-darwin -ldflags="-s -w"

windows:
	GOOS=windows GOARCH=amd64 go build -o qcal.exe -ldflags="-s -w"

clean:
	rm -f qcal

install: qcal
	install -d $(PREFIX)/bin/
	install -m 755 qcal $(PREFIX)/bin/qcal

uninstall:
	rm -f $(PREFIX)/bin/qcal
