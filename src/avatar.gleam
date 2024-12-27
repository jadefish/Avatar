import gleam/erlang/process
import gleam/int
import gleam/io
import login_server

// TODO: Make this configurable. The defaults in UO's login.cfg specify 4 login
// servers running on two different ports (7775 and 7776).
const port = 7775

pub fn main() {
  let subject = process.new_subject()

  let _login_server =
    login_server.new(subject, port, 10)
    |> login_server.start()
  io.println("start_login_server(" <> int.to_string(port) <> ")")

  print_auth_result(subject)
}

fn print_auth_result(subject) {
  process.receive_forever(subject) |> io.debug

  print_auth_result(subject)
}
