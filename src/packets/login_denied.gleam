import cipher

/// `0x82` Login Denied. Length 2, unencrypted.
///
/// Sent by a login server to indicate that the client's attempt to log in has
/// been denied due to the specified reason.
pub type LoginDenied {
  LoginDenied(reason: Reason)
}

/// The specific reason the login process was denied by a login server.
pub type Reason {
  /// A generic rejection when no more-specific reason is available.
  /// The message displayed to the user suggests a variety of potential
  /// solutions, including checking the caps lock key, updating the client,
  /// contacting customer support, etc.
  GenericDenial

  /// The user's account is in use.
  AccountInUse

  /// The user's account has been banned.
  AccountBanned

  /// The user's provided credentials are not valid.
  InvalidCredentials

  /// A communication or connectivity problem has occurred.
  CommunicationProblem

  /// "The IGR concurrency limit has been met."
  /// Supposedly a relic of an unfinished Origin system to support internet
  /// gaming rooms (net cafes).
  IGRConcurrencyLimitMet

  /// "The IGR time limit has been met."
  /// Supposedly a relic of an unfinished Origin system to support internet
  /// gaming rooms (net cafes).
  IGRTimeLimitMet

  /// "A general IGR authentication failure has occurred."
  /// Supposedly a relic of an unfinished Origin system to support internet
  /// gaming rooms (net cafes).
  GeneralIGRAuthenticationFailure
}

pub fn encode(packet: LoginDenied) -> cipher.PlainText {
  let reason_byte = encode_reason(packet.reason)
  cipher.PlainText(<<0x82, reason_byte:8>>)
}

fn encode_reason(reason: Reason) -> Int {
  case reason {
    GenericDenial -> 0x00
    AccountInUse -> 0x01
    AccountBanned -> 0x02
    InvalidCredentials -> 0x03
    CommunicationProblem -> 0x04
    IGRConcurrencyLimitMet -> 0x05
    IGRTimeLimitMet -> 0x06
    GeneralIGRAuthenticationFailure -> 0x07
  }
}
