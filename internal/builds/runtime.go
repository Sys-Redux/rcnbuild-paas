package builds

import (
	"context"
	"fmt"

	"github.com/Sys-Redux/rcnbuild-paas/internal/github"
)

// Represents the detected application runtime
type Runtime string

const (
    RuntimeNodeJS  Runtime = "nodejs"
    RuntimePython  Runtime = "python"
    RuntimeGo      Runtime = "go"
    RuntimeStatic  Runtime = "static"
    RuntimeDocker  Runtime = "docker"
    RuntimeUnknown Runtime = "unknown"
)

// Represents detected runtime information and suggested commands
type RuntimeInfo struct {
    Runtime      Runtime `json:"runtime"`
    BuildCommand string  `json:"build_command"`
    StartCommand string  `json:"start_command"`
    Port         int     `json:"port"`
}

// Analyzes a repository to determine its runtime
func DetectRuntime(ctx context.Context, client *github.Client, owner, repo,
	branch, rootDir string) (*RuntimeInfo, error) {
    // Path to check (empty string = root)
    checkPath := rootDir
    if checkPath == "." || checkPath == "" {
        checkPath = ""
    }

    // Check for Dockerfile first (highest priority - user has custom build)
    if exists, _ := client.FileExists(ctx, owner, repo, joinPath(checkPath,
		"Dockerfile"), branch); exists {
        return &RuntimeInfo{
            Runtime:      RuntimeDocker,
            BuildCommand: "",
            StartCommand: "",
            Port:         3000,
        }, nil
    }

    // Check for Node.js (package.json)
    if exists, _ := client.FileExists(ctx, owner, repo, joinPath(checkPath,
		"package.json"), branch); exists {
        return detectNodeJSRuntime(ctx, client, owner, repo, branch, checkPath)
    }

    // Check for Python
    if exists, _ := client.FileExists(ctx, owner, repo, joinPath(checkPath,
		"requirements.txt"), branch); exists {
        return &RuntimeInfo{
            Runtime:      RuntimePython,
            BuildCommand: "pip install -r requirements.txt",
            StartCommand: "python app.py",
            Port:         8000,
        }, nil
    }

    if exists, _ := client.FileExists(ctx, owner, repo, joinPath(checkPath,
		"pyproject.toml"), branch); exists {
        return &RuntimeInfo{
            Runtime:      RuntimePython,
            BuildCommand: "pip install .",
            StartCommand: "python -m app",
            Port:         8000,
        }, nil
    }

    if exists, _ := client.FileExists(ctx, owner, repo, joinPath(checkPath,
		"Pipfile"), branch); exists {
        return &RuntimeInfo{
            Runtime:      RuntimePython,
            BuildCommand: "pipenv install",
            StartCommand: "pipenv run python app.py",
            Port:         8000,
        }, nil
    }

    // Check for Go
    if exists, _ := client.FileExists(ctx, owner, repo, joinPath(checkPath,
		"go.mod"), branch); exists {
        return &RuntimeInfo{
            Runtime:      RuntimeGo,
            BuildCommand: "go build -o app .",
            StartCommand: "./app",
            Port:         8080,
        }, nil
    }

    // Check for static site (index.html)
    if exists, _ := client.FileExists(ctx, owner, repo, joinPath(checkPath,
		"index.html"), branch); exists {
        return &RuntimeInfo{
            Runtime:      RuntimeStatic,
            BuildCommand: "",
            StartCommand: "",
            Port:         80,
        }, nil
    }

    // Unknown runtime
    return &RuntimeInfo{
        Runtime:      RuntimeUnknown,
        BuildCommand: "",
        StartCommand: "",
        Port:         3000,
    }, nil
}

// Determines Node.js specifics (npm, yarn, pnpm, framework)
func detectNodeJSRuntime(ctx context.Context, client *github.Client, owner,
	repo, branch, checkPath string) (*RuntimeInfo, error) {
    info := &RuntimeInfo{
        Runtime: RuntimeNodeJS,
        Port:    3000,
    }

    // Determine package manager
    packageManager := "npm"
    runCmd := "npm run"

    if exists, _ := client.FileExists(ctx, owner, repo,
		joinPath(checkPath,"pnpm-lock.yaml"), branch); exists {
        packageManager = "pnpm"
        runCmd = "pnpm"
    } else if exists, _ := client.FileExists(ctx, owner, repo,
		joinPath(checkPath, "yarn.lock"), branch); exists {
        packageManager = "yarn"
        runCmd = "yarn"
    } else if exists, _ := client.FileExists(ctx, owner, repo,
		joinPath(checkPath, "bun.lockb"), branch); exists {
        packageManager = "bun"
        runCmd = "bun run"
    }

    // Check for Next.js
    if exists, _ := client.FileExists(ctx, owner, repo,
		joinPath(checkPath, "next.config.js"), branch); exists {
        info.BuildCommand = packageManager + " install && " + runCmd + " build"
        info.StartCommand = runCmd + " start"
        return info, nil
    }
    if exists, _ := client.FileExists(ctx, owner, repo,
		joinPath(checkPath, "next.config.mjs"), branch); exists {
        info.BuildCommand = packageManager + " install && " + runCmd + " build"
        info.StartCommand = runCmd + " start"
        return info, nil
    }
    if exists, _ := client.FileExists(ctx, owner, repo,
		joinPath(checkPath, "next.config.ts"), branch); exists {
        info.BuildCommand = packageManager + " install && " + runCmd + " build"
        info.StartCommand = runCmd + " start"
        return info, nil
    }

    // Check for Vite/static build
    if exists, _ := client.FileExists(ctx, owner, repo,
		joinPath(checkPath, "vite.config.js"), branch); exists {
        info.BuildCommand = packageManager + " install && " + runCmd + " build"
        info.StartCommand = runCmd + " preview"
        info.Port = 4173
        return info, nil
    }
    if exists, _ := client.FileExists(ctx, owner, repo,
		joinPath(checkPath, "vite.config.ts"), branch); exists {
        info.BuildCommand = packageManager + " install && " + runCmd + " build"
        info.StartCommand = runCmd + " preview"
        info.Port = 4173
        return info, nil
    }

    // Default Node.js
    info.BuildCommand = packageManager + " install"
    info.StartCommand = runCmd + " start"

    return info, nil
}

// Joins path components, handling empty root
func joinPath(base, file string) string {
    if base == "" {
        return file
    }
    return base + "/" + file
}

// Returns a generated Dockerfile for the runtime
func GetDockerfileForRuntime(info *RuntimeInfo, buildCmd,
	startCmd string) string {
    switch info.Runtime {
    case RuntimeNodeJS:
        return generateNodeJSDockerfile(buildCmd, startCmd, info.Port)
    case RuntimePython:
        return generatePythonDockerfile(buildCmd, startCmd, info.Port)
    case RuntimeGo:
        return generateGoDockerfile(buildCmd, info.Port)
    case RuntimeStatic:
        return generateStaticDockerfile()
    default:
        return ""
    }
}

func generateNodeJSDockerfile(buildCmd, startCmd string, port int) string {
    return `FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN ` + buildCmd + `

FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app .
EXPOSE ` + itoa(port) + `
CMD ["` + startCmd + `"]
`
}

func generatePythonDockerfile(buildCmd, startCmd string, port int) string {
    return `FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE ` + itoa(port) + `
CMD ["` + startCmd + `"]
`
}

func generateGoDockerfile(buildCmd string, port int) string {
    return `FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/app .
EXPOSE ` + itoa(port) + `
CMD ["./app"]
`
}

func generateStaticDockerfile() string {
    return `FROM nginx:alpine
COPY . /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
`
}

func itoa(i int) string {
    return fmt.Sprintf("%d", i)
}
