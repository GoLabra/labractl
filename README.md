# LabraGo CLI (`labractl`)

[![Go Version](https://img.shields.io/badge/Go-1.24%2B-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Version](https://img.shields.io/badge/Version-v1.0.3-orange.svg)](https://github.com/GoLabra/labractl/releases)

> **The official CLI tool for creating and running [LabraGo](https://github.com/GoLabra/labra) projects — a modern headless CMS built in Go.**

LabraGo CLI (`labractl`) streamlines the development workflow for LabraGo projects by automating project setup, dependency management, and development server orchestration.

## ✨ Features

### 🚀 **Project Creation**
- **One-command project setup**: `labractl create <project-name>`
- **Automatic repository cloning** from the official LabraGo template
- **Smart dependency detection** and installation prompts
- **Environment configuration** with pre-configured `.env` files
- **Database setup** with PostgreSQL user and database creation

### 🔧 **Development Workflow**
- **Concurrent server management**: Run backend and frontend simultaneously
- **Package manager flexibility**: Support for both npm and yarn
- **Automatic script generation**: Creates necessary start scripts
- **Debug mode**: Comprehensive logging for troubleshooting

### 🛡️ **Robust Tool Detection**
- **Comprehensive Go detection**: Verifies installation, executability, and functionality
- **Package manager detection**: Automatically detects npm/yarn availability
- **Database tool validation**: Ensures PostgreSQL tools are properly installed
- **Graceful error handling**: Clear warnings and installation guidance

## 📋 Requirements

### **Required Tools**
- **[Go](https://golang.org/doc/install)** (1.24+) - Backend development
- **[Node.js](https://nodejs.org/)** (18+) - Frontend development
- **[PostgreSQL](https://www.postgresql.org/download/)** - Database
- **[Git](https://git-scm.com/)** - Version control

### **Package Managers** (Choose One)
- **[Yarn](https://yarnpkg.com/)** (Recommended)
- **[npm](https://www.npmjs.com/)** (Alternative)

### **System Requirements**
- **macOS**: 10.15+ (with Homebrew recommended)
- **Linux**: Ubuntu 18.04+ / CentOS 7+
- **Windows**: Windows 10+ (with WSL recommended)

## 🚀 Quick Start

### **1. Install labractl**

```bash
# Using Go
go install github.com/GoLabra/labractl@latest

# Or build from source
git clone https://github.com/GoLabra/labractl.git
cd labractl
go build -o labractl main.go
```

Add the Go bin directory to your PATH so `labractl` is recognized (macOS/Linux):

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

For Windows (PowerShell):

```powershell
[Environment]::SetEnvironmentVariable("Path", ("$env:USERPROFILE\go\bin;" + $env:Path), "User")
```

### **2. Create a new project**

```bash
labractl create my-awesome-project
```

This command will:
- ✅ Clone the LabraGo repository
- ✅ Configure development environment
- ✅ Install all dependencies
- ✅ Set up PostgreSQL database
- ✅ Generate configuration files

### **3. Start development servers**

```bash
cd my-awesome-project
labractl start
```

This launches:
- 🖥️ **Backend server** on `http://localhost:4000`
- 🎨 **Frontend admin** on `http://localhost:3000`
- 🔍 **GraphQL Playground** on `http://localhost:4000/playground`

## 📖 Detailed Usage

### **Project Creation**

```bash
# Basic project creation
labractl create myproject

# With automatic yes to prompts
labractl create myproject --yes

# With debug logging
labractl create myproject --debug
```

**What happens during creation:**
1. **Repository cloning** from `https://github.com/GoLabra/labra`
2. **Go module configuration** with local API replacement
3. **Environment setup** with backend `.env` and frontend `.env.local` files
4. **Dependency installation** using your preferred package manager
5. **Database initialization** with PostgreSQL user and database
6. **Code generation** with `go mod tidy` and `go generate`

### **Development Server**

```bash
# Start both servers
labractl start

# With debug mode
labractl start --debug
```

**Server configuration:**
- **Backend**: Go server with GraphQL API
- **Frontend**: Next.js admin interface
- **Database**: PostgreSQL with automatic migrations
- **Real-time**: WebSocket subscriptions enabled

### **Available Commands**

```bash
labractl create <project-name>  # Create new project
labractl start                  # Start development servers
labractl version               # Show version information
labractl help                  # Show help information
```

## 🔧 Configuration

### **Environment Variables**

```bash
# Enable debug logging
export LABRA_DEBUG=1

# Or use flag
labractl create myproject --debug
```

### **Selecting Labra Version/Branch**

Use `--labra-ref` to choose a specific Labra tag or branch when creating a project:

```bash
# Use a release tag
labractl create myproject --labra-ref v1.2.3

# Or a branch name
labractl create myproject --labra-ref feature/new-auth
```

### **Package Manager Selection**

The CLI will prompt you to choose between npm and yarn:

```bash
📦 Choose package manager (npm/yarn) [default: yarn]: yarn
```

### **Database Configuration**

PostgreSQL is automatically configured with:
- **Host**: `localhost`
- **Port**: `5432`
- **User**: `postgres`
- **Password**: `postgres`
- **Database**: `<project-name>`

## 🛠️ Development

### **Building from Source**

```bash
# Clone repository
git clone https://github.com/GoLabra/labractl.git
cd labractl

# Install dependencies
go mod tidy

# Build binary
go build -o labractl main.go

# Test the build
./labractl version
```

### **Running Tests**

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...
```

### **Local Development**

```bash
# Build and run
go run main.go create test-project

# With debug
go run main.go create test-project --debug
```

## 🔍 Troubleshooting

### **Common Issues**

#### **Go Not Detected**
```bash
# Install Go
# macOS
brew install go

# Ubuntu
sudo apt install golang-go

# Windows
# Download from https://golang.org/dl/
```

#### **PostgreSQL Issues**
```bash
# macOS
brew install postgresql
brew services start postgresql

# Ubuntu
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
```

#### **Package Manager Issues**
```bash
# Install Yarn
npm install -g yarn

# Or use npm
labractl create myproject
# Choose 'npm' when prompted
```

### **Debug Mode**

Enable detailed logging to troubleshoot issues:

```bash
# Set environment variable
export LABRA_DEBUG=1

# Or use flag
labractl create myproject --debug
```

### **Manual Database Setup**

If automatic database setup fails:

```bash
# Create PostgreSQL user
createuser -s postgres

# Create database
createdb -U postgres myproject
```

## 🤝 Contributing

We welcome contributions! Here's how to get started:

### **Development Setup**

1. **Fork the repository**
2. **Clone your fork**
   ```bash
   git clone https://github.com/your-username/labractl.git
   cd labractl
   ```
3. **Create a feature branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```
4. **Make your changes**
5. **Test thoroughly**
   ```bash
   go test ./...
   go build -o labractl main.go
   ./labractl version
   ```
6. **Commit with conventional commits**
   ```bash
   git commit -m "feat: add amazing new feature"
   ```
7. **Push and create a Pull Request**

### **Code Style**

- Follow Go conventions and `gofmt`
- Add tests for new features
- Update documentation as needed
- Use conventional commit messages

### **Reporting Issues**

When reporting issues, please include:
- **Operating system** and version
- **Go version** (`go version`)
- **labractl version** (`labractl version`)
- **Debug output** (`labractl create test --debug`)
- **Steps to reproduce**

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI framework
- Inspired by modern development workflows
- Community-driven development

---

**Made with ❤️ by the GoLabra team**

For more information about LabraGo, visit [https://github.com/GoLabra/labra](https://github.com/GoLabra/labra)
