.PHONY: build docs clean

ancient-places : main.go
	go build

docs :
	./build_docs.sh

clean :
	rm ancient-places
	find . -iname "*~" -exec rm '{}' ';'
