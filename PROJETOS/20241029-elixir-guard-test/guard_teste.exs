#!/usr/bin/env elixir

defmodule Teste do
  def cmp(x, y) when x > y do
    x > y
  end
end


IO.puts (Teste.cmp 10, 9)

IO.puts (Teste.cmp 9, 10)
