import game_server
import gleam/erlang/process
import gleam/io
import gleam/list
import gleam/string
import login_server
import time_zone as tz

// TODO: Make login ports configurable. The defaults in UO's login.cfg specify 4
// login servers running on two different ports (7775 and 7776).

pub fn main() {
  let auth_chan = process.new_subject()
  let login_servers = [
    login_server.new(auth_chan, port: 7775, pool_size: 10),
    login_server.new(auth_chan, port: 7776, pool_size: 10),
  ]
  let game_chan = process.new_subject()
  let game_servers = [
    game_server.new(
      game_chan,
      port: 7080,
      pool_size: 10,
      time_zone: tz.AmericaDetroit,
      capacity: 10_000,
    ),
  ]

  let login_pids =
    list.map(login_servers, fn(server) {
      process.start(fn() { login_server.start(server) }, linked: True)
    })

  echo login_pids

  let game_pids =
    list.map(game_servers, fn(server) {
      process.start(fn() { game_server.start(server) }, linked: True)
    })

  echo game_pids

  process.start(fn() { print_auth_result(auth_chan) }, linked: True)
  process.sleep_forever()
}

fn print_auth_result(chan: process.Subject(login_server.LoginResult)) {
  case process.receive_forever(chan) {
    Ok(client) ->
      io.println(
        "Successfully authenticated client: " <> string.inspect(client),
      )
    Error(error) -> {
      // TODO: Probably need a reference to the errant Client in here.
      io.println("Auth error: " <> string.inspect(error))
    }
  }

  print_auth_result(chan)
}
