import aio.{type Closer, type Reader, type Writer}
import error
import gleam/bit_array
import gleam/bytes_tree
import gleam/int
import gleam/result
import glisten
import glisten/socket
import glisten/tcp

/// An `aio.Reader` that reads from a TCP socket.
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

/// An `aio.Writer` that writes to a TCP socket.
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

/// An `aio.Closer` that closes a TCP socket connection.
pub fn closer(socket: socket.Socket) -> Closer {
  fn() {
    case tcp.close(socket) {
      Ok(_) -> Ok(Nil)
      Error(_) -> Error(error.IOError(error.CloseError))
    }
  }
}

pub fn client_addr_string(conn: glisten.Connection(_)) -> String {
  glisten.get_client_info(conn)
  |> result.map(fn(connection_info) {
    let glisten.ConnectionInfo(port, ip) = connection_info
    glisten.ip_address_to_string(ip) <> ":" <> int.to_string(port)
  })
  |> result.unwrap("(unknown)")
}
