# wcorefx

## Version Control

This project uses **Jujutsu (jj)** for version management (colocated with git).

### Common Workflow

```bash
# Start new feature branch
jj new

# Code changes go into working copy automatically...

# Commit changes (creates a commit and starts new empty working copy)
jj commit -m "Add DNS cache enumeration"
# Or amend changes into last commit
jj squash

# View log / status / diff
jj log
jj st
jj diff --git

# Sync with remote
jj git fetch
jj rebase -d main

# Push (bookmark must exist first)
jj bookmark create my-feature
jj git push -b my-feature
```

### Key Differences from Git

- No staging area — working directory IS a commit (`@`)
- Commits are mutable (edit, squash, absorb, rebase freely)
- Bookmarks are jj's equivalent of git branches, but DON'T auto-advance
- Before pushing, move bookmark: `jj bookmark move my-feature --to @`

### Colocated Repo

Both `.jj/` and `.git/` exist. jj and git commands can be used interchangeably, but prefer jj for daily work. Don't mix commands while the working copy has changes.

### Commit Messages

Use imperative verb phrase, sentence case, no full stop:
- `"Add user authentication to login endpoint"`
- `"Fix null pointer in payment processor"`
- `"Update dependencies to latest versions"`
