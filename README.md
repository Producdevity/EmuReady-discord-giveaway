# EmuReady Discord Giveaway

Discord bot service for running an EmuReady giveaway tied to GitHub repository stars. Users enter through a Discord slash command, complete GitHub OAuth, and receive the persistent Giveaway Pings role after the service verifies that their GitHub account has starred the configured repository.

The service is written in Go, uses PostgreSQL for entrant storage, and ships with Docker Compose for local development and single-server deployment.

## Local Development

```sh
docker compose up --build -d
curl http://localhost:8080/health
curl http://localhost:8080/ready
```

The default local Compose file uses development placeholders so the service can boot without real Discord or GitHub credentials. Real command and OAuth flows require values from a Discord application and a GitHub OAuth app.

## Checks

```sh
make lint
go test ./...
go test -race ./...
go vet ./...
```
