# Show do Milhão 3

Desktop web view of the game in `milhao3.zip`. The process opens that zip and serves it with lewkit's SPA middleware. Files stay in the archive.

From this directory:

```sh
mise exec -- go test ./...
mise run desktop
```

`mise run run` stamps `br.tec.lew.milhao3` and opens the page in the system web view. The window is 960×720. The game board is 640×480 and scales to fit.
