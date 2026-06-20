#!/usr/bin/env bash
OUTFILE=/tmp/programa
function rodarc () {
    gcc $1 -o $OUTFILE -Wall && $OUTFILE
}

function buildc () {
    gcc $* -o $OUTFILE -Wall && echo "Seu programa está situado em $OUTFILE."
}

function buildgo () {
    go build -o $OUTFILE $* && echo "Seu programa está situado em $OUTFILE"
}
