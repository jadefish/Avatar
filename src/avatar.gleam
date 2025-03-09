import gleam/erlang/process
import gleam/io
import gleam/string
import login_server

// TODO: Make this configurable. The defaults in UO's login.cfg specify 4 login
// servers running on two different ports (7775 and 7776).
const port = 7775

pub fn main() {
  let subject = process.new_subject()

  let _login_server =
    login_server.new(subject, port, 10)
    |> login_server.start()

  print_auth_result(subject)
}

fn print_auth_result(subject: process.Subject(login_server.LoginResult)) {
  case process.receive_forever(subject) {
    Ok(client) ->
      io.println(
        "Successfully authenticated client: " <> string.inspect(client),
      )
    Error(error) -> {
      // TODO: Probably need a reference to the errant Client in here.
      io.println("Auth error: " <> string.inspect(error))
    }
  }

  print_auth_result(subject)
}
