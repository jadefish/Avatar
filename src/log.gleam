import logging.{Alert, Critical, Debug, Emergency, Error, Info, Notice, Warning}

// Just a re-export so callers don't need to import both modules.
pub fn configure() -> Nil {
  logging.configure()
  // TODO: use an env var for this:
  // logging.set_level(Debug)
}

pub fn emergency(string: String) -> Nil {
  logging.log(Emergency, string)
}

pub fn alert(string: String) -> Nil {
  logging.log(Alert, string)
}

pub fn critical(string: String) -> Nil {
  logging.log(Critical, string)
}

pub fn error(string: String) -> Nil {
  logging.log(Error, string)
}

pub fn warning(string: String) -> Nil {
  logging.log(Warning, string)
}

pub fn notice(string: String) -> Nil {
  logging.log(Notice, string)
}

pub fn info(string: String) -> Nil {
  logging.log(Info, string)
}

pub fn debug(string: String) -> Nil {
  logging.log(Debug, string)
}
