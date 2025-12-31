pub type Error {
  InvalidSeed
  UnsupportedVersion

  DecodeError
  EncodeError

  UnexpectedPacket
  IOError(IOError)

  AuthenticationError(AuthenticationError)
}

pub type IOError {
  ReadError
  WriteError
  CloseError
}

pub type AuthenticationError {
  InvalidCredentials
  AccountInUse
  AccountBanned
}
