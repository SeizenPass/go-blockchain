## Installation
```bash
// For installing all dependencies
go mod tidy

go install ./cmd/aitu/...
```

## CMD commands

| Command | Flags     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `aitu version`      | - | Describes version of blockchain |
| `aitu wallet new-account`      | **REQUIRED**<br/>`--datadir` | Creates a new account with a new set of elliptic-curve Private + Public keys |
| `aitu balances list`      | **REQUIRED**<br/>`--datadir` | Lists all balances |
| `aitu run`      | **REQUIRED**<br/>`--datadir`<br/>Optional<br/>`--ip`<br/>`--miner`<br/>`--port`<br/>`--bootstrap-ip`<br/>`--bootstrap-port`<br/>`--bootstrap-account`| Launches the AITU node and its HTTP API |

## Launching AITU node locally

```bash
aitu run --datadir=.
```

## API Reference

#### Get all balances

```http
GET /balances/list
```

#### Add transaction

```http
POST /tx/add
```
| Parameter | Type     | Description                |
| :-------- | :------- | :------------------------- |
| `from` | `string` | **Required**. Address(0x...) from which coins are sent |
| `from_pwd` | `string` | **Required**. Password of the account from whom coins will be sent |
| `to` | `string` | **Required**. Address(0x...) to which coins are sent |
| `value` | `uint` | **Required**. Number of coins sent |

#### Get node status

```http
GET /node/status
```