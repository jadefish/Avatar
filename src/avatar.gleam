import gleam/bit_array
import gleam/bytes_builder
import gleam/erlang/process
import gleam/int
import gleam/io
import gleam/option.{None}
import gleam/otp/actor
import gleam/result
import gleam/string
import glisten.{Packet, User}

const port = 4000

pub fn main() {
  let assert Ok(_) =
    glisten.handler(fn(_conn) { #(Nil, None) }, fn(msg, state, conn) {
      case msg {
        Packet(bit_array) -> {
          let assert Ok(info) = glisten.get_client_info(conn)
          let ip = glisten.ip_address_to_string(info.ip_address)
          let ip_and_port = ip <> ":" <> int.to_string(info.port)
          let bytes = bit_array.inspect(bit_array)
          let n = bit_array.byte_size(bit_array)
          let text =
            bit_array.to_string(bit_array)
            |> result.unwrap("(error)")
            |> string.trim()

          io.println(
            ip_and_port
            <> ": "
            <> text
            <> " ("
            <> int.to_string(n)
            <> " bytes)\n\t"
            <> bytes,
          )

          let assert Ok(_) =
            glisten.send(conn, bytes_builder.from_string("ok\n"))

          actor.continue(state)
        }
        User(_user_message) -> {
          todo
        }
      }
    })
    |> glisten.serve(port)

  io.println("Listening on port " <> int.to_string(port))

  process.sleep_forever()
}
