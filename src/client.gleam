import aio.{type Closer, type Reader, type Writer}
import cipher.{type Cipher, type Ciphertext, type Plaintext}
import error
import gleam/bit_array
import gleam/int
import gleam/io
import gleam/result
import gleam/string
import youid/uuid

pub const max_packet_size = 0xF000

// TODO: Currently, all outgoing data is immediately written to the client's
// underlying writer. This should likely change to be a periodic process.

pub opaque type Client {
  Client(
    uuid: uuid.Uuid,
    reader: Reader,
    writer: Writer,
    closer: Closer,
    buffer: BitArray,
    cipher: Cipher,
  )
}

pub fn new(reader: Reader, writer: Writer, closer: Closer) {
  Client(uuid.v7(), reader, writer, closer, <<>>, cipher.nil())
}

pub fn with_cipher(client: Client, cipher cipher: cipher.Cipher) -> Client {
  Client(..client, cipher: cipher)
}

pub fn close(client: Client) -> Result(Client, error.Error) {
  case client.closer() {
    Ok(_) -> Ok(client)

    Error(reason) -> {
      // TODO: Is this actually interesting enough to print or bubble up?
      // If the connection can't be closed, I'm not sure much else can be done
      // with it, anyway.
      let reason = string.inspect(reason)
      io.println(inspect(client) <> ": couldn't close connection: " <> reason)
      Error(error.IOError(error.CloseError))
    }
  }
}

pub fn inspect(client: Client) -> String {
  uuid.to_string(client.uuid)
}

fn underlying_read(client: Client, size) -> Result(Client, error.Error) {
  // `size` assumed to be in the range 0 < size < |client.buffer|.

  use data <- result.try(client.reader(size))
  let buffer = <<client.buffer:bits, data:bits>>

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

  io.println(inspect(client) <> ": read: want " <> wanted <> ", have " <> have)

  case size {
    0 -> Ok(#(client, cipher.Ciphertext(<<>>)))

    n if n <= buffer_size -> {
      io.println(inspect(client) <> ": read: pulling from buffer")

      let assert <<bits:bytes-size(n), rest:bytes>> = client.buffer
      let new_client = Client(..client, buffer: rest)

      Ok(#(new_client, cipher.Ciphertext(bits)))
    }

    n if n > buffer_size -> {
      // TODO: Under the "tick" model, this should just wait until the next
      // tick, after which more data may've arrived in the client's buffer.
      io.println(inspect(client) <> ": read: reading from socket")

      case underlying_read(client, n - buffer_size) {
        Ok(new_client) -> read(new_client, size)
        Error(error) -> Error(error)
      }
    }

    _ -> panic as "client.read: unreachable case"
  }
}

pub fn write(client: Client, data: Ciphertext) -> Result(Client, error.Error) {
  use _n <- result.try(client.writer(data.bits))
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
