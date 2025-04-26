import gleam/int
import gleam/result
import glisten

/// Remove all NUL bytes from the provided bit array.
pub fn trim_nul(bits: BitArray) -> BitArray {
  trim_nul_loop(bits, <<>>)
}

fn trim_nul_loop(bits: BitArray, acc: BitArray) {
  case bits {
    <<>> -> acc
    <<x, rest:bits>> -> {
      case x {
        0x00 -> trim_nul_loop(rest, acc)
        _ -> trim_nul_loop(rest, <<acc:bits, x>>)
      }
    }
    _ -> acc
  }
}

/// Shorthand for `result.try(result |> result.replace_error(error))`.
///
/// ## Examples
/// ```gleam
/// type SomeError {
///   Kaboom
/// }
///
/// use foo <- try_replace(Ok(4), Kaboom)
/// // Ok(4)
///
/// use bar <- try_replace(Error(-1), Kaboom)
/// // Error(Kaboom)
/// ```
pub fn try_replace(
  result: Result(a, e),
  with error: f,
  apply fun: fn(a) -> Result(c, f),
) -> Result(c, f) {
  result.try(result |> result.replace_error(error), fun)
}

/// Shorthand for `result.try(result |> result.map_error(mapping))`.
///
/// ## Examples
/// ```gleam
/// type SomeError(a) {
///   Kaboom(a)
/// }
///
/// use foo <- try_map(Ok(4), Kaboom)
/// // Ok(4)
///
/// use bar <- try_map(Error(-1), Kaboom)
/// // Error(Kaboom(-1))
/// ```
pub fn try_map(
  result: Result(a, e),
  with mapping: fn(e) -> f,
  apply fun: fn(a) -> Result(c, f),
) -> Result(c, f) {
  result.try(result |> result.map_error(mapping), fun)
}

/// Returns a string address for a glisten connection (client or server) in the
/// format `ip:port`.
///
/// ## Examples
/// ```gleam
/// connection_addr(glisten.get_client_info(some_conn))
/// // "127.0.0.1:51484"
///
/// connection_addr(glisten.get_server_info(some_server))
/// // "127.0.0.1:7775"
/// ```
pub fn connection_addr(result: Result(glisten.ConnectionInfo, e)) {
  case result {
    Ok(glisten.ConnectionInfo(port, ip)) ->
      glisten.ip_address_to_string(ip) <> ":" <> int.to_string(port)

    Error(_) -> "(unknown)"
  }
}

// Pack the provided byte-aligned bit array into a big-endian integer.
pub fn pack_bytes(bits: BitArray) -> Int {
  pack_bytes_loop(0, bits).1
}

fn pack_bytes_loop(acc: Int, bits: BitArray) -> #(BitArray, Int) {
  case bits {
    <<next:int, rest:bytes>> ->
      int.bitwise_shift_left(acc, 8)
      |> int.bitwise_or(next)
      |> pack_bytes_loop(rest)
    _ -> #(<<>>, acc)
  }
}
