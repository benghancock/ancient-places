#! /bin/bash

# Build HTML documentation from text source
# This script requires `pandoc` <https://pandoc.org/>

CSSPATH='/static/style.css'

echo "* Converting ancient-places documentation to HTML ..."

pandoc --verbose -s --css=$CSSPATH -f markdown -t html \
       -A doc/html_components/footer.html \
       doc/about.txt > public/about.html

echo "* Done"
