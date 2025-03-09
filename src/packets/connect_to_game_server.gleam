import cipher
import gleam/bytes_tree
import packets/game_server_list

/// `0x8C` Connect to Game Server. Length 11, unencrypted.
///
/// Sent by a login server to a client upon receiving the `0xA0` Select Game
/// Server packet, after the client has selected a game server (shard) on which
/// to play.
pub type ConnectToGameServer {
  ConnectToGameServer(game_server: game_server_list.GameServer, new_key: Int)
}

pub fn encode(packet: ConnectToGameServer) -> cipher.Plaintext {
  let ConnectToGameServer(server, new_key) = packet
  let ip_bytes = <<server.ip.0:8, server.ip.1:8, server.ip.2:8, server.ip.3:8>>

  bytes_tree.from_bit_array(<<0x8C>>)
  |> bytes_tree.append(ip_bytes)
  |> bytes_tree.append(<<server.port:16, new_key:32>>)
  |> bytes_tree.to_bit_array()
  |> cipher.Plaintext()
}
