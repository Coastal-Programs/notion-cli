# Gemini Project Context: Notion CLI

This document provides a comprehensive overview of the `@coastal-programs/notion-cli` project for AI agent interactions.

## 1. Project Overview

This is a sophisticated, non-interactive command-line interface (CLI) for the Notion API, specifically designed for use by AI agents and in automation scripts. It provides a robust and reliable way to interact with Notion without the need for a graphical user interface.

- **Purpose**: To offer a fast, script-friendly, and AI-oriented interface to the Notion API.
- **Technologies**:
  - **Language**: TypeScript
  - **Framework**: [oclif](https://oclif.io/) v3 (a framework for building CLIs in Node.js)
  - **Node.js Version**: >=18.0.0
- **Key Architectural Patterns**:
  - **Command-Based Structure**: Functionality is organized into commands and subcommands (e.g., `db query`, `page create`), located in `src/commands/`.
  - **Standardized Response Envelope**: A custom `BaseCommand` class (`src/base-command.ts`) ensures all JSON responses are wrapped in a consistent `{success, data, metadata}` envelope, which simplifies parsing for AI agents.
  - **Robust Error Handling**: Errors are standardized into a `NotionCLIError` format and delivered in the same JSON envelope, providing predictable error responses.
  - **Smart ID Resolution**: A utility at `src/utils/notion-resolver.ts` automatically handles various Notion ID formats (e.g., URLs, database_ids, data_source_ids).
  - **Caching Layer**: The CLI implements both in-memory and persistent workspace caching (`src/cache.ts`) to improve performance and reduce API calls.

## 2. Building and Running

The project uses `npm` for package management and scripts.

- **Install Dependencies**:
  ```bash
  npm install
  ```
- **Build the Project**: Compiles TypeScript source code from `src/` to JavaScript in `dist/`.
  ```bash
  npm run build
  ```
- **Run Tests**: Executes the test suite using Mocha and Chai. Test files are located in the `test/` directory.
  ```bash
  npm test
  ```
- **Lint the Code**: Checks for code style and quality issues using ESLint.
  ```bash
  npm run lint
  ```
- **Execute Commands Locally**: After building, you can run the CLI using the provided run scripts.
  ```bash
  # Example: Run the 'whoami' command
  ./bin/run whoami
  ```

## 3. Development Conventions

Adhering to these conventions is crucial when modifying the codebase.

- **Framework**: All commands are built using the oclif framework. To create a new command, add a file in the `src/commands/` directory.
- **Base Command**: All new commands should extend the `BaseCommand` class from `src/base-command.ts`. This provides automatic JSON envelope formatting and consistent error handling.
- **Output Handling**:
  - For JSON output (`--json` or `--compact-json`), use the `this.outputSuccess(data, flags)` method. The base class handles the formatting and printing.
  - For non-JSON output (e.g., tables, markdown), the command is responsible for generating and printing the output directly using `this.log()` or `ux.table()`.
- **Error Handling**:
  - Use `try...catch` blocks to capture potential errors.
  - Inside the `catch` block, call `this.outputError(error, flags)` to ensure errors are formatted correctly for both JSON and human-readable output.
- **File Structure**:
  - Source code is in `src/`.
  - Corresponding tests are in `test/`, mirroring the source directory structure.
  - Reusable helper functions are located in `src/helper.ts` and `src/utils/`.
- **Testing**:
  - Tests are written with Mocha and Chai.
  - New features or bug fixes should be accompanied by corresponding tests in the `test/` directory.
- **Code Style**: The project uses ESLint and Prettier for code formatting. Run `npm run lint` to check your changes against the project's style guidelines.
