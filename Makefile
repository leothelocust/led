SHELL := /bin/bash

init:
	@echo -n "Building Imports..."; \
	cd ~/go/src/editor/utils/; go build .; cd ..; \
	echo -e "\nDone"; \
	echo "Running Editor"; \
	go run main.go file.txt;
