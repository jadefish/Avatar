import gleam/bit_array
import packets
import utils as u

pub type LoginRequest {
  /// 0x80 Login Request. Length 62, encrypted.
  ///
  /// Sent by clients to a login server. Contains account credentials
  LoginRequest(account: String, password: String, next_key: Int)
}

// TODO: bits should be strongly typed as plaintext, not ciphertext. (i.e.,
// this function doesn't do any decryption.)
pub fn decode(bits: BitArray) -> Result(LoginRequest, packets.Error) {
  case bits {
    <<0x80, account:30-bytes, password:30-bytes, next_key:int>> -> {
      use account <- u.try_replace(
        account |> u.trim_nul() |> bit_array.to_string(),
        packets.DecodeError,
      )
      use password <- u.try_replace(
        password |> u.trim_nul() |> bit_array.to_string(),
        packets.DecodeError,
      )

      Ok(LoginRequest(account, password, next_key))
    }
    _ -> Error(packets.DecodeError)
  }
}
