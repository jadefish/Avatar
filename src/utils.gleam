import envoy
import gleam/int
import gleam/result

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

/// An implementation of `result.unwrap` that works with `use`.
///
/// ## Examples
/// ```gleam
/// use foo <- try_unwrap(Ok(4), 42)
/// // Ok(4)
///
/// use bar <- try_unwrap(Error(-1), 42)
/// // 42
/// ```
pub fn try_unwrap(
  result: Result(a, e),
  or default: f,
  apply fun: fn(a) -> f,
) -> f {
  case result {
    Ok(v) -> fun(v)
    Error(_) -> default
  }
}

/// Pack the provided byte-aligned bit array into a big-endian integer.
pub fn pack_bytes(bits: BitArray) -> Int {
  pack_bytes_loop(0, bits)
}

fn pack_bytes_loop(acc: Int, bits: BitArray) -> Int {
  case bits {
    <<next:int, rest:bytes>> ->
      int.bitwise_shift_left(acc, 8)
      |> int.bitwise_or(next)
      |> pack_bytes_loop(rest)
    _ -> acc
  }
}

/// Determines whether an environment variable has been enabled.
///
/// Unset environment variables and those set to "false" or the empty string are
/// considered to not be enabled.
/// Environment variables set to any other value are considered enabled.
pub fn env_var_enabled(name: String) -> Bool {
  case envoy.get(name) {
    Error(_) | Ok("") | Ok("false") -> False
    _ -> True
  }
}
