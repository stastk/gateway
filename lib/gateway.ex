defmodule Gateway do
  @moduledoc """
  Documentation for `Gateway`.
  """

  @doc """
  Hello world.

  ## Examples

      iex> Gateway.hello()
      :world

  """

  use Tesla

  plug Tesla.Middleware.BaseUrl, "http://ne3a.ru/remapper"
  plug Tesla.Middleware.JSON

  def req do
    post("/v2?t=NDksOTcsMTAwLDEyMiw5Nw==&d=gibberish", "")
  end

end