import ansi
import cipher
import client.{type Client}
import error
import gleam/bit_array
import gleam/crypto
import gleam/dict
import gleam/erlang/process.{type Subject}
import gleam/int
import gleam/list
import gleam/option.{None}
import gleam/otp/actor
import gleam/result
import gleam/string
import glisten
import log
import packets/connect_to_game_server
import packets/game_server_list
import packets/login_denied
import packets/login_request
import packets/login_seed
import packets/select_game_server
import tcp
import time_zone as tz
import utils as u

// TODO: need listener import for get_server_info, but it's internal.
import glisten/internal/listener

/// A login server accepts a port and a parent subject. Once started, the server
/// authenticates connecting clients. Clients that have been successfully
/// authenticated are ready to be relayed to a game server.
pub opaque type Server {
  Server(
    parent: Subject(LoginResult),
    listener_name: process.Name(listener.Message),
    port: Int,
    pool_size: Int,
    clients: dict.Dict(glisten.ConnectionInfo, Client),
  )
}

pub type LoginResult =
  Result(Client, error.Error)

pub opaque type Action {
  Start
  Stop
}

pub fn new(
  parent: Subject(LoginResult),
  port port: Int,
  pool_size pool_size: Int,
) -> Subject(Action) {
  let name = process.new_name("login_server")
  let server = Server(parent, name, port, pool_size, dict.new())
  let assert Ok(actor.Started(_pid, subject)) =
    actor.new(server)
    |> actor.on_message(loop)
    |> actor.start
  subject
}

pub fn start(subject: Subject(Action)) -> Subject(Action) {
  actor.send(subject, Start)
  subject
}

pub fn stop(subject: Subject(Action)) -> Subject(Action) {
  actor.send(subject, Stop)
  subject
}

fn inspect(server: Server) -> String {
  let glisten.ConnectionInfo(port, ip) =
    glisten.get_server_info(server.listener_name, 1000)
  let ip = glisten.ip_address_to_string(ip)
  let port = int.to_string(port)

  ansi.bold(ansi.magenta("LOGIN")) <> " " <> ip <> ":" <> port
}

fn find_client(
  server: Server,
  conn: glisten.Connection(a),
) -> Result(Client, Nil) {
  let assert Ok(info) = glisten.get_client_info(conn)
  dict.get(server.clients, info)
}

fn loop(server: Server, action: Action) -> actor.Next(Server, Action) {
  case action {
    Start -> {
      let init = fn(conn) {
        let addr = tcp.client_addr_string(conn)
        log.info(inspect(server) <> ": new connection: " <> addr)

        let client =
          client.new(
            addr,
            tcp.reader(conn.socket, timeout: 5000),
            tcp.writer(conn.socket),
            tcp.closer(conn.socket),
          )
        let assert Ok(connection_info) = glisten.get_client_info(conn)
        let clients = dict.insert(server.clients, connection_info, client)
        let new_server = Server(..server, clients: clients)
        #(new_server, None)
      }

      let result =
        glisten.new(init, handle_message)
        |> glisten.bind("0.0.0.0")
        |> glisten.with_pool_size(server.pool_size)
        // TODO: |> glisten.with_close(...) 
        // TODO: start_with_listener_name is marked @internal, but it's used
        // in one of glisten's example programs. also, the listener_name is
        // eventually needed by the inspect function, above.
        |> glisten.start_with_listener_name(server.port, server.listener_name)

      case result {
        Ok(_supervisor) -> {
          log.info(inspect(server) <> ": listening")
          actor.continue(server)
        }

        Error(error) -> {
          log.error(
            inspect(server) <> ": start error: " <> string.inspect(error),
          )
          actor.stop_abnormal(string.inspect(error))
        }
      }
    }

    Stop -> {
      actor.stop()
    }
  }
}

const test_ip = #(127, 0, 0, 1)

// TODO: fetch from ... somehere.
const game_servers = [
  game_server_list.GameServer("US East", tz.AmericaDetroit, test_ip, 7080),
  game_server_list.GameServer("US West", tz.AmericaLosAngeles, test_ip, 7081),
]

fn handle_message(
  server: Server,
  message: glisten.Message(a),
  conn: glisten.Connection(b),
) -> glisten.Next(Server, glisten.Message(a)) {
  // User-type messages are never sent to the server's subject, so this
  // assertion is safe.
  let assert glisten.Packet(bits) = message

  // TODO: Reconsider this design. Currently, there are two ways to read data
  // from a client: glisten's message handler (this function) and the
  // client.read function.
  //
  // Instead, maybe maintain an internal "authentication step" state variable
  // for each client to inform which packet is expected to be received and what
  // to send back.
  //
  // Occasionally, a client will send the 0xD9 Spy On Client packet during
  // login, which implies that login server packets were meant to be handled
  // out-of-order, anyway.

  use client <- u.lazy_unwrap_error(
    find_client(server, conn) |> result.replace_error(glisten.continue(server)),
  )

  // TODO: this doesn't work when Client is opaque (which is should be)
  let client = client.Client(..client, buffer: bits)

  let result = case bits {
    <<0xEF, _:bits>> -> {
      // TODO: a client sending 0xD9 Spy On Client will break this process.
      use client <- result.try(handle_login_seed(server, client))
      use client <- result.try(handle_login_request(server, client))
      use client <- result.try(send_game_server_list(
        server,
        client,
        game_servers,
      ))
      // Client may send 0xD9 before Select Game Server here.
      use #(client, game_server) <- result.try(select_game_server(
        server,
        client,
      ))
      use client <- result.try(send_connect_to_game_server(
        server,
        client,
        game_server,
      ))

      Ok(client)
    }

    bits -> {
      log.warning(
        inspect(server) <> ": bad packet: " <> bit_array.inspect(bits),
      )
      Error(error.UnexpectedPacket)
    }
  }

  let result = case result {
    Ok(client) -> client.close(client)

    Error(error) ->
      case error {
        error.AuthenticationError(auth_error) -> {
          case auth_error {
            error.AccountBanned ->
              deny_login(server, client, login_denied.AccountBanned)

            error.AccountInUse ->
              deny_login(server, client, login_denied.AccountInUse)

            error.InvalidCredentals ->
              deny_login(server, client, login_denied.InvalidCredentials)
          }
        }

        error.DecodeError
        | error.EncodeError
        | error.IOError(_)
        | error.UnexpectedPacket
        | error.InvalidSeed -> {
          log.error(
            inspect(server)
            <> ": error handling message: "
            <> string.inspect(error),
          )
          deny_login(server, client, login_denied.CommunicationProblem)
        }

        _ -> deny_login(server, client, login_denied.GenericDenial)
      }
  }

  actor.send(server.parent, result)
  glisten.continue(server)
}

fn handle_login_seed(
  server: Server,
  client: Client,
) -> Result(Client, error.Error) {
  // Receive 0xEF Login Seed (unencrypted, length 21):
  use #(client, bits) <- result.try(client.read(client, 21))
  let plaintext = cipher.Plaintext(bits.bits)
  use login_seed <- result.try(login_seed.decode(plaintext))
  let cipher = cipher.login(login_seed.seed, login_seed.version)

  log.info(
    inspect(server)
    <> ": got Login Seed from client "
    <> client.inspect(client)
    <> ": "
    <> string.inspect(login_seed),
  )

  Ok(client.with_cipher(client, cipher: cipher))
}

fn handle_login_request(
  server: Server,
  client: Client,
) -> Result(Client, error.Error) {
  // Receive 0x80 Login Request (encrypted, length 62):
  use #(client, bits) <- result.try(client.read(client, 62))
  let #(client, plaintext) = client.decrypt(client, bits)
  use login_request <- result.try(login_request.decode(plaintext))

  // TODO: mask password in this output
  log.info(
    inspect(server)
    <> ": got Login Request from client "
    <> client.inspect(client)
    <> ": "
    <> login_request.inspect(login_request),
  )

  // TODO: Authenticate the client:
  // 1. credential match
  // 2. ban check
  // 3. account-in-use check

  Ok(client)
}

fn deny_login(
  server: Server,
  client: Client,
  reason: login_denied.Reason,
) -> Result(Client, error.Error) {
  let packet = login_denied.LoginDenied(reason)
  let plaintext = login_denied.encode(packet)

  log.info(
    inspect(server)
    <> ": sending Login Denied to client "
    <> client.inspect(client)
    <> ": "
    <> string.inspect(packet),
  )

  client.write(client, cipher.Ciphertext(plaintext.bits))
  // No need to close the connection here – client closes its end.
}

fn send_game_server_list(
  server: Server,
  client: Client,
  game_servers: List(game_server_list.GameServer),
) {
  let game_server_list =
    game_server_list.GameServerList(
      game_servers,
      game_server_list.DoNotSendSystemInfo,
    )

  log.info(
    inspect(server)
    <> ": sending Game Server List to client "
    <> client.inspect(client)
    <> ": "
    <> string.inspect(game_server_list),
  )

  let plaintext = game_server_list.encode(game_server_list)
  client.write(client, cipher.Ciphertext(plaintext.bits))
}

fn select_game_server(
  server: Server,
  client: Client,
) -> Result(#(Client, game_server_list.GameServer), error.Error) {
  use #(client, bits) <- result.try(client.read(
    client,
    select_game_server.length,
  ))
  let #(client, plaintext) = client.decrypt(client, bits)
  use packet <- result.try(select_game_server.decode(plaintext))

  log.info(
    inspect(server)
    <> ": got Select Game Server from client "
    <> client.inspect(client)
    <> ": "
    <> string.inspect(packet),
  )

  // TODO: can't assume there will always be a game server – remove this assert.
  let assert Ok(game_server) =
    list.drop(game_servers, up_to: packet.index) |> list.first

  Ok(#(client, game_server))
}

fn send_connect_to_game_server(
  server: Server,
  client: Client,
  game_server: game_server_list.GameServer,
) -> Result(Client, error.Error) {
  let new_key = crypto.strong_random_bytes(4) |> u.pack_bytes()
  let packet = connect_to_game_server.ConnectToGameServer(game_server, new_key)

  log.info(
    inspect(server)
    <> ": sending Connect To Game Server to client "
    <> client.inspect(client)
    <> ": "
    <> string.inspect(packet),
  )

  let plaintext = connect_to_game_server.encode(packet)
  client.write(client, cipher.Ciphertext(plaintext.bits))
}
