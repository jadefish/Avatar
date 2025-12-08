import aio.{type Closer, type Reader, type Writer}
import cipher.{type Cipher, type Ciphertext, type Plaintext}
import error
import gleam/bit_array
import gleam/int
import gleam/result
import log

pub const max_packet_size = 0xF000

pub type Client {
  Client(
    id: String,
    // For now, a TCP client socket "IP:port" string
    reader: Reader,
    writer: Writer,
    closer: Closer,
    buffer: BitArray,
    cipher: Cipher,
  )
}

pub fn new(id: String, reader: Reader, writer: Writer, closer: Closer) {
  Client(id, reader, writer, closer, <<>>, cipher.nil())
}

pub fn with_cipher(client: Client, cipher cipher: cipher.Cipher) -> Client {
  Client(..client, cipher: cipher)
}

pub fn close(client: Client) -> Result(Client, error.Error) {
  let _ = client.closer()
  Ok(client)
}

pub fn inspect(client: Client) -> String {
  client.id
}

fn underlying_read(client: Client, size) -> Result(Client, error.Error) {
  // `size` assumed to be in the range 0 < size < |client.buffer|.

  use data <- result.try(client.reader(size))
  let buffer = <<client.buffer:bits, data:bits>>

  log.debug(
    inspect(client)
    <> ": read "
    <> int.to_string(bit_array.byte_size(data))
    <> " from socket",
  )

  Ok(Client(..client, buffer:))
}

pub fn read(
  client: Client,
  size: Int,
) -> Result(#(Client, Ciphertext), error.Error) {
  let size = int.clamp(size, min: 0, max: max_packet_size)
  let buffer_size = bit_array.byte_size(client.buffer)
  let wanted = int.to_string(size)
  let have = int.to_string(buffer_size)

  log.debug(inspect(client) <> ": read: want " <> wanted <> ", have " <> have)

  case size {
    0 -> Ok(#(client, cipher.Ciphertext(<<>>)))

    n if n <= buffer_size -> {
      let assert <<bits:bytes-size(n), rest:bytes>> = client.buffer
      let new_client = Client(..client, buffer: rest)

      log.debug(
        inspect(client)
        <> ": read: pulled "
        <> int.to_string(n)
        <> " from buffer",
      )

      Ok(#(new_client, cipher.Ciphertext(bits)))
    }

    n if n > buffer_size -> {
      log.debug(inspect(client) <> ": read: reading from socket")

      case underlying_read(client, n - buffer_size) {
        Ok(new_client) -> read(new_client, size)
        Error(error) -> Error(error)
      }
    }

    _ -> panic as "client.read: unreachable case"
  }
}

pub fn write(client: Client, data: Ciphertext) -> Result(Client, error.Error) {
  use n <- result.try(client.writer(data.bits))

  log.debug(inspect(client) <> ": wrote " <> int.to_string(n))

  Ok(client)
}

pub fn decrypt(client: Client, data: Ciphertext) -> #(Client, Plaintext) {
  let #(cipher, plaintext) = cipher.decrypt(client.cipher, data)
  let new_client = with_cipher(client, cipher: cipher)

  #(new_client, plaintext)
}

pub fn encrypt(client: Client, data: Plaintext) -> #(Client, Ciphertext) {
  let #(cipher, ciphertext) = cipher.encrypt(client.cipher, data)
  let new_client = with_cipher(client, cipher: cipher)

  #(new_client, ciphertext)
}
