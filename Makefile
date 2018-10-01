
.PHONY: build

build:
	@go build -o fony

dev: build
	SUITE_URL=https://raw.githubusercontent.com/schigh/fony/master/examples/sample.json ./fony -p 8888
