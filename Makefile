.PHONY: build clean

NOW_TIME != date -u
CSSPATH = '/static/style.css'
DOC_TEMPL = 'doc/template.html'
PANDOC_OPTS = --css $(CSSPATH) \
	--template $(DOC_TEMPL) \
	-f markdown -t html \
	-M date="$(NOW_TIME)"

all: ancient-places public/about.html public/guide.html

ancient-places : main.go
	go build
	go mod tidy

public/about.html : doc/about.txt
	pandoc $(PANDOC_OPTS) doc/about.txt > public/about.html

public/guide.html : doc/guide.txt
	pandoc --toc $(PANDOC_OPTS) doc/guide.txt > public/guide.html

clean :
	rm ancient-places
	find . -iname "*~" -exec rm '{}' ';'
