import cipher
import gleam/bytes_tree
import gleam/int
import gleam/list
import gleam/result
import gleam/string
import ipv4.{type IPv4}
import time_zone.{type TimeZone}

/// `0xA8` Game Server List. Variable length, unencrypted.
///
/// Sent to clients by a login server after successful authentication.
pub type GameServerList {
  GameServerList(servers: List(GameServer), system_info_flag: SystemInfoFlag)
}

pub type GameServer {
  GameServer(name: String, time_zone: TimeZone, ip: IPv4, port: Int)
}

// TODO: Does this control whether 0xD9 Spy On Client is sent?
pub type SystemInfoFlag {
  SendSystemInfo
  DoNotSendSystemInfo
}

pub fn encode(servers: GameServerList) -> cipher.Plaintext {
  let GameServerList(servers, system_info_flag) = servers
  let count = list.length(servers)
  let server_list =
    list.index_fold(servers, bytes_tree.new(), fn(bytes, server, i) {
      bytes
      |> bytes_tree.append(<<i:16>>)
      |> bytes_tree.append_tree(encode_game_server(server))
    })
  let length = 6 + bytes_tree.byte_size(server_list)
  let flag_byte = encode_system_info_flag(system_info_flag)

  bytes_tree.from_bit_array(<<0xA8, length:16, flag_byte:8, count:16>>)
  |> bytes_tree.append_tree(server_list)
  |> bytes_tree.to_bit_array
  |> cipher.Plaintext
}

fn encode_time_zone(time_zone: TimeZone) -> Int {
  // Client expects offset from UTC, in hours, as a single byte.
  let seconds = time_zone.offset_in_seconds(time_zone) |> result.unwrap(0)
  seconds / 60 / 60
}

fn encode_system_info_flag(flag: SystemInfoFlag) -> Int {
  // https://docs.polserver.com/packets/index.php?Packet=0xA8
  case flag {
    DoNotSendSystemInfo -> 0xCC
    SendSystemInfo -> 0x64
  }
}

fn encode_game_server(game_server: GameServer) -> bytes_tree.BytesTree {
  // Server name must be exactly 32 bytes (padded with null bytes).
  let name =
    game_server.name
    |> string.pad_end(32, "\u{0000}")
    |> string.slice(0, 32)
  let ip_bytes =
    int.bitwise_shift_left(game_server.ip.0, 24)
    |> int.bitwise_or(int.bitwise_shift_left(game_server.ip.1, 16))
    |> int.bitwise_or(int.bitwise_shift_left(game_server.ip.2, 8))
    |> int.bitwise_or(game_server.ip.3)

  bytes_tree.new()
  |> bytes_tree.append_string(name)
  // TODO: percent full
  |> bytes_tree.append(<<0:8>>)
  |> bytes_tree.append(<<encode_time_zone(game_server.time_zone)>>)
  |> bytes_tree.append(<<ip_bytes:32-little>>)
}
