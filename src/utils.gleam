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
