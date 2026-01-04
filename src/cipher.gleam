import error
import gleam/bool
import gleam/bytes_tree.{type BytesTree}
import gleam/int

// TODO: The current implementation of this stream cipher is extremely weak.
//
// 1. Keys are fully determined by (IP, version), so e.g. two users using the
// same client version behind a NAT would use the same keystream.
// 2. There is no entropy in place.
// 3. Key derivation is reversible (no hashing or KDF).
// 4. Mask is derived from seed and will be repeated when seed is repeated.
//
// Stream ciphers are secure only when keystreams don't repeat. Since these are
// based on small deterministic data, they're easily broken.
// Per Wikipedia:
//   > For a stream cipher to be secure, its keystream must have a large period,
//   > and it must be impossible to recover the cipher's key or internal state
//   > from the keystream.
//
// The ciphertext format cannot change, as that'd break client communication.
// However, some hardening internal to the server can be added:
//
// 1. Per-session (or even per-command) nonce/salt XOR'd into the key (after
// compute_key), before creating the mask.
//   * Per Wikipedia: "Securely using a ... stream cipher requires that one
//    never reuse the same keystream twice. ... a different nonce or key must
//    be supplied to each invocation of the cipher."
// 2. Pass (key1, key2, key3) through a KDF (along with the per-session salt/
// nonce) to produce a stronger keystream without changing the ciphertext
// format.
// 3. HMAC ciphertext using server-side key to verify integrity and prevent
// insertion of new messages.
//   * Per Wikipedia: "... stream ciphers provide not authenticity but privacy:
//   encrypted messages may still have been modified in transit."
// 4. Rolling timer- or counter-based rekey (via changing session nonce)
// 5. Future: consider XORing something modern like chacha20 on top of this.

const lo_mask1 = 0x00001357

const lo_mask2 = 0xffffaaaa

const lo_mask3 = 0x0000ffff

const hi_mask1 = 0x43210000

const hi_mask2 = 0xabcdffff

const hi_mask3 = 0xffff0000

// SERENITY NOW!
const not = int.bitwise_not

const or = int.bitwise_or

const and = int.bitwise_and

const xor = int.bitwise_exclusive_or

const shift_left = int.bitwise_shift_left

const shift_right = int.bitwise_shift_right

// TODO: Opaque Seed is becoming clunky to use. It's nice to have "non-negative
// integer" enforced by the constructor function, but is it worth the hassle?

pub opaque type Seed {
  Seed(value: Int)
}

pub fn seed(value: Int) -> Result(Seed, error.Error) {
  use <- bool.guard(when: value <= 0, return: Error(error.InvalidSeed))

  Ok(Seed(value))
}

pub fn seed_value(seed: Seed) -> Int {
  seed.value
}

pub opaque type Version {
  Version(major: Int, minor: Int, patch: Int, revision: Int)
}

pub fn version(
  major: Int,
  minor: Int,
  patch: Int,
  revision: Int,
) -> Result(Version, error.Error) {
  use <- bool.guard(major < 1, return: Error(error.UnsupportedVersion))
  use <- bool.guard(minor < 0, return: Error(error.UnsupportedVersion))
  use <- bool.guard(patch < 0, return: Error(error.UnsupportedVersion))
  use <- bool.guard(revision < 0, return: Error(error.UnsupportedVersion))

  Ok(Version(major, minor, patch, revision))
}

pub opaque type Cipher {
  NilCipher
  LoginCipher(seed: Seed, mask: KeyPair, key: KeyPair)
  // GameCipher
}

/// Truncate the provided integer to 32 bits.
fn uint32(int: Int) -> Int {
  and(int, 0xFFFFFFFF)
}

/// Create a new login chiper, capable of decrypting packets sent during a
/// client's authentication against a login server.
pub fn login(seed: Seed, version: Version) -> Cipher {
  let key = key_for_version(version)
  let value = seed.value

  // ((^seed ^ lo_mask1) << 16) | ((seed ^ lo_mask2) & lo_mask3)
  let mask_lo =
    { not(value) |> xor(lo_mask1) |> shift_left(16) }
    |> or(value |> xor(lo_mask2) |> and(lo_mask3))
    |> uint32()

  // ((seed ^ hi_mask1) >> 16) | ((^seed ^ hi_mask2) & hi_mask3)
  let mask_hi =
    { value |> xor(hi_mask1) |> shift_right(16) }
    |> or(not(value) |> xor(hi_mask2) |> and(hi_mask3))
    |> uint32()

  let mask = KeyPair(mask_lo, mask_hi)
  LoginCipher(seed, mask, key)
}

/// Create a new nil cipher, which always returns its input unmodified.
pub fn nil() -> Cipher {
  NilCipher
}

// pub fn game() -> Cipher {
//   GameCipher
// }

pub type Plaintext {
  Plaintext(bits: BitArray)
}

pub type Ciphertext {
  Ciphertext(bits: BitArray)
}

/// Encrypt data using the provided cipher.
///
/// Encryption utilizes a rolling cipher on both ends, so a new Cipher is
/// returned along with the decrypted data. The old Cipher will no longer be
/// capable of encrypting data, so it should be discarded.
pub fn encrypt(cipher: Cipher, plaintext: Plaintext) -> #(Cipher, Ciphertext) {
  case cipher {
    NilCipher -> #(cipher, Ciphertext(plaintext.bits))
    // The Login cipher doesn't support encrypting data.
    LoginCipher(_, _, _) -> #(cipher, Ciphertext(plaintext.bits))
    // GameCipher -> todo
  }
}

/// Decrypt data using the provided cipher.
///
/// Decryption utilizes a rolling cipher on both ends, so a new Cipher is
/// returned along with the decrypted data. The old Cipher will no longer be
/// capable of decrypting data, so it should be discarded.
pub fn decrypt(cipher: Cipher, ciphertext: Ciphertext) -> #(Cipher, Plaintext) {
  case cipher {
    NilCipher -> #(cipher, Plaintext(ciphertext.bits))
    LoginCipher(seed, mask, key) -> {
      let #(plaintext_bytes, new_mask, new_key) =
        login_decrypt_loop(mask, key, ciphertext.bits, bytes_tree.new())
      #(
        LoginCipher(seed, new_mask, new_key),
        Plaintext(bytes_tree.to_bit_array(plaintext_bytes)),
      )
    }
    // GameCipher -> todo
  }
}

fn login_decrypt_loop(
  mask: KeyPair,
  key: KeyPair,
  ciphertext: BitArray,
  plaintext: BytesTree,
) -> #(BytesTree, KeyPair, KeyPair) {
  case ciphertext {
    <<>> -> #(plaintext, mask, key)

    <<byte:8, remaining_bytes:bytes>> -> {
      // dst[i] = src[i] ^ byte(cs.maskLo)
      let plain_byte = and(mask.lo, 0xFF) |> xor(byte)

      // cs.maskLo = ((maskLo >> 1) | (maskHi << 31)) ^ cs.keyLo
      let new_mask_lo =
        or(mask.lo |> shift_right(1), mask.hi |> shift_left(31))
        |> xor(key.lo)
        |> uint32()

      // maskHi = ((maskHi >> 1) | (maskLo << 31)) ^ cs.keyHi
      let mask_hi =
        or(mask.hi |> shift_right(1), mask.lo |> shift_left(31))
        |> xor(key.hi)
        |> uint32()

      // cs.maskHi = ((maskHi >> 1) | (maskLo << 31)) ^ cs.keyHi
      let new_mask_hi =
        or(mask_hi |> shift_right(1), mask.lo |> shift_left(31))
        |> xor(key.hi)
        |> uint32()

      let new_mask = KeyPair(new_mask_lo, new_mask_hi)
      let decrypted_bytes = bytes_tree.append(plaintext, <<plain_byte>>)

      login_decrypt_loop(new_mask, key, remaining_bytes, decrypted_bytes)
    }

    _ -> panic as "cipher.login_decrypt_loop: found unaligned bit array"
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
  let temp =
    a
    |> shift_left(9)
    |> or(b)
    |> shift_left(10)
    |> or(c)
    |> xor(c * c |> shift_left(5))

  // key2 = (temp << 4) ^ (b * b) ^ (b * 0x0B000000) ^ (c * 0x380000) ^ 0x2C13A5FD
  let key2 =
    shift_left(temp, 4)
    |> xor(b * b)
    |> xor(b * 0x0B000000)
    |> xor(c * 0x00380000)
    |> xor(0x2C13A5FD)

  // temp = (((((a << 9) | c) << 10) | b) * 8) ^ (c * c * 0x0c00)
  let temp =
    { { shift_left(a, 9) |> or(c) |> shift_left(10) |> or(b) } * 8 }
    |> xor(c * c * 0x00000c00)

  // key3 = temp ^ (b * b) ^ (b * 0x6800000) ^ (c * 0x1c0000) ^ 0x0A31D527F
  let key3 =
    temp
    |> xor(b * b)
    |> xor(b * 0x06800000)
    |> xor(c * 0x001c0000)
    |> xor(0xA31D527F)

  // key1 = key2 - 1
  let key1 = key2 - 1 |> uint32()

  #(key1, key2, key3)
}

fn key_for_version(version: Version) -> KeyPair {
  case version {
    // 2.0.3.x (literal "x") is a special case.
    Version(2, 0, 3, 0x78) -> KeyPair(0x2D13A5FD, 0xA39D527F)
    Version(major, minor, patch, _) -> {
      let #(_, hi, lo) = compute_key(major, minor, patch)
      KeyPair(lo, hi)
    }
  }
}
