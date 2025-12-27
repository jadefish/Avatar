import gleam/function
import utils.{env_var_enabled}

const esc = "\u{001b}"

const text_black = esc <> "[30m"

const text_red = esc <> "[31m"

const text_green = esc <> "[32m"

const text_yellow = esc <> "[33m"

const text_blue = esc <> "[33m"

const text_magenta = esc <> "[35m"

const text_cyan = esc <> "[36m"

const text_white = esc <> "[37m"

const text_reset = esc <> "[0m"

const text_bold = esc <> "[1m"

fn with_escape_sequence(
  sequence: String,
  callback: fn(fn(String) -> String) -> String,
) -> String {
  case env_var_enabled("NO_COLOR") || env_var_enabled("NO_COLOUR") {
    True -> callback(function.identity)
    False ->
      callback(fn(format: String) -> String { sequence <> format <> text_reset })
  }
}

pub fn black(string: String) -> String {
  use formatter <- with_escape_sequence(text_black)
  formatter(string)
}

pub fn red(string: String) -> String {
  use formatter <- with_escape_sequence(text_red)
  formatter(string)
}

pub fn green(string: String) -> String {
  use formatter <- with_escape_sequence(text_green)
  formatter(string)
}

pub fn yellow(string: String) -> String {
  use formatter <- with_escape_sequence(text_yellow)
  formatter(string)
}

pub fn blue(string: String) -> String {
  use formatter <- with_escape_sequence(text_blue)
  formatter(string)
}

pub fn magenta(string: String) -> String {
  use formatter <- with_escape_sequence(text_magenta)
  formatter(string)
}

pub fn cyan(string: String) -> String {
  use formatter <- with_escape_sequence(text_cyan)
  formatter(string)
}

pub fn white(string: String) -> String {
  use formatter <- with_escape_sequence(text_white)
  formatter(string)
}

pub fn bold(string: String) -> String {
  use formatter <- with_escape_sequence(text_bold)
  formatter(string)
}

pub fn reset(string: String) -> String {
  text_reset <> string
}
