import cipher
import error

/// `0xA0` Select Game Server. Length 3, encrypted.
///
/// Sent by a client to a login server after the client has selected a shard (a
/// game server) on which to play.
pub type SelectGameServer {
  SelectGameServer(index: Int)
}

pub const length = 3

pub fn decode(data: cipher.Plaintext) -> Result(SelectGameServer, error.Error) {
  case data.bits {
    <<0xA0, index:16>> -> Ok(SelectGameServer(index))
    _ -> Error(error.DecodeError)
  }
}
