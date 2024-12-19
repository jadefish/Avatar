import gleam/bool
import gleam/int
import gleam/result

const lo_mask1 = 0x00001357

const lo_mask2 = 0xffffaaaa

const lo_mask3 = 0x0000ffff

const hi_mask1 = 0x43210000

const hi_mask2 = 0xabcdffff

const hi_mask3 = 0xffff0000

// SERENITY NOW!
const bnot = int.bitwise_not

const bor = int.bitwise_or

const band = int.bitwise_and

const bxor = int.bitwise_exclusive_or

const bsl = int.bitwise_shift_left

const bsr = int.bitwise_shift_right

pub type Error {
  InvalidSeed
  UnsupportedVersion
}

pub opaque type Seed {
  Seed(value: Int)
}

pub fn new_seed(value: Int) -> Result(Seed, Error) {
  use <- bool.guard(when: value <= 0, return: Error(InvalidSeed))

  Ok(Seed(value))
}

pub opaque type Version {
  Version(major: Int, minor: Int, patch: Int, revision: Int)
}

pub fn new_version(
  major: Int,
  minor: Int,
  patch: Int,
  revision: Int,
) -> Result(Version, Error) {
  use <- bool.guard(major < 1, return: Error(UnsupportedVersion))
  use <- bool.guard(minor < 0, return: Error(UnsupportedVersion))
  use <- bool.guard(patch < 0, return: Error(UnsupportedVersion))
  use <- bool.guard(revision < 0, return: Error(UnsupportedVersion))

  Ok(Version(major, minor, patch, revision))
}

pub opaque type Cipher {
  NilCipher
  LoginCipher(seed: Seed, mask: KeyPair, key: KeyPair)
}

/// Truncate the provided integer to 32 bits.
fn uint32(int: Int) -> Int {
  band(int, 0xFFFFFFFF)
}

/// Create a new login chiper, capable of working with client packets during
/// authentication against a login server.
pub fn login(seed: Seed, version: Version) -> Result(Cipher, Error) {
  use key <- result.map(key_for_version(version))
  let value = seed.value

  // ((^seed ^ lo_mask1) << 16) | ((seed ^ lo_mask2) & lo_mask3)
  let mask_lo =
    bnot(value)
    |> bxor(lo_mask1)
    |> bsl(16)
    |> bor(value |> bxor(lo_mask2) |> band(lo_mask3))
    |> uint32()

  // ((seed ^ hi_mask1) >> 16) | ((^seed ^ hi_mask2) & hi_mask3)
  let mask_hi =
    value
    |> bxor(hi_mask1)
    |> bsr(16)
    |> bor(bnot(value) |> bxor(hi_mask2) |> band(hi_mask3))
    |> uint32()

  let mask = KeyPair(mask_lo, mask_hi)
  LoginCipher(seed, mask, key)
}

/// Create a new nil cipher, which always returns its input unmodified.
pub fn nil() -> Cipher {
  NilCipher
}

/// Encrypt data using the provided cipher.
///
/// Encryption utilizes a rolling cipher on both ends, so a new Cipher is
/// returned along with the decrypted data. The old Cipher will no longer be
/// capable of encrypting data, so it should be discarded.
pub fn encrypt(cipher: Cipher, plain data: BitArray) -> #(Cipher, BitArray) {
  case cipher {
    NilCipher -> #(cipher, data)
    // The Login cipher doesn't support encrypting data.
    LoginCipher(_, _, _) -> #(cipher, data)
  }
}

/// Decrypt data using the provided cipher.
///
/// Decryption utilizes a rolling cipher on both ends, so a new Cipher is
/// returned along with the decrypted data. The old Cipher will no longer be
/// capable of decrypting data, so it should be discarded.
pub fn decrypt(cipher: Cipher, data: BitArray) -> #(Cipher, BitArray) {
  case cipher {
    NilCipher -> #(cipher, data)
    LoginCipher(seed, mask, key) -> {
      let #(data, new_mask, new_key) = login_decrypt_loop(mask, key, data, <<>>)
      #(LoginCipher(seed, new_mask, new_key), data)
    }
  }
}

fn login_decrypt_loop(
  mask mask: KeyPair,
  key key: KeyPair,
  cipher cipher_bits: BitArray,
  plain plain_bits: BitArray,
) -> #(BitArray, KeyPair, KeyPair) {
  case cipher_bits {
    <<byte, remaining:bytes>> -> {
      // dst[i] = src[i] ^ byte(cs.maskLo)
      let plain_byte = band(mask.lo, 0xFF) |> bxor(byte)

      // cs.maskLo = ((maskLo >> 1) | (maskHi << 31)) ^ cs.keyLo
      let new_mask_lo =
        bor(mask.lo |> bsr(1), mask.hi |> bsl(31))
        |> bxor(key.lo)
        |> uint32()

      // maskHi = ((maskHi >> 1) | (maskLo << 31)) ^ cs.keyHi
      let mask_hi =
        bor(mask.hi |> bsr(1), mask.lo |> bsl(31))
        |> bxor(key.hi)
        |> uint32()

      // cs.maskHi = ((maskHi >> 1) | (maskLo << 31)) ^ cs.keyHi
      let new_mask_hi =
        bor(mask_hi |> bsr(1), mask.lo |> bsl(31))
        |> bxor(key.hi)
        |> uint32()

      let mask = KeyPair(new_mask_lo, new_mask_hi)
      let derp = <<plain_bits:bits, plain_byte>>

      login_decrypt_loop(mask, key, remaining, derp)
    }
    <<>> | _ -> #(plain_bits, mask, key)
  }
}

type KeyPair {
  KeyPair(lo: Int, hi: Int)
}

// from https://github.com/ClassicUO/ClassicUO/blob/3ad74a6/src/ClassicUO.Client/Network/Encryption/Encryption.cs#L67-L98
fn compute_key(a: Int, b: Int, c: Int) -> #(Int, Int, Int) {
  // uint32 ints to 32-bit:
  let a = uint32(a)
  let b = uint32(b)
  let c = uint32(c)

  // temp = (((a << 9) | b) << 10) | c) ^ ((c * c) << 5
  let temp = a |> bsl(9) |> bor(b) |> bsl(10) |> bor(c) |> bxor(c * c |> bsl(5))

  // key2 = (temp << 4) ^ (b * b) ^ (b * 0x0B000000) ^ (c * 0x380000) ^ 0x2C13A5FD
  let key2 =
    bsl(temp, 4)
    |> bxor(b * b)
    |> bxor(b * 0x0B000000)
    |> bxor(c * 0x00380000)
    |> bxor(0x2C13A5FD)

  // temp = (((((a << 9) | c) << 10) | b) * 8) ^ (c * c * 0x0c00)
  let temp =
    { { bsl(a, 9) |> bor(c) |> bsl(10) |> bor(b) } * 8 }
    |> bxor(c * c * 0x00000c00)

  // key3 = temp ^ (b * b) ^ (b * 0x6800000) ^ (c * 0x1c0000) ^ 0x0A31D527F
  let key3 =
    temp
    |> bxor(b * b)
    |> bxor(b * 0x06800000)
    |> bxor(c * 0x001c0000)
    |> bxor(0xA31D527F)

  // key1 = key2 - 1
  let key1 = key2 - 1 |> uint32()

  #(key1, key2, key3)
}

fn key_for_version(version: Version) -> Result(KeyPair, Error) {
  case version {
    // 2.0.3.x is a special case.
    Version(2, 0, 3, 0x78) -> Ok(KeyPair(0x2D13A5FD, 0xA39D527F))
    Version(major, minor, patch, _) -> {
      let #(_, hi, lo) = compute_key(major, minor, patch)
      Ok(KeyPair(lo, hi))
    }
  }
}
