import client
import gleeunit
import gleeunit/should

pub fn main() {
  gleeunit.main()
}

pub fn read_test() {
  let reader = fn(_n) { Ok(<<1, 2, 3, 4>>) }
  let writer = fn(_bits) { Ok(0) }
  let closer = fn() { Ok(Nil) }
  let client = client.new("test", reader, writer, closer)
  let result = client.read(client, 4)

  result |> should.be_ok

  let assert Ok(#(_, data)) = client.read(client, 4)

  data.bits |> should.equal(<<1, 2, 3, 4>>)
}
