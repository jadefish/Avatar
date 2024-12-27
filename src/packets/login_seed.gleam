import cipher
import gleam/io
import packets
import utils as u

pub type LoginSeed {
  /// 0xEF Login Seed. Length 21, unencrypted.
  ///
  /// The first packet sent from a client to a login server. It is used to
  /// identify the client and establish a cipher.
  LoginSeed(seed: cipher.Seed, version: cipher.Version)
}

// TODO: bits should be strongly typed as plaintext, not ciphertext. (i.e.,
// this function doesn't do any decryption.)
// TODO: use decoder? https://discord.com/channels/768594524158427167/768594524158427170/1323780758401191966
pub fn decode(bits: BitArray) -> Result(LoginSeed, packets.Error) {
  io.debug(bits)
  case bits {
    <<0xEF, seed:32, major:32, minor:32, patch:32, revision:32>> -> {
      use seed <- u.try_replace(cipher.seed(seed), packets.DecodeError)
      use version <- u.try_replace(
        cipher.version(major, minor, patch, revision),
        packets.DecodeError,
      )

      Ok(LoginSeed(seed, version))
    }
    _ -> Error(packets.DecodeError)
  }
}
