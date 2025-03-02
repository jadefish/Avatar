import gleeunit
import gleeunit/should
import time_zone

pub fn main() {
  gleeunit.main()
}

pub fn offset_in_seconds_test() {
  // Using America/Phoenix here so this test doesn't fail half the year.
  let assert Ok(offset) = time_zone.offset_in_seconds(time_zone.AmericaPhoenix)
  offset |> should.equal(-25_200)

  let assert Ok(offset) = time_zone.offset_in_seconds(time_zone.UTC)
  offset |> should.equal(0)
}
