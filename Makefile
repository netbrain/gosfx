.PHONY: all

all: clean install linux86 linux64 windows86 windows64

clean: 
	rm -r ./bin/*

build:
	go build -v \
	github.com/netbrain/gosfx/cmd/gosfx-extractor \
	github.com/netbrain/gosfx/cmd/gosfx-packer \

linux86: 
	GOOS=linux \
	GOARCH=386 \
	go build -v -ldflags "-s -w" -o ./bin/linux_386/gosfx-extractor github.com/netbrain/gosfx/cmd/gosfx-extractor && \
	go build -v -ldflags "-s -w" -o ./bin/linux_386/gosfx-packer github.com/netbrain/gosfx/cmd/gosfx-packer && \
	upx -q --brute ./bin/linux_386/gosfx-extractor

linux64: 
	GOOS=linux \
	GOARCH=amd64 \
	go build -v -ldflags "-s -w" -o ./bin/linux_amd64/gosfx-extractor github.com/netbrain/gosfx/cmd/gosfx-extractor && \
	go build -v -ldflags "-s -w" -o ./bin/linux_amd64/gosfx-packer github.com/netbrain/gosfx/cmd/gosfx-packer && \
	upx -q --brute ./bin/linux_amd64/gosfx-extractor

windows86:
	GOOS=windows \
	GOARCH=386 \
	go build -v -ldflags "-s -w" -o ./bin/windows_386/gosfx-extractor.exe github.com/netbrain/gosfx/cmd/gosfx-extractor && \
	go build -v -ldflags "-s -w" -o ./bin/windows_386/gosfx-packer.exe github.com/netbrain/gosfx/cmd/gosfx-packer && \
	upx -q --brute ./bin/windows_386/gosfx-extractor.exe

windows64:
	GOOS=windows \
	GOARCH=amd64 \
	go build -v -ldflags "-s -w" -o ./bin/windows_amd64/gosfx-extractor.exe github.com/netbrain/gosfx/cmd/gosfx-extractor && \
	go build -v -ldflags "-s -w" -o ./bin/windows_amd64/gosfx-packer.exe github.com/netbrain/gosfx/cmd/gosfx-packer && \
	upx -q --brute ./bin/windows_amd64/gosfx-extractor.exe

install_fast:
	go install -v -ldflags "-s -w" github.com/netbrain/gosfx/cmd/gosfx-extractor && \
	go install -v -ldflags "-s -w" github.com/netbrain/gosfx/cmd/gosfx-packer

install: install_fast
	upx -q --brute ${GOPATH}/bin/gosfx-extractor || upx -q --brute ${GOPATH}/bin/gosfx-extractor.exe