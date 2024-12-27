import gleam/bytes_tree.{type BytesTree}
import gleam/int
import gleam/list
import gleam/result
import gleam/string
import ipv4.{type IPv4}
import packets
import time_zone.{type TimeZone}

pub type GameServerList {
  GameServerList(servers: List(GameServer))
}

// TODO: https://docs.polserver.com/packets/index.php?Packet=0xA8
const system_info_flag = 0xCC

pub fn encode(
  game_server_list: GameServerList,
) -> Result(BytesTree, packets.Error) {
  let GameServerList(servers) = game_server_list
  let count = list.length(servers)
  let server_list =
    list.index_fold(servers, bytes_tree.new(), fn(bytes, server, i) {
      bytes
      |> bytes_tree.append(<<i:16>>)
      |> bytes_tree.append_tree(encode_game_server(server))
    })
  let length = 6 + bytes_tree.byte_size(server_list)

  bytes_tree.from_bit_array(<<0xA8, length:16, system_info_flag:8, count:16>>)
  |> bytes_tree.append_tree(server_list)
  |> Ok
}

fn encode_time_zone(time_zone: TimeZone) -> Int {
  // Client expects offset from UTC, in hours, as a single byte.
  let seconds = time_zone.offset_in_seconds(time_zone) |> result.unwrap(0)
  seconds / 60 / 60
}

pub type GameServer {
  GameServer(name: String, time_zone: TimeZone, ip: IPv4, port: Int)
}

fn reversed_ipv4(ip: IPv4) -> Int {
  // (192 << 24) | (168 << 16) | (68 << 8) | 58
  // but, reversed... for some reason.
  int.bitwise_shift_left(ip.3, 24)
  |> int.bitwise_or(int.bitwise_shift_left(ip.2, 16))
  |> int.bitwise_or(int.bitwise_shift_left(ip.1, 8))
  |> int.bitwise_or(ip.0)
}

fn encode_game_server(game_server: GameServer) -> bytes_tree.BytesTree {
  // Server name must be exactly 32 bytes (padded with null bytes).
  let name =
    game_server.name
    |> string.pad_end(32, "\u{0000}")
    |> string.slice(0, 32)

  bytes_tree.new()
  |> bytes_tree.append_string(name)
  // TODO: percent full
  |> bytes_tree.append(<<0>>)
  |> bytes_tree.append(<<encode_time_zone(game_server.time_zone)>>)
  |> bytes_tree.append(<<reversed_ipv4(game_server.ip):32>>)
}
