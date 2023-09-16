.PHONY: build clean

NOW_TIME != date -u
CSSPATH = '/static/style.css'
DOC_TEMPL = 'doc/html_components/pandoc_template.html'

all: ancient-places public/about.html public/guide.html

ancient-places : main.go
	go build
	go mod tidy

public/about.html : doc/about.txt
	pandoc \
		--css $(CSSPATH) \
		--template $(DOC_TEMPL) \
		-f markdown -t html \
		-M date="$(NOW_TIME)" \
		doc/about.txt > public/about.html

public/guide.html : doc/guide.txt
	pandoc \
		--toc \
		--css $(CSSPATH) \
		--template $(DOC_TEMPL) \
		-f markdown -t html \
		-M date="$(NOW_TIME)" \
		doc/guide.txt > public/guide.html

clean :
	rm ancient-places
	find . -iname "*~" -exec rm '{}' ';'
