#! /bin/bash

# Build HTML documentation from text source
# This script requires `pandoc` <https://pandoc.org/>

CSSPATH='/static/style.css'

echo "* Converting ancient-places documentation to HTML ..."

for doc in doc/about.txt doc/guide.txt; do
    bn=$(basename "$doc" .txt)
    pandoc --verbose -s --css=$CSSPATH -f markdown -t html \
	   -A doc/html_components/footer.html \
	   $doc > public/${bn}.html
done

echo "* Done"
