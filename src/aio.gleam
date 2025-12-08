// This module should be named "io", but Gleam reserves this name.

import error

pub type Reader =
  fn(Int) -> Result(BitArray, error.Error)

pub type Writer =
  fn(BitArray) -> Result(Int, error.Error)

pub type Closer =
  fn() -> Result(Nil, error.Error)
