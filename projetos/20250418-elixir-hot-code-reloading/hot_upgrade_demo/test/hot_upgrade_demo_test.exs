defmodule HotUpgradeDemoTest do
  use ExUnit.Case
  doctest HotUpgradeDemo

  test "greets the world" do
    assert HotUpgradeDemo.hello() == :world
  end
end
