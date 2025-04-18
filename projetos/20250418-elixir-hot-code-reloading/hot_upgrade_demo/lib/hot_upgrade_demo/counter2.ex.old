defmodule HotUpgradeDemo.Counter2 do
  use GenServer

  # API
  def start_link(_) do
    GenServer.start_link(__MODULE__, 0, name: __MODULE__)
  end

  def bump(by), do: GenServer.call(__MODULE__, {:bump, by})

  # Callbacks
  def init(count), do: {:ok, count}

  def handle_call({:bump, by}, _from, count) do
    {:reply, :ok, count + by}
  end

  # Callback para migração de estado em upgrades
  def code_change(_old_vsn, count, _extra) do
    {:ok, count}
  end
end
