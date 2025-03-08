import cipher.{type Cipher}
import error
import gleam/bit_array
import gleam/bytes_tree
import gleam/erlang
import gleam/erlang/process
import gleam/int
import gleam/io
import gleam/result
import glisten
import glisten/socket
import glisten/tcp
import utils as u

pub const max_packet_size = 0xF000

// /// Encode a packet to a byte tree.
// fn encode_packet(packet: Packet) -> BytesTree {
//   let bytes = case packet {
//     ConnectToGameServer(game_server) -> {
//       bytes_tree.new()
//       |> bytes_tree.append(<<
//         0x8C,
//         reversed_ip(game_server.ip):32,
//         game_server.port:16,
//         0:32,
//       >>)
//     }
//   }

//   io.println(bit_array.inspect(bytes_tree.to_bit_array(bytes)))

//   bytes
// }

// fn handle_select_server(client: Client) -> Result(Client, ClientError) {
//   io.println(inspect_socket(client) <> ": waiting on server selection")
//   // TODO: this will hang if client has sent 0xD9 Client Info
//   let assert Ok(#(new_client, plaintext)) = read(client, 3)

//   // plaintext could be either 0xD9 Client Info or 0xA0 Select Server
//   case plaintext.bits {
//     <<0xA0, shard_index:16>> -> {
//       let game_server =
//         yielder.from_list(game_servers)
//         |> yielder.at(shard_index)
//         |> result.lazy_unwrap(fn() {
//           panic as "handle_select_server: index out of range"
//         })
//       io.debug(#(shard_index, game_server))
//       do_work(Client(..new_client, state: NeedsRelay(game_server)))
//     }
//     <<0xD9, _:bits>> -> todo
//     _ -> todo
//   }
// }

// fn relay_to_game_server(client: Client, server: GameServer) -> Result(Client, ClientError) {
//   io.println(inspect_socket(client) <> ": relaying to " <> server.name)

//   let packet = ConnectToGameServer(server)
//   let data = PlainText(encode_packet(packet) |> bytes_tree.to_bit_array)
//   let assert Ok(new_client) = write(client, data)

//   Ok(Client(..new_client, state: Relayed))
// }
type Connection =
  glisten.Connection(BitArray)

pub type Client {
  Client(conn: Connection, inbox: BitArray, outbox: BitArray, cipher: Cipher)
}

pub fn new(conn: Connection) {
  Client(conn, <<>>, <<>>, cipher.nil())
}

// pub opaque type Message {
//   Start
//   Stop

//   PushToInbox(data: BitArray)

//   Authenticate(reply_with: Subject(Result(Nil, Error)))
// }

// fn handle_message(
//   message: Message,
//   client: Client,
// ) -> actor.Next(Message, Client) {
//   case message {
//     Start -> {
//       // TODO
//       actor.continue(client)
//     }

//     Authenticate(reply_with) -> {
//       let result = {
//         use client <- result.try(handle_login_seed(client))
//         use client <- result.try(handle_login_request(client))
//         use client <- result.try(send_game_server_list(client, game_servers))
//         Ok(client)
//       }

//       case result {
//         Ok(client) -> {
//           actor.send(reply_with, Ok(Nil))
//           actor.continue(client)
//         }

//         Error(error) -> {
//           actor.send(reply_with, Error(error))
//           actor.Stop(process.Abnormal(string.inspect(error)))
//         }
//       }
//     }

//     Stop -> actor.Stop(process.Normal)

//     PushToInbox(data) -> {
//       let inbox = bit_array.append(client.inbox, data)
//       let n = int.to_string(bit_array.byte_size(data))
//       io.println(inspect(client) <> ": pushing " <> n <> " bytes to inbox")
//       io.debug(inbox)
//       actor.continue(Client(..client, inbox:))
//     }
//   }
// }

fn socket_string(client: Client) -> String {
  glisten.get_client_info(client.conn)
  |> result.map(fn(info) {
    let ip = glisten.ip_address_to_string(info.ip_address)
    let port = int.to_string(info.port)

    ip <> ":" <> port
  })
  |> result.unwrap("?")
}

pub fn inspect(client: Client) -> String {
  socket_string(client) <> " (" <> erlang.format(process.self()) <> ")"
}

fn socket_read(client: Client) -> Result(Client, socket.SocketReason) {
  use data <- result.try(tcp.receive_timeout(client.conn.socket, 0, 5000))
  let inbox = <<client.inbox:bits, data:bits>>

  Ok(Client(..client, inbox:))
}

pub fn read(
  client: Client,
  size: Int,
) -> Result(#(Client, cipher.CipherText), error.Error) {
  let size = case size {
    n if n < 0 -> 0
    n if n > max_packet_size -> max_packet_size
    _ -> size
  }
  let inbox_size = bit_array.byte_size(client.inbox)
  let wanted = int.to_string(size)
  let have = int.to_string(inbox_size)

  io.println(inspect(client) <> ": read: want " <> wanted <> ", have " <> have)

  case size {
    0 -> Ok(#(client, cipher.CipherText(<<>>)))

    n if n <= inbox_size -> {
      io.println(inspect(client) <> ": read: pulling from inbox")
      let assert <<bits:bytes-size(n), rest:bytes>> = client.inbox
      let new_client = Client(..client, inbox: rest)
      Ok(#(new_client, cipher.CipherText(bits)))
    }

    n if n > inbox_size -> {
      io.println(inspect(client) <> ": read: reading from socket")
      case socket_read(client) {
        Ok(new_client) -> read(new_client, size)
        Error(reason) -> Error(error.ReadError(reason))
      }
    }

    // TODO: is this reachable?
    _ ->
      panic as {
        inspect(client) <> ": read: inbox_size panic (" <> have <> ")"
      }
  }
}

pub fn write(
  client: Client,
  cipher_text: cipher.CipherText,
) -> Result(Client, error.Error) {
  let bits = bytes_tree.from_bit_array(cipher_text.bits)
  use _ <- u.try_map(tcp.send(client.conn.socket, bits), error.WriteError)
  Ok(client)
}
