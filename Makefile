BINARY_NAME=firewalld-tui

all: build

run:
	go run main.go

build:
	go build -o bin/$(BINARY_NAME) main.go

install:
	install -m 755 bin/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)

.PHONY: all run build install clean
