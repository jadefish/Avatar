import glisten/socket

pub type Error {
  InvalidSeed
  UnsupportedVersion

  DecodeError
  EncodeError

  UnexpectedPacket
  ReadError(socket.SocketReason)
  WriteError(socket.SocketReason)

  AuthenticationError(AuthenticationError)
}

pub type AuthenticationError {
  InvalidCredentals
  AccountInUse
  AccountBanned
}
