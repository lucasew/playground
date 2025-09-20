#!/usr/bin/env elixir

Mix.install([
  :phoenix,
  :phoenix_html,
  :cowboy,
  :plug,
  :plug_cowboy,
  :jason
], config: [
  app: [
    {App.Endpoint, [
      
      # url: [host: "localhost", port: 42069],
      root: Path.dirname(__DIR__),
      secret_key_base: "🧂🧂🧂🧂🧂🧂🧂🧂🧂",
      http: [port: 42070],
      render_errors: [accepts: ~w(html json)],
      debug_errors: true,
      code_reloader: true,
      check_origin: false
    ]}
  ],
  phoenix: [
    serve_endpoints: true,
    persistent: true
  ],
  logger: [
    level: :debug,
    console: [
      format: "$time $metadata[$level] $message\n",
      metadata: [:request_id]
    ]
  ]
])

require Logger

defmodule App.WebserverController do
  use Phoenix.Controller, formats: [:json]

  def index(conn, _params) do
    json conn, %{teste: true}
  end
end

defmodule App do
  use Application

  def start(_type, _args) do
    import Supervisor.Spec, warn: false

    children = [
      supervisor(App.Endpoint, []),
    ]

    opts = [strategy: :one_for_one, name: App.Supervisor]
    Supervisor.start_link(children, opts)
  end
end

defmodule App.Web do
  def router do
    quote do
      use Phoenix.Router
    end
  end

  @doc """
  When used, dispatch to the appropriate controller/view/etc.
  """
  defmacro __using__(which) when is_atom(which) do
    apply(__MODULE__, which, [])
  end
end

defmodule App.Router do
  use App.Web, :router

  pipeline :browser do
    plug :accepts, ["html", "json"]
    plug :fetch_session
    plug :fetch_flash
    plug :protect_from_forgery
    plug :put_secure_browser_headers
  end

  pipeline :api do
    plug :accepts, ["html", "json"]
  end

  scope "/", App do
    pipe_through :browser
    get "/", WebserverController, :index
  end

  scope "/api", App do
    pipe_through :api
    get "/", WebserverController, :index
  end
end

defmodule App.Endpoint do
  use Phoenix.Endpoint, otp_app: :app

  plug Plug.RequestId
  plug Plug.Logger
  plug Plug.MethodOverride
  plug Plug.Head
  plug Plug.Session,
    store: :cookie,
    key: "_app_key",
    signing_salt: "🧂🧂🧂🧂🧂🧂🧂🧂🧂🧂🧂🧂"

  plug App.Router
end

defmodule App.Mixfile do
  use Mix.Project

  def project do
    [app: :app, version: "0.0.1"]
  end

  def application do
    [mod: {App, []}, applications: [:phoenix, :phoenix_html, :cowboy, :logger]]
  end
end

{:ok, pid} = App.start(:hue, :br)

ref = Process.monitor(pid)

Logger.info "no aguardo"
receive do
  {:DOWN, ^ref, _, _} -> :ok
end

