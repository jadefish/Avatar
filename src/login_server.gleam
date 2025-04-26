import cipher
import client.{type Client}
import error
import gleam/bit_array
import gleam/crypto
import gleam/erlang/process.{type Subject}
import gleam/io
import gleam/list
import gleam/option.{None, Some}
import gleam/otp/actor
import gleam/result
import gleam/string
import glisten
import glisten/tcp
import packets/connect_to_game_server
import packets/game_server_list
import packets/login_denied
import packets/login_request
import packets/login_seed
import packets/select_game_server
import time_zone as tz
import utils as u

/// A login server accepts a port and a parent subject. Once started, the server
/// authenticates connecting clients. Clients that have been successfully
/// authenticated are ready to be relayed to a game server.
pub opaque type Server {
  StoppedServer(parent: Subject(LoginResult), port: Int, pool_size: Int)
  StartedServer(parent: Subject(LoginResult), server: glisten.Server)
}

pub type LoginResult =
  Result(Client, error.Error)

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

fn handle_message(action: Action, server: Server) {
  case action {
    Start ->
      case server {
        // Starting an already-started server does nothing.
        StartedServer(_, _) -> actor.continue(server)

        StoppedServer(parent, port, pool_size) -> {
          let init = fn(conn) {
            let addr = u.connection_addr(glisten.get_client_info(conn))
            io.println("login_server: new connection: " <> addr)
            #(StoppedServer(parent, port, pool_size), None)
          }

          let result =
            glisten.handler(init, message_handler)
            // UO clients don't seem to do IPv6.
            |> glisten.bind("0.0.0.0")
            |> glisten.with_pool_size(pool_size)
            |> glisten.start_server(port)

          case result {
            Ok(server) -> {
              let addr = u.connection_addr(glisten.get_server_info(server, 500))
              io.println("login_server: listening on " <> addr)
              actor.continue(StartedServer(parent, server))
            }

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

  // TODO: Reconsider this design. Currently, there are two ways to read data
  // from a client: glisten's message handler (this function) and the
  // client.read function.
  //
  // Since login server packets are strictly ordered, a design wherein glisten
  // simply pushes read data into an inbox isn't ideal. It could work, if an
  // internal "authentication step" state variable is maintained for each
  // client, but both the current design and this feel awkward to use.
  //
  // Occasionally, a client will send the 0xD9 Spy On Client packet during
  // login, which implies that login server packets were meant to be handled
  // out-of-order, anyway.

  let client = client.Client(conn, None, bits, <<>>, cipher.nil())
  let result = case bits {
    <<0xEF, _:bits>> -> {
      // TODO: a client sending 0xD9 Spy On Client will break this process.
      use client <- result.try(handle_login_seed(client))
      use client <- result.try(handle_login_request(client))
      use client <- result.try(send_game_server_list(client, game_servers))
      // CLient may send 0xD9 before Select Game Server here.
      use #(client, game_server) <- result.try(handle_game_server_selection(
        client,
      ))
      use client <- result.try(send_connect_to_game_server(client, game_server))

      Ok(client)
    }

    bits -> {
      io.println("login_server: bad packet: " <> bit_array.inspect(bits))

      // It's fine if the connection couldn't be closed.
      let _ = tcp.close(conn)

      Error(error.UnexpectedPacket)
    }
  }

  case result {
    Ok(client) -> actor.send(server.parent, Ok(client))

    Error(error) ->
      case error {
        error.AuthenticationError(auth_error) -> {
          let _ = case auth_error {
            error.AccountBanned ->
              deny_login(client, login_denied.AccountBanned)
            error.AccountInUse -> deny_login(client, login_denied.AccountInUse)
            error.InvalidCredentals ->
              deny_login(client, login_denied.InvalidCredentials)
          }

          actor.send(server.parent, result)
        }

        _ -> actor.send(server.parent, result)
      }
  }

  actor.continue(server)
}

fn handle_login_seed(client: Client) -> Result(Client, error.Error) {
  // Receive 0xEF Login Seed (unencrypted, length 21):
  use #(client, bits) <- result.try(client.read(client, 21))
  let #(cipher, plaintext) = cipher.decrypt(client.cipher, bits)
  use login_seed <- result.try(login_seed.decode(plaintext))
  let cipher = cipher.login(login_seed.seed, login_seed.version)

  echo login_seed

  Ok(client.Client(..client, login_seed: Some(login_seed.seed), cipher:))
}

fn handle_login_request(client: Client) {
  // Receive 0x80 Login Request (encrypted, length 62):
  use #(client, bits) <- result.try(client.read(client, 62))
  let #(cipher, plaintext) = cipher.decrypt(client.cipher, bits)
  use login_request <- result.try(login_request.decode(plaintext))

  // TODO: Authenticate the client:
  // 1. credential match
  // 2. ban check
  // 3. account-in-use check

  // TODO: The password should be masked when printed here.
  echo login_request

  Ok(client.Client(..client, cipher:))
}

fn deny_login(
  client: Client,
  reason: login_denied.Reason,
) -> Result(Client, error.Error) {
  let packet = login_denied.LoginDenied(reason)
  let plaintext = login_denied.encode(packet)
  client.write(client, cipher.Ciphertext(plaintext.bits))
}

fn send_game_server_list(
  client: Client,
  game_servers: List(game_server_list.GameServer),
) {
  let game_server_list =
    game_server_list.GameServerList(
      game_servers,
      game_server_list.DoNotSendSystemInfo,
    )
  echo game_server_list

  let plaintext = game_server_list.encode(game_server_list)
  client.write(client, cipher.Ciphertext(plaintext.bits))
}

fn handle_game_server_selection(
  client: Client,
) -> Result(#(Client, game_server_list.GameServer), error.Error) {
  use #(client, bits) <- result.try(client.read(
    client,
    select_game_server.length,
  ))
  let #(cipher, plaintext) = cipher.decrypt(client.cipher, bits)
  use packet <- result.try(select_game_server.decode(plaintext))

  // TODO: remove assert
  let assert Ok(game_server) =
    list.drop(game_servers, up_to: packet.index) |> list.first

  echo packet

  Ok(#(client.Client(..client, cipher:), game_server))
}

fn send_connect_to_game_server(
  client: Client,
  game_server: game_server_list.GameServer,
) -> Result(Client, error.Error) {
  let new_key = crypto.strong_random_bytes(4) |> u.pack_bytes()
  let packet = connect_to_game_server.ConnectToGameServer(game_server, new_key)
  
  echo packet

  let plaintext = connect_to_game_server.encode(packet)
  client.write(client, cipher.Ciphertext(plaintext.bits))
}
