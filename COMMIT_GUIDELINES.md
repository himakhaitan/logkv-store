# Commit Guidelines
Consistent and descriptive commit messages make collaboration smoother, changelogs clearer, and development faster.
This project follows the **[Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)** standard.

## Why This Matters

- **Automation:** Enables semantic-release, changelogs, and versioning tools to do their job.  
- **Clarity:** Makes the history easy to scan and understand.  
- **Consistency:** Every contributor, every repo — same format.  
- **Professionalism:** Maintains a clean commit history that speaks for itself.

## Commit Messages
Use clear, descriptive commit messages in the following format:

```
<type>(<scope>): <short description>
```

### Types:

| Type         | Description                                                              |
| ------------ | ------------------------------------------------------------------------ |
| **feat**     | A new feature or functionality                                           |
| **fix**      | A bug fix                                                                |
| **docs**     | Documentation-only changes                                               |
| **style**    | Code style, formatting, missing semicolons, etc. (no code logic changes) |
| **refactor** | Code restructuring without changing behavior                             |
| **perf**     | Performance improvements                                                 |
| **test**     | Adding or updating tests                                                 |
| **build**    | Changes to build tools, CI/CD, dependencies                              |
| **chore**    | Maintenance tasks like dependency bumps or housekeeping                  |
| **ci**       | Continuous Integration / Deployment related changes                      |
| **revert**   | Revert a previous commit                                                 |

### Scope
The scope describes the section of the codebase affected by the change.

#### Examples:
```
feat(auth): add password reset
fix(api): handle null responses correctly
docs(readme): clarify setup instructions
```

## Emperative Rules
- Use imperative mood — “add” not “added” or “adds.”
- Keep it short (max ~100 characters)
- No capitalization or punctuation at the end.
- Describe what and why, not how.

#### Bad
```bash
fixed bug
```

#### Good
```bash
fix(api): prevent crash when token is missing
```

## Commit Body (Optional)
The body provides more context. It’s optional but recommended for non-trivial commits.

Guidelines:
- Explain the motivation behind the change.
- Mention any limitations or related follow-up work.
- Wrap lines at ~100 characters.

Example:
```bash
feat(storage): migrate from localStorage to IndexedDB

The new implementation improves performance for large datasets.
Also added a migration utility for legacy data.
```

## Breaking Changes
If your change breaks backward compatibility, mention it in the footer starting with `BREAKING CHANGE:`.

Example:
```
feat(auth): switch JWT library

BREAKING CHANGE: tokens are no longer compatible with previous versions.
```

## Examples of Good Commits
```bash
feat(api): add rate limiting to endpoints
fix(auth): prevent login crash on empty password
docs(readme): update local setup steps
refactor(core): simplify middleware chain
chore(deps): bump <package> to 1.6.0
```

## Pro Tips
- Squash trivial commits (like fix typo) into meaningful ones.
- Use GitHub Draft PRs for ongoing work.
- Don’t commit generated files or large binaries unless necessary.

## Getting Help
Stuck? Confused? Broke something spectacularly? 😅
Don’t worry, we’ve all been there.

Hop into our GradrX Discord Community, where we hang out, debug, and build cool stuff together: 
[Join the GradrX Community Discord Server](https://discord.gg/GMvrvjtvPr)

You can ask for help in the `#contribution-help` or `#dev-help` channels — we’ll be there to assist! 💬