# TORA

Desktop web view of the page in `tora.zip`. The process opens that zip and serves it with lewkit's SPA middleware. Files stay in the archive.

From this directory:

```sh
mise exec -- go test ./...
mise run desktop
```

`mise run desktop` stamps `br.tec.lew.tora` and opens the page in the system web view. The window is 1060×780. The input shell is 980×700 and opens at (36, 28).
