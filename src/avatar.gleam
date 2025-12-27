import envoy
import game_server
import gleam/erlang/process
import gleam/list
import gleam/string
import log
import logging
import login_server
import time_zone as tz

// TODO: Make login ports configurable. The defaults in UO's login.cfg specify 4
// login servers running on two different ports (7775 and 7776).

fn get_log_level() -> logging.LogLevel {
  case envoy.get("AVATAR_LOG_LEVEL") {
    Ok(level_string) -> {
      case level_string |> string.lowercase {
        "alert" -> logging.Alert
        "critical" -> logging.Critical
        "debug" -> logging.Debug
        "emergency" -> logging.Emergency
        "error" -> logging.Error
        "notice" -> logging.Notice
        "warning" -> logging.Warning
        _ -> logging.Info
      }
    }

    Error(_) -> logging.Info
  }
}

pub fn main() {
  let log_level = get_log_level()
  log.configure(log_level)
  log.debug("Log level set to " <> string.inspect(log_level))

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

  let _login_pids =
    list.map(login_servers, fn(server) {
      process.spawn(fn() { login_server.start(server) })
    })

  let _game_pids =
    list.map(game_servers, fn(server) {
      process.spawn(fn() { game_server.start(server) })
    })

  process.spawn(fn() { print_auth_result(auth_chan) })
  process.sleep_forever()
}

fn print_auth_result(chan: process.Subject(login_server.LoginResult)) {
  case process.receive_forever(chan) {
    Ok(client) ->
      log.info("Successfully authenticated client: " <> string.inspect(client))
    Error(error) -> {
      // TODO: Probably need a reference to the errant Client in here.
      log.warning("Auth error: " <> string.inspect(error))
    }
  }

  print_auth_result(chan)
}
