import cipher
import error
import gleam/bit_array
import utils as u

/// `0x80` Login Request. Length 62, encrypted.
///
/// Sent by clients to a login server. Contains account credentials.
pub type LoginRequest {
  LoginRequest(account: String, password: String, next_key: Int)
}

pub fn decode(data: cipher.Plaintext) -> Result(LoginRequest, error.Error) {
  case data.bits {
    <<0x80, account:30-bytes, password:30-bytes, next_key:int>> -> {
      use account <- u.try_replace(
        account |> u.trim_nul() |> bit_array.to_string(),
        error.DecodeError,
      )
      use password <- u.try_replace(
        password |> u.trim_nul() |> bit_array.to_string(),
        error.DecodeError,
      )

      Ok(LoginRequest(account, password, next_key))
    }

    _ -> Error(error.DecodeError)
  }
}
