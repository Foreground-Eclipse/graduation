# Graduation
## Run Locally

Edit the config.env in graduation/config

create a postgres database in container(with the data you used in config)
```bash
docker run --name YOURNAME-p YOURPORT:YOURPORT -e POSTGRES_USER=YOURPOSTGRESUSER -e POSTGRES_PASSWORD=YOURPOSTGRESPASSWORD -e POSTGRES_DB=YOURPOSTGRESDB-d postgres
```

Build with docker

```bash
docker compose up --build
```

Start the server locally

download 2nd service
```bash
https://github.com/Foreground-Eclipse/grpcexchanger
```
run first service
```bash
go run cmd/server/main.go
```
run second service
```bash
go run cmd/server/main.go
```


Or just run it in docker)


## API Reference


#### Register

```http
  POST /api/v1/register
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `username`      | `string` | **Required**. username |
| `password`      | `string DEPOSIT or WITHDRAW` | **Required**. password |
| `email`      | `int` | **Required**. email |

#### Login

```http
  POST /api/v1/login
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `username`      | `string` | **Required**. username |
| `password`      | `string DEPOSIT or WITHDRAW` | **Required**. password |

#### Getting balance

```http
  GET /api/v1/balance
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `JWT Token in header`      | `string` | **Required**.  Auth method|

#### Deposit

```http
  POST /api/v1/wallet/deposit
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `amount`      | `float64` | **Required**. amount of how much to deposit|
| `currency`      | `string USD or RUB or EUR` | **Required**. currency to deposit|

#### Withdraw

```http
  POST /api/v1/wallet/withdraw
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `JWT Token`      | `Header` | **Required**. JWT Auth token|
| `anount`      | `float64` | **Required**. amount to withdraw|
| `currency`      | `string USD or RUB or EUR` | **Required**. currency to withdraw|

#### Get exchange rates

```http
  GET /api/v1/exchange/rates
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `JWT Token`      | `string` | **Required**. JWT Auth token|

#### Exchange currency

```http
  POST /api/v1/exchange
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `from_currency`      | `string` | **Required**. from what currency to exchange|
| `to_currency`      | `string` | **Required**. to what currency to exchange|
| `amount`      | `float64` | **Required**. how much to exchange|






