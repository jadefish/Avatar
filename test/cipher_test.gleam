import cipher
import gleeunit
import gleeunit/should

pub fn main() {
  gleeunit.main()
}

pub fn new_seed_test() {
  cipher.new_seed(-10) |> should.be_error()
  cipher.new_seed(0) |> should.be_error()
  cipher.new_seed(10) |> should.be_ok()
}

pub fn new_version_test() {
  cipher.new_version(-1, -1, -1, -1) |> should.be_error()
  cipher.new_version(0, 0, 0, 0) |> should.be_error()
  cipher.new_version(1, 0, 0, 0) |> should.be_ok()
  cipher.new_version(7, 0, 62, 0) |> should.be_ok()
}

pub fn login_decrypt_test() {
  // 192.168.68.60
  let seed = 0xC0_A8_44_3C
  let assert Ok(seed) = cipher.new_seed(seed)
  let assert Ok(version) = cipher.new_version(7, 0, 106, 21)
  let ciphertext = <<
    22, 85, 134, 110, 22, 182, 112, 132, 182, 142, 146, 155, 168, 43, 234, 138,
    186, 34, 238, 8, 251, 130, 62, 96, 207, 24, 115, 70, 220, 145, 55, 148, 108,
    138, 240, 201, 79, 29, 172, 42, 192, 181, 136, 33, 111, 72, 91, 210, 150, 52,
    229, 13, 121, 195, 30, 240, 135, 188, 161, 175, 168, 43,
  >> |> cipher.CipherText()
  let assert Ok(cipher) = cipher.login(seed, version)
  let #(_cipher, plaintext) = cipher.decrypt(cipher, ciphertext)
  let assert <<
    cmd:int,
    account_name:bytes-size(30),
    password:bytes-size(30),
    next_key:int,
  >> = plaintext.bits

  cmd |> should.equal(0x80)
  account_name |> should.equal(<<"account1234", 0x00:8-unit(19)>>)
  password |> should.equal(<<"password1234", 0x00:8-unit(18)>>)
  next_key |> should.equal(0x00)
}
