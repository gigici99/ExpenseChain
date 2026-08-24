# Contributing to ExpenseChain

Thank you for considering contributing! This project is a university prototype, so we keep the process simple.

## How to contribute

1. **Fork** the repository on GitHub.
2. **Clone** your fork locally:
   ```bash
   git clone https://github.com/yourusername/ExpenseChain.git
   cd ExpenseChain
   ```
3. Create a new branch for your change:
   ```bash
   git checkout -b my-feature
   ```
4. Make your changes.  For Go code run `go fmt ./...` and `go vet ./...`.  For the Vue frontend run `npm run lint`.
5. **Commit** with a clear message and push to your fork.
6. Open a **Pull Request** against the `main` branch of the original repository.

## Code style

- Go: `gofmt` and `go vet` must pass.
- Vue/JavaScript: use the provided ESLint configuration (`npm run lint`).
- Keep commit messages concise and descriptive.

## Testing

Automated tests are not present yet, but please make sure the application builds and runs locally before submitting a PR.

## License

All contributions are licensed under the same MIT license as the project.
