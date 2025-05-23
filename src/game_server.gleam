import cipher
import client
import gleam/bit_array
import gleam/erlang/process.{type Subject}
import gleam/int
import gleam/io
import gleam/option.{None}
import gleam/otp/actor
import gleam/string
import glisten
import time_zone as tz
import utils as u

pub opaque type Server {
  StoppedServer(
    parent: process.Subject(Nil),
    port: Int,
    pool_size: Int,
    time_zone: tz.TimeZone,
    capacity: Int,
  )

  StartedServer(
    parent: process.Subject(Nil),
    port: Int,
    pool_size: Int,
    time_zone: tz.TimeZone,
    capacity: Int,
    server: glisten.Server,
  )
}

pub type Action {
  Start
  Stop
}

pub fn new(
  parent: Subject(Nil),
  port port: Int,
  pool_size pool_size: Int,
  time_zone time_zone: tz.TimeZone,
  capacity capacity: Int,
) {
  let server = StoppedServer(parent, port, pool_size, time_zone, capacity)
  let assert Ok(subject) = actor.start(server, loop)
  subject
}

pub fn start(server: Subject(Action)) {
  actor.send(server, Start)
  server
}

pub fn stop(server: Subject(Action)) {
  actor.send(server, Stop)
  server
}

fn loop(action: Action, server: Server) {
  case action {
    Start ->
      case server {
        StartedServer(_, _, _, _, _, _) -> actor.continue(server)

        StoppedServer(parent, port, pool_size, tz, capacity) -> {
          let init = fn(conn) {
            let addr = u.connection_addr(glisten.get_client_info(conn))
            io.println(inspect(server) <> ": new connection: " <> addr)
            #(server, None)
          }

          let result =
            glisten.handler(init, handle_message)
            // UO clients don't seem to do IPv6.
            |> glisten.bind("0.0.0.0")
            |> glisten.with_pool_size(server.pool_size)
            |> glisten.start_server(server.port)

          case result {
            Ok(glisten_server) -> {
              let server =
                StartedServer(
                  parent,
                  port: port,
                  pool_size: pool_size,
                  time_zone: tz,
                  capacity: capacity,
                  server: glisten_server,
                )
              io.println(inspect(server) <> ": started")
              actor.continue(server)
            }

            Error(error) -> actor.Stop(process.Abnormal(string.inspect(error)))
          }
        }
      }

    Stop -> actor.Stop(process.Normal)
  }
}

fn inspect(server: Server) -> String {
  case server {
    StartedServer(_, _, _, _, _, server) -> {
      let addr = u.connection_addr(glisten.get_server_info(server, 5))
      "game_server(" <> addr <> ", " <> string.inspect(process.self()) <> ")"
    }

    StoppedServer(_, port, _, _, _) ->
      "game_server(stopped:" <> int.to_string(port) <> ")"
  }
}

fn handle_message(message, server: Server, conn) {
  // User-type messages are never sent to the server's subject, so this
  // assertion is safe.
  let assert glisten.Packet(bits) = message
  let client = client.Client(conn, None, bits, <<>>, cipher.nil())
  let client_addr = u.connection_addr(glisten.get_client_info(client.conn))
  let size = bit_array.byte_size(bits)

  // expecting 4 + 65 bytes: seed (IP, little endian; or whatever login server
  // sent as new_key pre-relay?), plus account credentials?

  // game_server(stopped:7080): 127.0.0.1:58638: 69 bytes:
  //   <<
  //     1, 0, 0, 127,
  //     109, 131, 198, 172, 207, 117, 2, 250, 49, 125, 186, 84, 51, 118, 26, 115, 107, 206, 42, 7, 6, 176, 128, 133, 139, 65, 140, 132, 74, 197,
  //     240, 191, 172, 18, 237, 82, 82, 99, 92, 165, 201, 234, 134, 216, 81, 194, 175, 195, 255, 22, 79, 214, 111, 224, 124, 102, 111, 146, 101, 164,
  //     93, 9, 166, 12, 177
  //   >>

  io.println(
    inspect(server)
    <> ": "
    <> client_addr
    <> ": "
    <> int.to_string(size)
    <> " bytes:\n\t"
    <> bit_array.inspect(bits),
  )

  actor.continue(server)
}
