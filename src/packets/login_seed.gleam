import cipher
import error
import gleam/result
import utils as u

/// `0xEF` Login Seed. Length 21, unencrypted.
///
/// The first packet sent from a client to a login server. It is used to
/// identify the client and establish a cipher.
pub type LoginSeed {
  LoginSeed(seed: cipher.Seed, version: cipher.Version)
}

pub fn decode(data: cipher.PlainText) -> Result(LoginSeed, error.Error) {
  case data.bits {
    <<0xEF, seed:32, major:32, minor:32, patch:32, revision:32>> -> {
      use seed <- result.try(cipher.seed(seed))
      use version <- u.try_replace(
        cipher.version(major, minor, patch, revision),
        error.DecodeError,
      )

      Ok(LoginSeed(seed, version))
    }

    _ -> Error(error.DecodeError)
  }
}
