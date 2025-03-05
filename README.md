# CPP Judge

An online judge system for C++ programming problems, built with Go.

## Features

- Users can view and solve informatics/competitive programming problems
- C++ code submission and automatic evaluation
- Secure sandbox for code execution
- Problem management and test case definition
- User authentication and submission history

## Project Structure

```
cppjudge/
├── cmd/            # Command-line applications
├── internal/       # Internal packages
│   ├── api/        # API handlers and routes
│   ├── auth/       # Authentication logic
│   ├── config/     # Configuration management
│   ├── db/         # Database interactions
│   ├── judge/      # Code judging logic
│   ├── models/     # Data models
│   ├── sandbox/    # Secure code execution
│   └── utils/      # Utility functions
├── web/            # Web assets
│   ├── static/     # Static files (CSS, JS)
│   └── templates/  # HTML templates
├── go.mod          # Go module file
└── main.go         # Application entry point
```

## Getting Started

1. Clone the repository
2. Run `go run main.go` to start the server
3. Access the web interface at http://localhost:8080

## Requirements

- Go 1.22 or later
- GCC/G++ compiler for C++ compilation
