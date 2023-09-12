.PHONY: build docs clean

ancient-places : main.go
	go build
	go mod tidy

docs :
	./build_docs.sh

clean :
	rm ancient-places
	find . -iname "*~" -exec rm '{}' ';'
