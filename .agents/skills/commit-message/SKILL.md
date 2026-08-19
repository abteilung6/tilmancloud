---
name: commit-message
description: >-
  Writes git commit messages for this repo: component name in the title plus a
  short human-readable body. Use when committing, staging a commit, amending a
  message, or when the user asks for a commit message.
---

# Commit messages

```
component: imperative summary

Why this change exists, in one or two plain sentences.
```

## Title

- **Component only** — `kind`, `clickhouse`, `api`, `operator`, `skills`. Not `feat`, `fix`, `chore`
- The area of the system, not a filename
- Imperative, present tense (`add`, not `added`)
- ≤72 chars, no trailing period, no emoji or ticket junk unless the user gave an id

## Body

- Explain **why**, not which files moved
- Readable by a human who did not write the patch
- One or two sentences. No bullet dump, no AI/co-author footers unless asked
- Pass the message with a HEREDOC

## Examples

```
kind: add cluster-up and cluster-down

Create a local four-node Kind cluster from Make so the lab can be recreated without installing tools in the Makefile.
```

```
makefile: stop downloading CLIs from cluster-up

Make only manages the cluster; kind and kubectl are installed by the agent or the human.
```

Bad: `update files` · `feat(kind): add cluster` · a body that lists every path.
