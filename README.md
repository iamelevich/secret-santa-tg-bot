# Secret Santa Bot

Simple bot to create secret santa in your company.

## Configuration

Create `.env` file in root of this repo. You can do it by:

```bash
cp .env.example .env
```

### Env variables

| Variable                          | Example value                                      |
| --------------------------------- | -------------------------------------------------- |
| `TELEGRAM_BOT_API_TOKEN`          | `telegram_bot_token`                               |
| `MONGODB_URL`                     | `mongodb://db_username:dp_password@mongodb:27017/` |
| `MONGODB_DATABASE`                | `dbname`                                           |
| `MONGO_INITDB_ROOT_USERNAME`      | `db_username`                                      |
| `MONGO_INITDB_ROOT_PASSWORD`      | `dp_password`                                      |
| `ME_CONFIG_MONGODB_ADMINUSERNAME` | `db_username_for_gui`                              |
| `ME_CONFIG_MONGODB_ADMINPASSWORD` | `dp_password_for_gui`                              |
| `ME_CONFIG_BASICAUTH_USERNAME`    | `db_username`                                      |
| `ME_CONFIG_BASICAUTH_PASSWORD`    | `dp_password`                                      |
| `ME_CONFIG_MONGODB_URL`           | `mongodb://db_username:dp_password@mongodb:27017/` |

### Makefile

| Command       | Description                         |
| ------------- | ----------------------------------- |
| `make vendor` | Just run `go mod vendor`            |
| `make build`  | Build binry with help of `go build` |

## Running

### Development

First you need to load all dependencies:

```bash
make vendor
```

Next - start database

```bash
docker compose -f docker-compose.dev.yml up -d
```

And last step - run you bot:

```bash
go run ./cmd/bot.go
```

### Production

First step - build image:

```bash
docker compose -f docker-compose.yml build
```

Next step run:

```bash
docker compose -f docker-compose.yml run -d
```