.PHONY: all build test clean deps

GOCMD=go
GOTEST=$(GOCMD) test

test:
	@ $(GOTEST) -count=1 ./...

clean:
	@ $(GOCMD) clean

