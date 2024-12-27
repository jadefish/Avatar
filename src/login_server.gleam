import cipher
import client.{type Client}
import gleam/bit_array
import gleam/erlang/process.{type Subject}
import gleam/int
import gleam/io
import gleam/option
import gleam/otp/actor
import gleam/result
import gleam/string
import glisten
import glisten/socket
import glisten/tcp
import packets
import packets/game_server_list
import packets/login_request
import packets/login_seed
import time_zone as tz
import utils as u

/// A login server accepts a port and a parent subject. Once started, the server
/// authenticates connecting clients. Clients that have been successfully
/// authenticated are ready to be relayed to a game server.
pub opaque type Server {
  StoppedServer(parent: Subject(LoginResult), port: Int, pool_size: Int)
  StartedServer(parent: Subject(LoginResult), server: glisten.Server)
}

pub type Error {
  NetError(socket.SocketReason)
  PacketError(packets.Error)
}

pub type LoginResult =
  Result(Client, AuthenticationError)

pub type AuthenticationError {
  InvalidCredentals
  AccountInUse
  AccountBanned

  Incomplete(error: Error)
}

pub opaque type Action {
  Start
  Stop
}

pub fn new(parent: Subject(LoginResult), port: Int, pool_size: Int) {
  let server = StoppedServer(parent, port, pool_size)
  let assert Ok(subject) = actor.start(server, handle_message)
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

fn connection_addr(conn: glisten.Connection(a)) {
  glisten.get_client_info(conn)
  |> result.map(fn(info) {
    glisten.ip_address_to_string(info.ip_address)
    <> ":"
    <> int.to_string(info.port)
  })
  |> result.unwrap("(unknown)")
}

fn handle_message(action: Action, server: Server) {
  case action {
    Start ->
      case server {
        // Starting an already-started server does nothing.
        StartedServer(_, _) -> actor.continue(server)

        StoppedServer(parent, port, pool_size) -> {
          let init = fn(conn) {
            let addr = connection_addr(conn)
            io.println("login_server: new connection: " <> addr)
            #(StoppedServer(parent, port, pool_size), option.None)
          }

          let result =
            glisten.handler(init, message_handler)
            |> glisten.with_pool_size(pool_size)
            |> glisten.start_server(port)

          io.println("login_server: listening on port " <> int.to_string(port))

          case result {
            Ok(server) -> actor.continue(StartedServer(parent, server))
            Error(error) -> actor.Stop(process.Abnormal(string.inspect(error)))
          }
        }
      }

    Stop -> actor.Stop(process.Normal)
  }
}

const test_ip = #(127, 0, 0, 1)

// TODO: fetch from ... somehere.
const game_servers = [
  game_server_list.GameServer("US East", tz.AmericaDetroit, test_ip, 7775),
  game_server_list.GameServer("US West", tz.AmericaLosAngeles, test_ip, 7775),
]

fn message_handler(message, server: Server, conn) {
  // User-type messages are never sent to the server's subject, so this
  // assertion is safe.
  let assert glisten.Packet(bits) = message

  let result = case bits {
    <<0xEF, _:bits>> -> {
      let client = client.Client(conn, bits, <<>>, cipher.nil())
      use client <- result.try(handle_login_seed(client))
      use client <- result.try(handle_login_request(client))
      use client <- result.try(send_game_server_list(client, game_servers))
      Ok(client)
    }

    bits -> {
      io.println("login_server: bad packet: " <> bit_array.inspect(bits))
      let _ = tcp.close(conn)
      // TODO: need to handle error?
      Error(NetError(socket.Closed))
    }
  }

  case result {
    Ok(client) -> actor.send(server.parent, Ok(client))
    Error(error) -> actor.send(server.parent, Error(Incomplete(error)))
  }

  actor.continue(server)
}

fn client_error_to_net_error(error: client.Error) {
  // TODO: this is gross
  case error {
    client.ReadError(reason) -> NetError(reason)
    // TODO: this isn't right
    _ -> NetError(socket.Closed)
  }
}

fn handle_login_seed(client: Client) {
  // Receive 0xEF Login Seed (unencrypted, length 21):
  use #(client, bits) <- u.try_map(
    client.read(client, 21),
    client_error_to_net_error,
  )
  use login_seed <- u.try_map(login_seed.decode(bits), PacketError)
  io.debug(login_seed)
  let cipher = cipher.login(login_seed.seed, login_seed.version)
  Ok(client.Client(..client, cipher:))
}

fn handle_login_request(client: Client) {
  // Receive 0x80 Login Request (encrypted, length 62):
  use #(client, bits) <- u.try_map(
    client.read(client, 62),
    client_error_to_net_error,
  )
  let #(cipher, plaintext) =
    cipher.decrypt(client.cipher, cipher.CipherText(bits))
  use login_request <- u.try_map(
    login_request.decode(plaintext.bits),
    PacketError,
  )
  io.debug(login_request)

  // TODO: authenticate the client.
  // 1. credential match
  // 2. IP-in-use check
  // 3. ban check

  Ok(client.Client(..client, cipher:))
}

fn send_game_server_list(
  client: Client,
  game_servers: List(game_server_list.GameServer),
) {
  // Send 0xA8 Game Server List (unencrypted, variable length):
  let game_server_list = game_server_list.GameServerList(game_servers)
  io.debug(game_server_list)
  use bytes <- u.try_map(game_server_list.encode(game_server_list), PacketError)

  client.write(client, bytes) |> result.map_error(client_error_to_net_error)
}
