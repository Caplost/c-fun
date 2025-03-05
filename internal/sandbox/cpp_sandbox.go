package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ExecutionResult represents the result of a code execution
type ExecutionResult struct {
	Status      string
	Output      string
	ErrorOutput string
	RunTime     int
	ExitCode    int
}

// CppSandbox handles safe execution of C++ code
type CppSandbox struct {
	TempDir       string
	TimeLimit     int
	MemoryLimit   int
	CompilerFlags []string
	CompilerPath  string
}

// NewCppSandbox creates a new C++ sandbox
func NewCppSandbox() (*CppSandbox, error) {
	tempDir, err := ioutil.TempDir("", "cppjudge-sandbox-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Default configuration
	return &CppSandbox{
		TempDir:       tempDir,
		TimeLimit:     2000,   // 2 seconds
		MemoryLimit:   256000, // 256 MB
		CompilerFlags: []string{"-std=c++17", "-O2", "-Wall"},
		CompilerPath:  "g++",
	}, nil
}

// Cleanup removes temporary files
func (s *CppSandbox) Cleanup() error {
	return os.RemoveAll(s.TempDir)
}

// Execute compiles and runs C++ code with the given input
func (s *CppSandbox) Execute(code, input string) (ExecutionResult, error) {
	result := ExecutionResult{}

	// Create unique filenames for this execution
	id := fmt.Sprintf("%d", time.Now().UnixNano())
	sourceFile := filepath.Join(s.TempDir, fmt.Sprintf("solution_%s.cpp", id))
	executableFile := filepath.Join(s.TempDir, fmt.Sprintf("solution_%s.exe", id))
	inputFile := filepath.Join(s.TempDir, fmt.Sprintf("input_%s.txt", id))

	// Write the source code to file
	if err := ioutil.WriteFile(sourceFile, []byte(code), 0644); err != nil {
		return result, fmt.Errorf("failed to write source file: %w", err)
	}

	// Write the input to file
	if err := ioutil.WriteFile(inputFile, []byte(input), 0644); err != nil {
		return result, fmt.Errorf("failed to write input file: %w", err)
	}

	// Compile the code
	compileResult, err := s.compile(sourceFile, executableFile)
	if err != nil {
		result.Status = "Compilation Error"
		result.ErrorOutput = compileResult
		return result, nil
	}

	// Run the code
	return s.run(executableFile, inputFile)
}

// compile compiles the C++ source code
func (s *CppSandbox) compile(sourceFile, executableFile string) (string, error) {
	args := append(s.CompilerFlags, "-o", executableFile, sourceFile)

	cmd := exec.Command(s.CompilerPath, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return stderr.String(), fmt.Errorf("compilation failed: %w", err)
	}

	return "", nil
}

// run executes the compiled binary with the provided input
func (s *CppSandbox) run(executableFile, inputFile string) (ExecutionResult, error) {
	result := ExecutionResult{}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.TimeLimit)*time.Millisecond)
	defer cancel()

	// Read input file
	input, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return result, fmt.Errorf("failed to read input file: %w", err)
	}

	// Create command
	cmd := exec.CommandContext(ctx, executableFile)
	cmd.Stdin = bytes.NewBuffer(input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set resource limits if possible
	// Note: This is platform-dependent and may require additional libraries
	// For production, consider using cgroups, Docker, or other containerization

	// Start timer
	startTime := time.Now()

	// Run the command
	err = cmd.Run()

	// Calculate execution time
	runTime := time.Since(startTime)
	result.RunTime = int(runTime.Milliseconds())

	// Check for timeout
	if ctx.Err() == context.DeadlineExceeded {
		result.Status = "Time Limit Exceeded"
		return result, nil
	}

	// Check for runtime errors
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Status = "Runtime Error"
			result.ErrorOutput = stderr.String()
			return result, nil
		}
		return result, err
	}

	// Success
	result.Status = "Success"
	result.Output = stdout.String()
	result.ExitCode = 0

	return result, nil
}

// CompareOutput compares the expected output with actual output
func CompareOutput(expected, actual string) bool {
	// Normalize line endings and whitespace
	expected = normalizeOutput(expected)
	actual = normalizeOutput(actual)

	return expected == actual
}

// normalizeOutput removes trailing whitespace, normalizes line endings, and trims
func normalizeOutput(output string) string {
	// Replace Windows line endings with Unix line endings
	output = strings.ReplaceAll(output, "\r\n", "\n")

	// Split into lines
	lines := strings.Split(output, "\n")

	// Trim whitespace from each line
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}

	// Join lines back together
	output = strings.Join(lines, "\n")

	// Trim trailing newlines
	output = strings.TrimSpace(output)

	return output
}
