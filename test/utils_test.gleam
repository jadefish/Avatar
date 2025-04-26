import gleeunit
import gleeunit/should
import utils as u

pub fn main() {
  gleeunit.main()
}

pub fn pack_bytes_test() {
  u.pack_bytes(<<192, 168, 68, 56>>) |> should.equal(3_232_252_984)
  u.pack_bytes(<<>>) |> should.equal(0)
  u.pack_bytes(<<1>>) |> should.equal(1)
  u.pack_bytes(<<1, 2, 3, 4, 5, 6, 7, 8, 9, 10>>)
  |> should.equal(4_759_477_275_222_530_853_130)
}
