#!/usr/bin/env bash

main () {
  ret=""
  if [ $(($1 % 3)) == 0 ]; then
    ret="${ret}Pling"
  fi
  if [ $(($1 % 5)) == 0 ]; then
    ret="${ret}Plang"
  fi
  if [ $(($1 % 7)) == 0 ]; then
    ret="${ret}Plong"
  fi
  if [ "$ret" == "" ]; then
    ret="${ret}$1"
  fi
  echo "$ret"
}
main "$@"
