import cipher.{type Cipher}
import error
import gleam/bit_array
import gleam/bytes_tree
import gleam/int
import gleam/io
import gleam/option.{type Option, None}
import gleam/result
import glisten
import glisten/socket
import glisten/tcp
import utils as u

pub const max_packet_size = 0xF000

type Connection =
  glisten.Connection(BitArray)

pub type Client {
  Client(
    conn: Connection,
    login_seed: Option(cipher.Seed),
    inbox: BitArray,
    outbox: BitArray,
    cipher: Cipher,
  )
}

pub fn new(conn: Connection) {
  Client(conn, None, <<>>, <<>>, cipher.nil())
}

pub fn inspect(client: Client) -> String {
  u.connection_addr(glisten.get_client_info(client.conn))
}

fn socket_read(client: Client) -> Result(Client, socket.SocketReason) {
  use data <- result.try(tcp.receive(client.conn.socket, 0))
  let inbox = <<client.inbox:bits, data:bits>>

  Ok(Client(..client, inbox:))
}

pub fn read(
  client: Client,
  size: Int,
) -> Result(#(Client, cipher.Ciphertext), error.Error) {
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
    0 -> Ok(#(client, cipher.Ciphertext(<<>>)))

    n if n <= inbox_size -> {
      io.println(inspect(client) <> ": read: pulling from inbox")
      let assert <<bits:bytes-size(n), rest:bytes>> = client.inbox
      let new_client = Client(..client, inbox: rest)
      Ok(#(new_client, cipher.Ciphertext(bits)))
    }

    n if n > inbox_size -> {
      io.println(inspect(client) <> ": read: reading from socket")
      case socket_read(client) {
        Ok(new_client) -> read(new_client, size)
        Error(reason) -> Error(error.ReadError(reason))
      }
    }

    _ -> panic as "client.read: unreachable case"
  }
}

pub fn write(
  client: Client,
  ciphertext: cipher.Ciphertext,
) -> Result(Client, error.Error) {
  let bytes = bytes_tree.from_bit_array(ciphertext.bits)
  use _ <- u.try_map(tcp.send(client.conn.socket, bytes), error.WriteError)
  Ok(client)
}
