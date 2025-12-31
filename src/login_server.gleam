import ansi
import cipher.{type Cipher, type Ciphertext, type Plaintext}
import client.{type Client}
import error
import gleam/bit_array
import gleam/dict
import gleam/erlang/process.{type Subject}
import gleam/int
import gleam/option.{None}
import gleam/otp/actor
import gleam/string
import glisten
import log
import packets/game_server_list
import packets/login_seed
import time_zone as tz

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
  Encrypted(buffer: Ciphertext, cipher: Cipher)
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
        |> glisten.with_close(fn(_server) {
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

fn message_received(
  server: Server,
  message: glisten.Message(a),
  conn: glisten.Connection(a),
) -> glisten.Next(Server, glisten.Message(a)) {
  let assert glisten.Packet(message) = message

  // TODO: remove these asserts:
  log.debug(inspect(server) <> " ============== START MESSAGE ==============")
  let assert Ok(conn_info) = glisten.get_client_info(conn)
  let assert Ok(session0) = dict.get(server.sessions, conn_info)

  log.debug(
    inspect(server)
    <> ": "
    <> int.to_string(bit_array.byte_size(message))
    <> " bytes from "
    <> string.inspect(conn_info)
    <> ":\n\t"
    <> bit_array.inspect(message),
  )

  let #(session1, actions) = handle_session(session0, message)

  log.debug(inspect(server) <> ": session1 =\n\t" <> string.inspect(session1))
  log.debug(inspect(server) <> ": actions =\n\t" <> string.inspect(actions))

  let sessions = dict.insert(server.sessions, conn_info, session1)
  let server = Server(..server, sessions: sessions)

  log.debug(inspect(server) <> " =============== END MESSAGE ===============")

  glisten.continue(server)
}

fn handle_session(
  session: Session,
  message: BitArray,
) -> #(Session, List(NextStep)) {
  case session {
    Handshake(buffer) -> handle_handshake(buffer, message)
    Encrypted(buffer, cipher) ->
      handle_encrypted_message(buffer, message, cipher)
  }
}

fn handle_handshake(
  buffer0: BitArray,
  message: BitArray,
) -> #(Session, List(NextStep)) {
  let buffer1 = bit_array.append(buffer0, suffix: message)

  case parse_handshake(buffer1) {
    HIncomplete -> #(Handshake(buffer1), [])

    HMalformed -> #(Handshake(<<>>), [CloseConnection])

    HSuccess(login_seed, remaining_encrypted_data) -> {
      let cipher = cipher.login(login_seed.seed, login_seed.version)
      let session0 = Encrypted(remaining_encrypted_data, cipher)
      handle_session(session0, <<>>)
    }
  }
}

fn parse_handshake(buffer0: BitArray) -> HandshakeOutcome {
  case buffer0 {
    <<data:bytes-size(21), rest:bits>> ->
      case login_seed.decode(cipher.Plaintext(data)) {
        Ok(login_seed) -> HSuccess(login_seed, cipher.Ciphertext(rest))

        Error(_) -> HMalformed
      }

    _ -> HIncomplete
  }
}

fn handle_encrypted_message(
  ciphertext0: Ciphertext,
  message: BitArray,
  cipher0: Cipher,
) -> #(Session, List(NextStep)) {
  let ciphertext1 =
    cipher.Ciphertext(bit_array.append(ciphertext0.bits, suffix: message))
  let #(ciphertext2, cipher1, actions) = parse_commands(ciphertext1, cipher0)

  #(Encrypted(ciphertext2, cipher1), actions)
}

type HandshakeOutcome {
  HSuccess(login_seed.LoginSeed, Ciphertext)
  HIncomplete
  HMalformed
}

type NextStep {
  CloseConnection
  Handle(data: Plaintext)
}

fn parse_commands(
  ciphertext0: Ciphertext,
  cipher0: Cipher,
) -> #(Ciphertext, Cipher, List(NextStep)) {
  case next_command(ciphertext0, cipher0) {
    // Partial data (not enough for one full command) – return input and wait
    // for more.
    Incomplete -> #(ciphertext0, cipher0, [])

    // Something went wrong during parsing, so just indicate that the connection
    // should be closed.
    Malformed -> #(cipher.Ciphertext(<<>>), cipher0, [CloseConnection])

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
        parse_commands(ciphertext1, cipher1 |> echo)
      echo cipher2
      #(ciphertext2, cipher2, [Handle(plaintext), ..actions])
    }
  }
}

type ParsedCommand {
  Incomplete
  Malformed
  Complete(data: Plaintext, rest: Ciphertext, new_cipher: Cipher)
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

    Decrypted(cipher.Plaintext(<<id:8>>), _n, cipher1) -> {
      echo cipher1
      log.debug("next_command: decrypted ID: 0x" <> int.to_base16(id))

      case get_command_spec(id) {
        Ok(Fixed(_name, length)) ->
          decrypt_fixed_length_command(ciphertext0, length, using: cipher0)

        Ok(Variable(_name)) ->
          decrypt_variable_length_command(ciphertext0, using: cipher0)

        Error(_) -> todo
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
  // cipher always decrypt all bytes, so n ≡ |plaintext|.
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
        cipher.decrypt(cipher0 |> echo, cipher.Ciphertext(data))
      Decrypted(plaintext, length, cipher1 |> echo)
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

          Complete(plaintext, ciphertext1, cipher1 |> echo)
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

                  Complete(plaintext, ciphertext1, cipher1)
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

type CommandSpec {
  Fixed(name: String, length: Int)
  Variable(name: String)
}

fn get_command_spec(command_id: Int) -> Result(CommandSpec, Nil) {
  case command_id {
    0xEF -> Ok(Fixed("Login Seed", 21))
    0x80 -> Ok(Fixed("Login Request", 62))
    0xD9 -> Ok(Fixed("Spy on Client", 199))
    0xA0 -> Ok(Fixed("Select Game Server", 3))
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
