#!/usr/bin/env elixir
Mix.install([ 
  :req, 
  {:jason, "~> 1.0"} 
])

Req.get!("https://api.github.com/repos/elixir-lang/elixir").body["description"]
|> dbg()


