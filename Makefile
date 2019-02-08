SHELL := /bin/bash
FILE  := tmp.txt
CAT   := bat

ifeq (, $(shell which bat))
	$(error "No bat in $(PATH), consider installing ti")
	CAT = cat
endif

build_and_run: clean build run

.PHONY : build_and_run build run clean testfile

build :
	@echo "-> Building"
	@cd utils/; go build .; cd ..
	@go build .
	@echo "->   Done"

run :
	@echo "-> Running"
	@./led $(FILE)

clean :
	@echo "-> Cleaning up"
	@-rm led

testfile :
	@echo -e "-> Generating test file"
	@echo -e "This is a line.\nThis is another line.\n\n\nThis is the end." > tmp.txt
	@$(CAT) tmp.txt
