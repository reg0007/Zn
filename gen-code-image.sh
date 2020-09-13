#!/bin/sh

SNIPPETS_DIR="./doc/snippets"
EXPORT_DIR="./doc/images"
ZNT="./znt"

# find and exec
for x in $(find $SNIPPETS_DIR -name "*.zn")
do
    FILTER_FILE=$(echo $x | sed s/\\//-/g | sed s/.-doc-snippets-//g | sed s/.zn//g)
    EXPORT_FILE="$EXPORT_DIR/$FILTER_FILE.png"

    echo "going to export $x -> $EXPORT_FILE..."
    $ZNT gen-code-image -o $EXPORT_FILE $x
done