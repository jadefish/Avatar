# Avatar

Avatar is an emulator for Ultima Online: Renaissance.

## Running

```shell script
$ mkdir logs
$ login 2>>logs/login.log &
$ game 2>>logs/game.log &
```

## Client compatibility

TODO

## Building

**Requirements:**

* [goenv](https://github.com/syndbg/goenv)
* GNU Make

```shell script
$ goenv install
$ make
```

Find `login` and `game` in the `bin/` directory.

or:

```shell script
$ OUT_DIR=/some/where/else GOOS=windows GOARCH=amd64 make
```

Read the `Makefile` for more options.

## Configuration

Provide configuration via a `.env` file located in the runtime directory
or via values present in the runtime environment.

Get started by creating `.env`:
```shell script
$ cp .env.example ./bin/.env
```

Values provided in the runtime environment override values present in the
`.env` file.

### Connections

#### `login`

By default, `login` listens on `localhost:7775`. Change this address by
specifying a value for `LOGIN_ADDR`:

```.env
LOGIN_ADDR=10.1.2.3:55940
```

#### `game`

TODO

### Storage

Specify a storage provider by setting the value of the `STORAGE_PROVIDER`
environment variable.

* PostgreSQL: `postgres`
  * `DB_CONNECTION_STRING` must also be present.
* Memory: `memory` (not recommended)

```.env
STORAGE_PROVIDER=postgres
DB_CONNECTION_STRING="user=worker password=foobarbazbat"
```

### Passwords

Specify a password hashing algorithm by setting the value of the
`PASSWORD_CIPHER` environment variable.

* bcrypt: `bcrypt`
  * `BCRYPT_COST` (default: 10) may optionally be present.

You should not change the password hashing algorithm after creating accounts
as subsequent password verification would fail.

```.env
PASSWORD_CIPHER=bcrypt
BCRYPT_COST=7
```
