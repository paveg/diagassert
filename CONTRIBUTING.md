# Contributing to diagassert

First off, thank you for considering contributing to diagassert! It's people
like you that make diagassert such a great tool.

## How Can I Contribute?

### Reporting Bugs

This is one of the simplest and most effective ways to contribute. If you find a
bug, please ensure the bug was not already reported by searching on GitHub under
[Issues](https://github.com/paveg/diagassert/issues).

If you're unable to find an open issue addressing the problem, [open a new
one](https://github.com/paveg/diagassert/issues/new). Be sure to include a
**title and clear description**, as much relevant information as possible, and a
**code sample** or an **executable test case** demonstrating the expected
behavior that is not occurring.

### Suggesting Enhancements

If you have an idea for a new feature or an improvement to an existing one,
please open an issue with the "enhancement" label. Clearly describe the proposed
enhancement, why it's needed, and how it would work.

### Pull Requests

We love pull requests! If you're planning to contribute code, please follow
these steps:

1. **Fork the repository** and create your branch from `main`.
2. **Set up your development environment**. You'll need Go and Node.js (for
   `npx`) installed. Install the development tools by running:

    ```sh
    make install-tools
    ```

3. **Install Git hooks** by running:

    ```sh
    make install-hooks
    ```

4. **Make your changes**.
5. **Add tests** for your changes. We take testing seriously.
6. **Ensure the test suite passes**. When you commit your changes, `lefthook`
   will automatically run formatters and linters.
7. **Create a pull request** to the `main` branch of the `paveg/diagassert`
   repository. Please provide a clear description of the problem and solution.
   Include the relevant issue number if applicable.

## Styleguides

### Git Commit Messages

- Use the present tense ("Add feature" not "Added feature").
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...").
- Limit the first line to 72 characters or less.
- Reference issues and pull requests liberally after the first line.

### Go Styleguide

We follow the standard Go community style guidelines. Use `go fmt` and
`goimports` to format your code. The `make fmt` command will do this for you.

## Any other questions?

Feel free to open an issue and we'll be happy to help you.

Thank you for your contribution!
