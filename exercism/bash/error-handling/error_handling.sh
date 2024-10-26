#!/usr/bin/env bash

main () {
  case "$#" in
    1) echo "Hello, $1";;
    *)
      echo "Usage: error_handling.sh <person>"
      exit 1
    ;;
  esac
}
main "$@"
