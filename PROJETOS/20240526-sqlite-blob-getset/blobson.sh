#!/usr/bin/env bash

function sql {
  sqlite3 database.db "$@"
}

sql '.read setup.sql'

if [ $# -eq 0 ]; then
  sql 'select rowid, description from blobs'
  exit
fi

_tmpdir=$(mktemp -d)
echo $_tmpdir >&2

# trap "rm -rf $_tmpdir" EXIT

itemID="$1"; shift

_fifo=$_tmpdir/fifo
# mkfifo $_fifo

if [ "$itemID" == "-" ]; then
  cat /dev/stdin >> $_fifo
  sql "insert into blobs(description, data) values ('$*', readfile('$_fifo'))"
else
  sql "select writefile('$_fifo', data) from blobs where rowid = $itemID" > /dev/null
  cat $_fifo
fi
