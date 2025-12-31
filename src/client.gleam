import aio.{type Closer, type Writer}
import cipher.{type Cipher, type Ciphertext, type Plaintext}
import error
import gleam/int
import gleam/result
import log

pub type Client {
  Client(
    // For now, a TCP client socket "IP:port" string
    id: String,
    writer: Writer,
    closer: Closer,
    buffer: BitArray,
    cipher: Cipher,
  )
}

pub fn new(id: String, writer: Writer, closer: Closer) {
  Client(id, writer, closer, <<>>, cipher.nil())
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
