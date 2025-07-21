# LabraGo CLI (`labractl`)

This is the official CLI tool for creating and running [LabraGo](https://github.com/GoLabra/labra) projects — a headless CMS built in Go.

---

## 🚀 Features

- `labractl create <project>` — creates a new project by cloning the LabraGo repo
- Automatically patches `go.mod` for local development
- Generates `.env` files for backend (`src/app`) and frontend (`src/admin`)
- Installs dependencies: Go, Yarn, and PostgreSQL setup
- Ensures PostgreSQL user/database exist
- `labractl start` — runs both backend and frontend concurrently
- No configuration needed — it just works™

---

## 🧱 Requirements

- [Go](https://golang.org/doc/install)
- [Yarn](https://yarnpkg.com/)
- [PostgreSQL](https://www.postgresql.org/download/)  
  Ensure `psql`, `createuser`, and `createdb` are available in your terminal (in PATH).

---

## ⚙️ Usage

### 1. Create a new project

```bash
labractl create myproject
```

This will:

- Clone the LabraGo repo into `./myproject`
- Patch `go.mod` to use the local API
- Create `.env` files with correct defaults
- Run `go mod tidy` and `go generate ./...`
- Install frontend dependencies with Yarn
- Ensure PostgreSQL user and DB are in place

---

### 2. Start the project

```bash
cd myproject
labractl start
```

This will:

- Ensure `package.json` exists in root
- Add start scripts if missing
- Install `concurrently` if needed
- Start both backend and frontend:
    - Backend: `cd src/app && go run main.go start`
    - Frontend: `cd src/admin && yarn dev`

Verbose logs can be enabled with `--debug` or by setting `LABRA_DEBUG=1`.

---

## 🛠 Development

If you want to build or test the CLI locally:

```bash
go build -o labractl main.go
./labractl help
```

---

## 🤝 Contributing

PRs welcome. Open an issue or fork away if you want to improve the CLI.

---

## License

MIT © 2025 GoLabra
