#!/usr/bin/env bash


main () {
  if [ $# != 2 ]; then
    echo "Usage: hamming.sh <string1> <string2>"
    exit 1
  fi
  a=($(echo "$1" | sed 's;\([^ ]\);\1 ;g')); shift
  b=($(echo "$1" | sed 's;\([^ ]\);\1 ;g')); shift
  if [ "${#a[@]}" != "${#b[@]}" ]; then
    echo "strands must be of equal length"
    exit 1
  fi
  diffs=0

  i=0
  for a_item in "${a[@]}"; do
    b_item="${b[$i]}"
    if [ "$a_item" != "$b_item" ]; then
      diffs=$((diffs+1))
    fi 
    i=$((i+1))
  done
  echo $diffs
}

main "$@"
