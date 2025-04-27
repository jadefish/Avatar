import gleam/result
import gleam/int
import aio.{type Closer, type Reader, type Writer}
import error
import gleam/bit_array
import gleam/bytes_tree
import glisten
import glisten/socket
import glisten/tcp

pub fn reader(socket: socket.Socket, timeout timeout: Int) -> Reader {
  case timeout {
    n if n <= 0 -> {
      // Size is ignored – always read all available data.
      fn(_n) {
        case tcp.receive_timeout(socket, 0, timeout) {
          Ok(data) -> Ok(data)
          Error(_) -> Error(error.IOError(error.ReadError))
        }
      }
    }

    _ -> {
      fn(_n) {
        // Size is ignored – always read all available data.
        case tcp.receive(socket, 0) {
          Ok(data) -> Ok(data)
          Error(_) -> Error(error.IOError(error.ReadError))
        }
      }
    }
  }
}

pub fn writer(socket: socket.Socket) -> Writer {
  fn(bits) {
    let bytes = bytes_tree.from_bit_array(bits)
    let n = bit_array.byte_size(bits)

    case tcp.send(socket, bytes) {
      // TODO: gen_tcp:send doesn't indicate how much data was written. Is it
      // safe here to assume all data was written?
      Ok(_) -> Ok(n)
      Error(_) -> Error(error.IOError(error.WriteError))
    }
  }
}

pub fn closer(socket: socket.Socket) -> Closer {
  fn() {
    case tcp.close(socket) {
      Ok(_) -> Ok(Nil)
      Error(_) -> Error(error.IOError(error.CloseError))
    }
  }
}

pub type SocketType(a) {
  Client(glisten.Connection(a))
  Server(glisten.Server)
}

/// Returns a string address for a glisten socket (client or server) in the
/// format `ip:port`.
///
/// ## Examples
/// ```gleam
/// socket_addr(Client(client_socket.conn))
/// // "127.0.0.1:51484"
///
/// socket_addr(Server(some_server))
/// // "127.0.0.1:7775"
/// ```
pub fn socket_addr(derp: SocketType(a)) {
  let result = case derp {
    Client(conn) -> glisten.get_client_info(conn)
    Server(server) ->
      glisten.get_server_info(server, 500) |> result.replace_error(Nil)
  }

  case result {
    Ok(connection_info) -> {
      let glisten.ConnectionInfo(port, ip) = connection_info
      glisten.ip_address_to_string(ip) <> ":" <> int.to_string(port)
    }

    Error(_) -> "(unknown)"
  }
}
