defmodule GitHub do
  use Tesla

  plug Tesla.Middleware.BaseUrl, "https://api.github.com"
  plug Tesla.Middleware.Headers, [{"authorization", "token xyz"}]
  plug Tesla.Middleware.JSON

  def user_repos(login) do
    get("/users/" <> login <> "/repos")

    {:ok, response} = GitHub.user_repos("teamon")

    response.status
    # => 200

    response.body
    # => [%{…}, …]

    response.headers
    # => [{"content-type", "application/json"}, ...]

  end
end