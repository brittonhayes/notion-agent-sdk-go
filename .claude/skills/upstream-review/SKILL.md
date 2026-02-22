---
name: upstream-review
description: Review recent changes to the Notion Agents SDK JS upstream and report feature parity gaps against our Go SDK.
user-invocable: true
---

# Upstream Review

Review recent changes to the upstream Notion Agents SDK JS (`makenotion/notion-agents-sdk-js`) and produce a feature parity report against this Go SDK.

## Arguments

- Optional: a time range like "last 7 days", "last 30 days", "since 2026-01-01". Defaults to last 30 days.

## Steps

### 1. Fetch recent upstream changes

Use `gh` CLI to get recent commits and their diffs from the upstream repo:

```bash
# Recent commits (adjust --since based on argument)
gh api repos/makenotion/notion-agents-sdk-js/commits --jq '.[].commit.message' -q 'since=YYYY-MM-DDTHH:MM:SSZ'

# For each interesting commit, get the diff to understand what changed
gh api repos/makenotion/notion-agents-sdk-js/commits/{sha} --jq '.files[].filename'
```

Focus on commits that touch `src/` files (types, client, streaming, etc.). Skip docs-only and dependency bump commits.

### 2. Fetch upstream type definitions and API surface

Read the key upstream source files to understand the current SDK surface:

```bash
# Get the source tree
gh api repos/makenotion/notion-agents-sdk-js/git/trees/main?recursive=1 --jq '.tree[].path' | grep '^src/'

# Read key type/interface files
gh api repos/makenotion/notion-agents-sdk-js/contents/src/{file} --jq '.content' | base64 -d
```

Look for:
- Exported types and interfaces (request/response shapes)
- Client methods (API endpoints being called)
- Streaming event types
- Error classification types

### 3. Analyze our Go SDK

Read the SDK source files in this repo. Catalog:
- All struct types in `types.go`
- All client and agent methods across `client.go`, `agent.go`, `agents.go`, `thread.go`
- Stream chunk types handled in `stream.go`
- Pagination iterators in `pagination.go`
- Error types in `errors.go`
- Helper functions in `helpers.go`

### 4. Produce the parity report

Output a markdown report with these sections:

#### New Upstream Features
Features added upstream that we don't have. For each:
- What it is and which upstream commit/PR introduced it
- The API endpoint or type involved
- Implementation complexity estimate (small/medium/large)

#### Type/Struct Gaps
Fields or types present upstream but missing from our Go structs. Show the upstream definition and what we need to add.

#### Behavioral Differences
Differences in how we handle things vs upstream (error handling, polling, streaming, pagination).

#### Parity Summary Table

| Feature | Upstream | Our SDK | Priority | Effort |
|---------|----------|---------|----------|--------|
| ... | ... | ... | High/Med/Low | Small/Med/Large |

#### Recommended Next Steps
Ordered list of what to implement next, prioritized by:
1. Features that affect correctness (breaking changes, new required fields)
2. Features that users would expect (attachments, verbose output)
3. Nice-to-haves (pagination helpers, error classification)

## Important

- Use `gh api` for all GitHub API calls (do not clone the repo)
- Compare against the actual code, not just README descriptions
- Be specific about Go struct changes needed (show the field to add, the JSON tag, etc.)
- If `gh` is not authenticated, tell the user to run `gh auth login` first
