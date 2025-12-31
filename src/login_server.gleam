import ansi
import cipher.{type Cipher, type Ciphertext, type Plaintext}
import client.{type Client}
import error
import gleam/bit_array
import gleam/bool
import gleam/bytes_tree
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
import glisten/tcp
import log
import packets/connect_to_game_server
import packets/game_server_list
import packets/login_request
import packets/login_seed
import packets/select_game_server
import time_zone as tz
import utils

// TODO: need listener import for get_server_info, but it's internal.
import glisten/internal/listener

/// Login servers authenticate new client connections and facilitate relaying
/// known clients to a game server.
pub opaque type Server {
  Server(
    parent: Subject(LoginResult),
    listener_name: process.Name(listener.Message),
    port: Int,
    pool_size: Int,
    sessions: dict.Dict(glisten.ConnectionInfo, Session),
  )
}

pub type LoginResult =
  Result(Client, error.Error)

@internal
pub type Action {
  Start
  Stop
}

type Session {
  // Handshake's buffer isn't Ciphertext or Plaintext because it may contain
  // both unencrypted and encrypted data. Clients can send multiple commands at
  // the same time, and Glisten reads all available data from the socket, so
  // the initial message received by the server may be the unencrypted Login
  // Seed and (all or part of) the encrypted Login Request.
  Handshake(buffer: BitArray)
  Encrypted(buffer: Ciphertext, cipher: Cipher, state: SessionState)
}

type SessionState {
  AwaitingCredentials
  AwaitingGameServerSelection
  ReadyToRelay
}

/// Create a new login server. Once started, it will accept and process
/// connections on the provided port. Newly-connected clients are authenticated.
/// Regardless of the outcome of the process, results are sent to the provided
/// subject.
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

/// Start the server, allowing it to accept new connections.
pub fn start(subject: Subject(Action)) -> Subject(Action) {
  actor.send(subject, Start)
  subject
}

/// Stop the server, preventing new connections from being accepted. Existing
/// connections are not terminated.
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

fn loop(server: Server, action: Action) -> actor.Next(Server, Action) {
  case action {
    Start -> {
      let new_connection = fn(conn) {
        log.info(
          inspect(server)
          <> ": new connection: "
          <> string.inspect(conn)
          <> " ------------------------------",
        )

        // let writer = tcp.writer(conn.socket)
        // let closer = tcp.closer(conn.socket)
        // let client = client.new(addr, writer, closer)
        let session = Handshake(<<>>)
        let assert Ok(connection_info) = glisten.get_client_info(conn)
        let sessions = dict.insert(server.sessions, connection_info, session)
        let new_server = Server(..server, sessions: sessions)
        #(new_server, None)
      }

      let result =
        glisten.new(new_connection, message_received)
        |> glisten.bind("0.0.0.0")
        |> glisten.with_pool_size(server.pool_size)
        |> glisten.with_close(fn(server) {
          // It'd be nice to have a reference to the closed connection in here.
          log.debug(inspect(server) <> ": connection closed.")
        })
        // TODO: start_with_listener_name is marked @internal, but it's used
        // in one of glisten's example programs. also, the listener_name is
        // eventually needed by the inspect function, above.
        |> glisten.start_with_listener_name(server.port, server.listener_name)

      case result {
        Ok(_started) -> {
          log.info(inspect(server) <> ": ready!")
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

    Stop -> actor.stop()
  }
}

const test_ip = #(127, 0, 0, 1)

// TODO: fetch from ... somehere.
const game_servers = [
  game_server_list.GameServer("US East", tz.AmericaDetroit, test_ip, 7080),
  game_server_list.GameServer("US West", tz.AmericaLosAngeles, test_ip, 7081),
]

const max_packet_size_in_bytes = 0xF000

fn get_session(
  server: Server,
  conn: glisten.Connection(_),
) -> Result(Session, Nil) {
  let assert Ok(conn_info) = glisten.get_client_info(conn)
  dict.get(server.sessions, conn_info)
}

fn put_session(
  server: Server,
  conn: glisten.Connection(_),
  session: Session,
) -> Server {
  let assert Ok(conn_info) = glisten.get_client_info(conn)
  let sessions = dict.insert(server.sessions, conn_info, session)
  Server(..server, sessions: sessions)
}

fn remove_session(server: Server, conn: glisten.Connection(_)) -> Server {
  let assert Ok(conn_info) = glisten.get_client_info(conn)
  let sessions = dict.delete(server.sessions, conn_info)
  Server(..server, sessions: sessions)
}

fn message_received(
  server0: Server,
  message: glisten.Message(a),
  conn: glisten.Connection(a),
) -> glisten.Next(Server, glisten.Message(a)) {
  let assert glisten.Packet(message) = message
  let n = bit_array.byte_size(message)

  use <- bool.lazy_guard(n > max_packet_size_in_bytes, fn() {
    log.notice(
      inspect(server0)
      <> ": "
      <> string.inspect(conn)
      <> " sent overlong packet ("
      <> int.to_string(n)
      <> " bytes)",
    )
    let _ = tcp.close(conn)
    glisten.continue(server0)
  })

  log.debug(inspect(server0) <> " ============== START MESSAGE ==============")

  use session0 <- utils.try_unwrap(
    get_session(server0, conn),
    glisten.continue(server0),
  )

  log.debug(
    inspect(server0)
    <> ": "
    <> int.to_string(bit_array.byte_size(message))
    <> " bytes from "
    <> string.inspect(session0)
    <> ":\n\t"
    <> bit_array.inspect(message),
  )

  let #(session1, effects) = handle_session(session0, message)

  log.debug(inspect(server0) <> ": session1 =\n\t" <> string.inspect(session1))
  log.debug(inspect(server0) <> ": effects =\n\t" <> string.inspect(effects))
  let server1 = put_session(server0, conn, session1)

  execute_effects(conn, effects)

  log.debug(inspect(server1) <> " =============== END MESSAGE ===============")

  glisten.continue(server1)
}

fn execute_effects(conn: glisten.Connection(a), effects0: List(Effect)) -> Nil {
  list.fold(effects0, conn, execute_effect)
  Nil
}

fn execute_effect(
  conn: glisten.Connection(a),
  effect: Effect,
) -> glisten.Connection(a) {
  case effect {
    CloseConnection -> {
      let _ = tcp.close(conn)
      conn
    }

    Write(bits) -> {
      let _ = tcp.send(conn.socket, bytes_tree.from_bit_array(bits))
      conn
    }
  }
}

fn handle_session(
  session: Session,
  message: BitArray,
) -> #(Session, List(Effect)) {
  case session {
    Handshake(buffer) -> handle_handshake(buffer, message)
    Encrypted(buffer, cipher, state) ->
      handle_encrypted_message(buffer, message, cipher, state)
  }
}

fn handle_handshake(
  buffer0: BitArray,
  message: BitArray,
) -> #(Session, List(Effect)) {
  let buffer1 = bit_array.append(buffer0, suffix: message)

  case parse_handshake(buffer1) {
    IncompleteHandshake -> #(Handshake(buffer1), []) |> echo

    MalformedHandshake -> #(Handshake(<<>>), [CloseConnection]) |> echo

    SuccessfulHandshake(login_seed, ciphertext) -> {
      let cipher = cipher.login(login_seed.seed, login_seed.version)
      let session = Encrypted(ciphertext, cipher, AwaitingCredentials)
      handle_session(session, <<>>) |> echo
    }
  }
}

fn parse_handshake(buffer0: BitArray) -> HandshakeOutcome {
  case buffer0 {
    <<data:bytes-size(21), rest:bits>> ->
      case login_seed.decode(cipher.Plaintext(data)) {
        Ok(login_seed) ->
          SuccessfulHandshake(login_seed, cipher.Ciphertext(rest))

        Error(_) -> MalformedHandshake
      }

    _ -> IncompleteHandshake
  }
}

type Effect {
  CloseConnection
  Write(data: BitArray)
}

fn handle_encrypted_message(
  ciphertext0: Ciphertext,
  message: BitArray,
  cipher0: Cipher,
  state0: SessionState,
) -> #(Session, List(Effect)) {
  let ciphertext1 =
    cipher.Ciphertext(bit_array.append(ciphertext0.bits, suffix: message))
  let #(ciphertext2, cipher1, commands) = parse_commands(ciphertext1, cipher0)
  let #(state2, effects) = process_commands(state0, commands)

  #(Encrypted(ciphertext2, cipher1, state2), effects)
}

// Maps commands to effects for later execution. Also handles session state
// transitions.
fn process_commands(
  state: SessionState,
  commands: List(ParserOutcome),
) -> #(SessionState, List(Effect)) {
  list.fold(commands, #(state, []), process_command)
}

fn process_command(
  tuple: #(SessionState, List(Effect)),
  command: ParserOutcome,
) -> #(SessionState, List(Effect)) {
  let #(state, effects) = tuple

  echo #(state, command)

  case state, command {
    _, MalformedCommand -> #(state, [CloseConnection])

    AwaitingCredentials, Handle(LoginRequest(account, password, next_key)) -> {
      // TODO: process login request
      log.debug("awaiting credentials: got login request")

      let gsl =
        game_server_list.GameServerList(
          game_servers,
          game_server_list.SendSystemInfo,
        )
      let bits = game_server_list.encode(gsl).bits
      #(AwaitingGameServerSelection, [Write(bits), ..effects])
    }

    AwaitingGameServerSelection, Handle(SpyOnClient(data)) -> {
      log.debug(
        "awaiting game server selection: got spy on client:\n\t"
        <> string.inspect(data),
      )
      // TODO: do something with data?
      #(state, effects)
    }

    AwaitingGameServerSelection, Handle(SelectGameServer(index)) -> {
      // TODO
      log.debug("awaiting game server selection: got select game server")

      // TODO: remove this assert – index may be out-of-bounds.
      let assert Ok(game_server) =
        list.drop(game_servers, up_to: index) |> list.first
      let new_key = crypto.strong_random_bytes(4) |> utils.pack_bytes()
      let ctgs =
        connect_to_game_server.ConnectToGameServer(game_server, new_key)
      let bits = connect_to_game_server.encode(ctgs).bits

      #(ReadyToRelay, [Write(bits), ..effects])
    }

    ReadyToRelay, _ -> #(state, [CloseConnection])

    _, _ -> #(state, [CloseConnection, ..effects])
  }
}

// TODO: move incomplete and malformed into error.Error and change to
// Result(Handshake, error.Error)?
type HandshakeOutcome {
  SuccessfulHandshake(login_seed.LoginSeed, Ciphertext)
  IncompleteHandshake
  MalformedHandshake
}

type ParserOutcome {
  MalformedCommand
  Handle(InboundCommand)
}

fn parse_commands(
  ciphertext0: Ciphertext,
  cipher0: Cipher,
) -> #(Ciphertext, Cipher, List(ParserOutcome)) {
  case next_command(ciphertext0, cipher0) {
    // Partial data (not enough for one full command) – return input and wait
    // for more.
    Incomplete -> #(ciphertext0, cipher0, [])

    // Something went wrong during parsing, so just indicate that the connection
    // should be closed.
    Malformed -> #(cipher.Ciphertext(<<>>), cipher0, [MalformedCommand])

    // Client sent at least one full command. It may have also sent a subsequent
    // partial or whole command, so try to parse another:
    Complete(plaintext, ciphertext1, cipher1) -> {
      log.debug(
        "parse_commands: complete!\n\tplaintext: "
        <> string.inspect(plaintext)
        <> "\n\tciphertext1: "
        <> string.inspect(ciphertext1),
      )
      let #(ciphertext2, cipher2, actions) =
        parse_commands(ciphertext1, cipher1)
      #(ciphertext2, cipher2, [Handle(plaintext), ..actions])
    }
  }
}

type ParsedCommand {
  Incomplete
  Malformed
  Complete(command: InboundCommand, rest: Ciphertext, new_cipher: Cipher)
}

fn next_command(ciphertext0: Ciphertext, cipher0: Cipher) -> ParsedCommand {
  // Decrypt the first byte of a message, attempting to determine its ID.
  // Regardless of if this operation succeeded, the original cipher (cipher0) is
  // returned so the entire command (including the ID byte) can later be
  // decrypted.
  case decrypt_id(ciphertext0, using: cipher0) {
    InsufficientData -> {
      log.debug("next_command: incomplete!")
      Incomplete
    }

    Decrypted(cipher.Plaintext(<<id:8>>), _n, _cipher1) -> {
      log.debug("next_command: decrypted ID: 0x" <> int.to_base16(id))

      case get_command_spec(id) {
        Ok(Fixed(_name, length, decoder)) ->
          decrypt_fixed_length_command(
            ciphertext0,
            length,
            using: cipher0,
            decoder: decoder,
          )

        Ok(Variable(_name, decoder)) ->
          decrypt_variable_length_command(
            ciphertext0,
            using: cipher0,
            decoder: decoder,
          )

        Error(_) -> {
          log.debug("next_command: unknown command: 0x" <> int.to_base16(id))
          Malformed
        }
      }
    }

    _ -> {
      log.debug("next_command: malformed!")
      Malformed
    }
  }
}

type DecryptResult {
  // TODO: n isn't really needed. Callers can just pass a sliced bit array, and
  // cipher always decrypts all bytes, so n ≡ |plaintext|.
  Decrypted(plaintext: Plaintext, n: Int, cipher: Cipher)
  InsufficientData
  DecryptError
}

fn decrypt_prefix(
  cipher0: Cipher,
  ciphertext: Ciphertext,
  length: Int,
) -> DecryptResult {
  log.debug(
    "decrypt_prefix: "
    <> int.to_string(length)
    <> " bytes from:\n\t"
    <> string.inspect(ciphertext),
  )
  case ciphertext.bits {
    _ if length <= 0 -> DecryptError

    // |ciphertext| ≥ length > 0
    <<data:bytes-size(length), _:bytes>> -> {
      let #(cipher1, plaintext) =
        cipher.decrypt(cipher0, cipher.Ciphertext(data))
      Decrypted(plaintext, length, cipher1)
    }

    _ -> InsufficientData
  }
}

fn decrypt_id(
  from ciphertext: Ciphertext,
  using cipher: Cipher,
) -> DecryptResult {
  decrypt_prefix(cipher, ciphertext, 1)
}

fn decrypt_fixed_length_command(
  ciphertext0: Ciphertext,
  length: Int,
  using cipher0: Cipher,
  decoder decoder: CommandDecoder,
) -> ParsedCommand {
  // Although decrypt_prefix is being called, the "prefix" is the entire length
  // of the command.
  case decrypt_prefix(cipher0, ciphertext0, length) {
    InsufficientData -> {
      log.debug(
        "decrypt_fixed_length_command: incomplete!\n\t"
        <> string.inspect(ciphertext0),
      )
      Incomplete
    }

    DecryptError -> {
      log.debug(
        "decrypt_fixed_length_command: malformed (1)!\n\t"
        <> string.inspect(ciphertext0),
      )
      Malformed
    }

    Decrypted(plaintext, n, cipher1) -> {
      case ciphertext0.bits, plaintext.bits {
        <<_:bytes-size(n), rest:bytes>>, <<data:bytes-size(n)>> -> {
          let plaintext = cipher.Plaintext(data)
          let ciphertext1 = cipher.Ciphertext(rest)
          log.debug(
            "decrypt_fixed_length_command: complete!\n\tplaintext: "
            <> string.inspect(plaintext)
            <> "\n\tciphertext1: "
            <> string.inspect(ciphertext1),
          )

          case decoder(plaintext) {
            Ok(command) -> Complete(command, ciphertext1, cipher1)
            Error(_) -> Malformed
          }
        }

        _, _ -> {
          log.debug(
            "decrypt_fixed_length_command: malformed (2)!\n\tciphertext0: "
            <> string.inspect(ciphertext0)
            <> "\n\t",
          )
          Malformed
        }
      }
    }
  }
}

// <<id:8, length:16>>
const variable_command_header_size = 3

fn decrypt_variable_length_command(
  ciphertext0: Ciphertext,
  using cipher0: Cipher,
  decoder decoder: CommandDecoder,
) -> ParsedCommand {
  case decrypt_prefix(cipher0, ciphertext0, variable_command_header_size) {
    InsufficientData -> {
      log.debug(
        "decrypt_varible_length_command: incomplete!\n\t"
        <> string.inspect(ciphertext0),
      )
      Incomplete
    }

    DecryptError -> {
      log.debug(
        "decrypt_varible_length_command: malformed (1)!\n\t"
        <> string.inspect(ciphertext0),
      )
      Malformed
    }

    Decrypted(plaintext, _n, _cipher1) -> {
      case plaintext.bits {
        <<_id:8, length:16>> -> {
          case decrypt_prefix(cipher0, ciphertext0, length) {
            InsufficientData -> Incomplete

            Decrypted(plaintext, n, cipher1) -> {
              case ciphertext0.bits, plaintext.bits {
                <<_:bytes-size(n), rest:bytes>>, <<data:bytes-size(n)>> -> {
                  let plaintext = cipher.Plaintext(data)
                  let ciphertext1 = cipher.Ciphertext(rest)
                  log.debug(
                    "decrypt_variable_length_command: complete!\n\tplaintext: "
                    <> string.inspect(plaintext)
                    <> "\n\tciphertext1: "
                    <> string.inspect(ciphertext1),
                  )

                  case decoder(plaintext) {
                    Ok(command) -> Complete(command, ciphertext1, cipher1)
                    Error(_) -> Malformed
                  }
                }

                _, _ -> {
                  log.debug(
                    "decrypt_variable_length_command: malformed (2)!\n\tciphertext0: "
                    <> string.inspect(ciphertext0)
                    <> "\n\t",
                  )
                  Malformed
                }
              }
            }

            _ -> Malformed
          }
        }

        _ -> {
          log.debug(
            "decrypt_varible_length_command: malformed 2)!\n\tciphertext0: "
            <> string.inspect(ciphertext0)
            <> "\n\tplaintext: "
            <> string.inspect(plaintext),
          )
          Malformed
        }
      }
    }
  }
}

type InboundCommand {
  LoginRequest(account: String, password: String, next_key: Int)
  SpyOnClient(data: BitArray)
  SelectGameServer(index: Int)
}

type CommandDecoder =
  fn(Plaintext) -> Result(InboundCommand, error.Error)

type OutboundCommand {
  GameServerList(game_server: List(game_server_list.GameServer))
  ConnectToGameServer(index: Int)
}

type CommandEncoder =
  fn() -> BitArray

type CommandSpec {
  Fixed(name: String, length: Int, decoder: CommandDecoder)
  Variable(name: String, decoder: CommandDecoder)
}

// Only needs to know about encrypted commands, post-handshake, so Login Seed
// isn't included.
fn get_command_spec(command_id: Int) -> Result(CommandSpec, Nil) {
  case command_id {
    0x80 ->
      Ok(
        Fixed("Login Request", 62, fn(plaintext) {
          login_request.decode(plaintext)
          |> result.map(fn(lr) {
            LoginRequest(lr.account, lr.password, lr.next_key)
          })
        }),
      )

    0xD9 -> Ok(Fixed("Spy on Client", 199, fn(_plaintext) { todo }))

    0xA0 ->
      Ok(
        Fixed("Select Game Server", 3, fn(plaintext) {
          select_game_server.decode(plaintext)
          |> result.map(fn(sgs) { SelectGameServer(index: sgs.index) })
        }),
      )

    _ -> Error(Nil)
  }
}
// fn handle_login_seed(
//   server: Server,
//   client: Client,
//   data: BitArray,
// ) -> Result(Client, error.Error) {
//   use login_seed <- result.try(login_seed.decode(data))
//   let cipher = cipher.login(login_seed.seed, login_seed.version)

//   log.info(
//     inspect(server)
//     <> ": got Login Seed from client "
//     <> client.inspect(client)
//     <> ": "
//     <> string.inspect(login_seed),
//   )

//   Ok(client.with_cipher(client, cipher: cipher))
// }

// fn handle_login_request(
//   server: Server,
//   client: Client,
//   data: BitArray,
// ) -> Result(Client, error.Error) {
//   use login_request <- result.try(login_request.decode(data))

//   // TODO: mask password in this output
//   log.info(
//     inspect(server)
//     <> ": got Login Request from client "
//     <> client.inspect(client)
//     <> ": "
//     <> login_request.inspect(login_request),
//   )

//   // TODO: Authenticate the client:
//   // 1. credential match
//   // 2. ban check
//   // 3. account-in-use check

//   Ok(client)
// }

// fn deny_login(
//   server: Server,
//   client: Client,
//   reason: login_denied.Reason,
// ) -> Result(Client, error.Error) {
//   let packet = login_denied.LoginDenied(reason)
//   let plaintext = login_denied.encode(packet)

//   log.info(
//     inspect(server)
//     <> ": sending Login Denied to client "
//     <> client.inspect(client)
//     <> ": "
//     <> string.inspect(packet),
//   )

//   client.write(client, Ciphertext(plaintext.bits))
//   // No need to close the connection here – client closes its end.
// }

// fn send_game_server_list(
//   server: Server,
//   client: Client,
//   game_servers: List(game_server_list.GameServer),
// ) {
//   let game_server_list =
//     game_server_list.GameServerList(
//       game_servers,
//       game_server_list.DoNotSendSystemInfo,
//     )

//   log.info(
//     inspect(server)
//     <> ": sending Game Server List to client "
//     <> client.inspect(client)
//     <> ": "
//     <> string.inspect(game_server_list),
//   )

//   let plaintext = game_server_list.encode(game_server_list)
//   client.write(client, Ciphertext(plaintext.bits))
// }

// fn select_game_server(
//   server: Server,
//   client: Client,
// ) -> Result(#(Client, game_server_list.GameServer), error.Error) {
//   use #(client, bits) <- result.try(client.read(
//     client,
//     select_game_server.length,
//   ))
//   let #(client, plaintext) = client.decrypt(client, length: 3)
//   use packet <- result.try(select_game_server.decode(plaintext))

//   log.info(
//     inspect(server)
//     <> ": got Select Game Server from client "
//     <> client.inspect(client)
//     <> ": "
//     <> string.inspect(packet),
//   )

//   // TODO: can't assume there will always be a game server – remove this assert.
//   let assert Ok(game_server) =
//     list.drop(game_servers, up_to: packet.index) |> list.first

//   Ok(#(client, game_server))
// }

// fn send_connect_to_game_server(
//   server: Server,
//   client: Client,
//   game_server: game_server_list.GameServer,
// ) -> Result(Client, error.Error) {
//   let new_key = crypto.strong_random_bytes(4) |> u.pack_bytes()
//   let packet = connect_to_game_server.ConnectToGameServer(game_server, new_key)

//   log.info(
//     inspect(server)
//     <> ": sending Connect To Game Server to client "
//     <> client.inspect(client)
//     <> ": "
//     <> string.inspect(packet),
//   )

//   let plaintext = connect_to_game_server.encode(packet)
//   client.write(client, Ciphertext(plaintext.bits))
// }
