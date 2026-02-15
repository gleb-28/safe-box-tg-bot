---
apply: always
---

# Project Rules

## 1. General Principles
- Always prefer minimal, safe changes.
- Do not rewrite architecture unless explicitly requested.
- Preserve backward compatibility.
- Follow existing project patterns over "better" alternatives.
- Avoid introducing new dependencies unless necessary.

## 2. Code Style
- Follow existing formatting and naming conventions.
- Match project language idioms.
- Keep functions small and composable.
- Avoid premature optimization.
- Prefer readability over cleverness.

## 3. Architecture Rules
- Respect module boundaries.
- Do not move files or rename APIs without instruction.
- Reuse existing utilities instead of creating duplicates.
- Avoid global state.

## 4. Testing
- Run all tests after changes.
- Add tests for new functionality.
- Do not remove failing tests — fix code instead.
- Prefer unit tests over integration tests unless required.

## 5. Dependencies
- Do not install packages automatically.
- Justify any new dependency.
- Prefer standard library.

## 6. Security
- Never expose secrets.
- Validate all external input.
- Avoid unsafe operations.
- Follow least privilege principle.

## 7. Performance
- Do not optimize without evidence.
- Avoid unnecessary allocations.
- Consider algorithmic complexity.

## 8. When Unsure
- Ask for clarification.
- Choose the safest option.
- Explain tradeoffs.

## 9. Decision Policy
- NEVER guess requirements.
- NEVER hallucinate APIs.
- NEVER assume library behavior.
- ALWAYS inspect code before modifying.
- ALWAYS search project before implementing new code.

## 10. Logging
- Use existing logging framework.
- Do not introduce new logging libraries.
- Do not log secrets or sensitive data.
- Keep logs structured and actionable.
- Avoid excessive logging.

## 11. Documentation
- Update documentation when behavior changes.
- Document public APIs.
- Keep comments concise and useful.
- Do not explain obvious code.
- Prefer self-documenting code over comments.

## 12. Configuration
- Do not hardcode configuration values.
- Use existing config mechanisms.
- Provide sensible defaults.
- Avoid environment-specific assumptions.

## 13. Change Scope
- Modify only files necessary for the task.
- Do not refactor unrelated code.
- Do not reformat entire files.
- Avoid large sweeping changes.

## 14. Naming
- Use descriptive names.
- Follow project naming conventions.
- Do not introduce abbreviations without precedent.
- Prefer consistency over personal preference.

## 15. Go Specific Rules
- Follow gofmt formatting.
- Use `context.Context` where appropriate.
- Avoid global variables.
- Prefer composition over inheritance patterns.
- Avoid data races.
- Respect existing synchronization patterns.
- Do not introduce blocking operations in async flows.
- Use context cancellation properly.
- Close files and network connections.
- Avoid resource leaks.
- Respect timeouts and cancellation.
- Pass context as first argument.
- Do not store context in structs.
- Use idiomatic error handling patterns.
- Prefer explicit interfaces.

## 16. Error Handling
- Never ignore errors.
- Fail explicitly rather than silently.
- Provide meaningful error messages.
- Preserve existing error handling patterns.
- Do not introduce panic unless consistent with project.
- Wrap errors with context when appropriate.

## 16. Investigation First
- Understand the existing implementation before changing it.
- Read related files.
- Identify root cause before fixing.

## 17. Data Safety
- Never delete or overwrite data without explicit instruction.
- Prefer non-destructive changes.
- Validate data before processing.
- Preserve schema compatibility.

## 18. Refactoring
- Refactor only when required for the task.
- Preserve behavior during refactoring.
- Avoid speculative improvements.
- Prefer incremental changes.
