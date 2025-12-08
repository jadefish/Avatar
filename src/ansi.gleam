// TODO: add support for NO_COLOR/NO_COLOUR

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

pub fn black(string: String) -> String {
  text_black <> string <> text_reset
}

pub fn red(string: String) -> String {
  text_red <> string <> text_reset
}

pub fn green(string: String) -> String {
  text_green <> string <> text_reset
}

pub fn yellow(string: String) -> String {
  text_yellow <> string <> text_reset
}

pub fn blue(string: String) -> String {
  text_blue <> string <> text_reset
}

pub fn magenta(string: String) -> String {
  text_magenta <> string <> text_reset
}

pub fn cyan(string: String) -> String {
  text_cyan <> string <> text_reset
}

pub fn white(string: String) -> String {
  text_white <> string <> text_reset
}

pub fn bold(string: String) -> String {
  text_bold <> string <> text_reset
}

pub fn reset(string: String) -> String {
  text_reset <> string
}
