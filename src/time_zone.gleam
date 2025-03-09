import gleam/erlang/atom
import gleam/result

pub type TimeZone {
  AmericaDetroit
  AmericaPhoenix
  AmericaLosAngeles

  UTC
}

/// The IANA tzdata canonical name for a time zone.
pub fn name(time_zone: TimeZone) -> String {
  case time_zone {
    AmericaDetroit -> "America/Detroit"
    AmericaPhoenix -> "America/Phoenix"
    AmericaLosAngeles -> "America/Los_Angeles"
    UTC -> "Etc/UTC"
  }
}

@external(erlang, "Elixir.TimeZoneFFI", "utc_offset")
fn utc_offset_in_seconds(name: String) -> Result(Int, atom.Atom)

pub fn offset_in_seconds(time_zone: TimeZone) -> Result(Int, Nil) {
  name(time_zone)
  |> utc_offset_in_seconds()
  |> result.replace_error(Nil)
}
