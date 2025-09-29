#!/usr/bin/env bash

main () {
  ret=`{
    (($1 % 3 == 0)) && echo -n Pling
    (($1 % 5 == 0 )) && echo -n Plang
    (($1 % 7 == 0 )) && echo -n Plong
  }`
  if [ -z "$ret" ]; then
    ret="$1"
  fi
  echo "$ret"
}

main "$@"
