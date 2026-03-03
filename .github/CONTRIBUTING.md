# Contributing to Kage Package

## Reporting Issues

Did you find a bug or have a suggestion for improvement?

1. Check if the issue already exists by searching the [Issues](https://github.com/nanoninja/kage/issues) section.

2. If you can't find an existing issue, feel free to [open a new one](https://github.com/nanoninja/kage/issues/new). Please include:
   - A clear title and description
   - An example showing the unexpected behavior
   - The version of Go you're using
   - Any relevant code snippets

## Code Contributions

This package aims to remain simple and focused. Before submitting a pull request for a new feature, please open an issue first to discuss the proposed changes.

### Pull Request Process

1. Ensure your code follows the existing code style
2. Add or update tests as needed
3. Update documentation if necessary
4. Verify all tests pass locally
5. Sign your commits

### Development Setup

```bash
# Clone the repository
git clone https://github.com/nanoninja/kage.git

# Install dependencies
go mod download

# Run tests
go test -v ./...
```