defmodule TimeZoneFFI do
  require Tzdata

  def utc_offset(name) do
    now =
      System.system_time(:second)
      |> :calendar.system_time_to_universal_time(:seconds)
      |> :calendar.datetime_to_gregorian_seconds()

    case Tzdata.periods_for_time(name, now, :utc) do
      [first_period | _] ->
        {:ok, Map.get(first_period, :utc_off)}

      {:error, _} ->
        {:error, :not_found}
    end
  end
end
