import ansi
import client
import gleam/bit_array
import gleam/erlang/process.{type Subject}
import gleam/int
import gleam/io
import gleam/option
import gleam/otp/actor
import gleam/string
import glisten
import log
import tcp
import time_zone as tz

pub opaque type Server {
  Server(
    parent: process.Subject(Nil),
    port: Int,
    pool_size: Int,
    time_zone: tz.TimeZone,
    capacity: Int,
    clients: List(client.Client),
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
  let server = Server(parent, port, pool_size, time_zone, capacity, [])
  let assert Ok(actor.Started(_pid, subject)) =
    actor.new(server)
    |> actor.on_message(loop)
    |> actor.start
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

fn loop(server: Server, action: Action) -> actor.Next(Server, Action) {
  case action {
    Start -> {
      let init = fn(conn) {
        let addr = tcp.client_addr_string(conn)
        log.info(inspect(server) <> ": new connection: " <> addr)
        #(server, option.None)
      }

      let result =
        glisten.new(init, handle_message)
        // UO clients don't seem to do IPv6.
        |> glisten.bind("0.0.0.0")
        |> glisten.with_pool_size(server.pool_size)
        |> glisten.start(server.port)

      case result {
        Ok(actor.Started(_pid, _supervisor)) -> {
          log.info(inspect(server) <> ": listening")
          actor.continue(server)
        }

        Error(error) -> actor.stop_abnormal(string.inspect(error))
      }
    }

    Stop -> actor.stop()
  }
}

fn inspect(server: Server) -> String {
  // TODO: get IP from glisten? not sure how to get its process name for
  // glisten's get_server_info function.
  ansi.bold(ansi.green("GAME")) <> " 0.0.0.0:" <> int.to_string(server.port)
}

fn handle_message(
  server: Server,
  message: glisten.Message(a),
  conn: glisten.Connection(b),
) -> glisten.Next(Server, glisten.Message(a)) {
  // User-type messages are never sent to the server's subject, so this
  // assertion is safe.
  let assert glisten.Packet(bits) = message
  let addr = tcp.client_addr_string(conn)

  let client =
    client.new(
      addr,
      tcp.reader(conn.socket, timeout: 5000),
      tcp.writer(conn.socket),
      tcp.closer(conn.socket),
    )

  // TODO: surely there's a better way to update these records.
  let server = Server(..server, clients: [client, ..server.clients])
  // let client_addr = tcp.connection_addr(glisten.get_client_info(conn))
  // let client_addr = glisten.get_client_info(conn) |> result.map(fn(ci) { tcp.connection_addr2(ci) }) |> result.unwrap("(unknown)")
  // let client_addr = tcp.socket_addr(tcp.Client(conn))
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
    <> "client"
    // TODO
    <> ": "
    <> int.to_string(size)
    <> " bytes:\n\t"
    <> bit_array.inspect(bits),
  )

  glisten.continue(server)
}
