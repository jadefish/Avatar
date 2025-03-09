import gleam/list
import gleam/result
import gleam/string
import gleeunit
import gleeunit/should
import packets/game_server_list
import time_zone

pub fn main() {
  gleeunit.main()
}

pub fn encode_test() {
  let server =
    game_server_list.GameServer(
      "name",
      time_zone.AmericaPhoenix,
      #(192, 168, 10, 20),
      7776,
    )
  let packet =
    game_server_list.GameServerList(
      [server],
      game_server_list.DoNotSendSystemInfo,
    )
  let padded_name = server.name <> string.repeat("\u{0000}", times: 28)
  let time_zone_offset =
    { time_zone.offset_in_seconds(server.time_zone) |> result.unwrap(0) }
    / 60
    / 60
  let expected_length = 46

  game_server_list.encode(packet).bits
  |> should.equal(<<
    0xA8,
    expected_length:16,
    0xCC,
    list.length(packet.servers):16,
    0:16,
    padded_name:utf8,
    0:8,
    time_zone_offset:8,
    server.ip.3:8,
    server.ip.2:8,
    server.ip.1:8,
    server.ip.0:8,
  >>)
}
