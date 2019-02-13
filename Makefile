SHELL := /bin/bash
FILE  := tmp.txt
CAT   := bat

ifeq (, $(shell which bat))
	$(error "No bat in $(PATH), consider installing ti")
	CAT = cat
endif

build_and_run: clean build run

.PHONY : build_and_run build run clean testfile install

build :
	@echo "-> Building"
	@cd utils/; go build .; cd ..
	@go build .
	@echo "->   Done"

build_silent :
	@cd utils/; go build .; cd ..
	@go build .

run :
	@echo "-> Running"
	@./led $(FILE)

clean :
	@echo "-> Cleaning up"
	@-rm led

install : build_silent
	@echo "WOW, you're either brave or very stupid..."
	@echo "-> Installing led in /usr/local/bin/led"
	@ln -sF $(shell pwd)/led /usr/local/bin/led
	@echo "->   Done"

testfile :
	@echo -e "-> Generating test file"
	@echo -e "This is a line.\nThis is another line.\n\n\nThis is the end." > tmp.txt
	@$(CAT) tmp.txt
