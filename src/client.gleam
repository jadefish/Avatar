import aio.{type Closer, type Reader, type Writer}
import cipher.{type Cipher}
import error
import gleam/bit_array
import gleam/int
import gleam/io
import gleam/result
import gleam/string
import youid/uuid

pub const max_packet_size = 0xF000

pub opaque type Client {
  Client(
    uuid: uuid.Uuid,
    reader: Reader,
    writer: Writer,
    closer: Closer,
    inbox: BitArray,
    outbox: BitArray,
    cipher: Cipher,
  )
}

pub fn new(reader: Reader, writer: Writer, closer: Closer) {
  Client(uuid.v7(), reader, writer, closer, <<>>, <<>>, cipher.nil())
}

pub fn with_cipher(client: Client, cipher cipher: cipher.Cipher) -> Client {
  Client(..client, cipher: cipher)
}

pub fn close(client: Client) -> Nil {
  case client.closer() {
    Ok(_) -> Nil

    Error(reason) -> {
      // TODO: Is this actually interesting enough to print or bubble up?
      // If the connection can't be closed, I'm not sure much else can be done
      // with it, anyway.
      let reason = string.inspect(reason)
      io.println(inspect(client) <> ": couldn't close connection: " <> reason)
      Nil
    }
  }
}

pub fn inspect(client: Client) -> String {
  uuid.to_string(client.uuid)
}

fn underlying_read(client: Client, size) -> Result(Client, error.Error) {
  // `size` assumed to be in the range 0 < size < |client.inbox|.

  use data <- result.try(client.reader(size))
  let inbox = <<client.inbox:bits, data:bits>>

  Ok(Client(..client, inbox:))
}

pub fn read(
  client: Client,
  size: Int,
) -> Result(#(Client, cipher.Ciphertext), error.Error) {
  let size = int.clamp(size, min: 0, max: max_packet_size)
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
      // TODO: In the "tick" model, this shouldn't read from the client's
      // reader. Instead, it should just wait until the next tick when more data
      // may've arrived in the client's inbox.
      io.println(inspect(client) <> ": read: reading from socket")

      case underlying_read(client, n - inbox_size) {
        Ok(new_client) -> read(new_client, size)
        Error(error) -> Error(error)
      }
    }

    _ -> panic as "client.read: unreachable case"
  }
}

pub fn write(client: Client, ciphertext: cipher.Ciphertext) -> Client {
  Client(..client, outbox: <<client.outbox:bits, ciphertext.bits:bits>>)
}
