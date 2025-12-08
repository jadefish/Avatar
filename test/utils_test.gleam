import gleeunit
import gleeunit/should
import utils.{pack_bytes}

pub fn main() {
  gleeunit.main()
}

pub fn pack_bytes_test() {
  pack_bytes(<<192, 168, 68, 56>>) |> should.equal(3_232_252_984)
  pack_bytes(<<>>) |> should.equal(0)
  pack_bytes(<<1>>) |> should.equal(1)
  pack_bytes(<<1, 2, 3, 4, 5, 6, 7, 8, 9, 10>>)
  |> should.equal(4_759_477_275_222_530_853_130)
}
