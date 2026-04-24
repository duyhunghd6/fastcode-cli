---
description: Run E2E query capability comparison between FastCode Go and original FastCode Python by intercepting LLM prompts
---

# E2E Query Prompt Comparison Workflow: Go vs Python

This workflow compares the **query** step of FastCode Go vs FastCode Python. Because LLM answers are naturally non-deterministic, we compare the **exact prompts and tool sets** that are passed to the context window just before the LLM is triggered. Both systems are configured to intercept the prompt call, dump it to JSON, and abort before spending tokens.

## Prerequisites

- Go 1.21+ installed
- FastCode-CLI Go binary buildable at `~/duyhunghd6/fastcode-cli`
- Python reference at `~/duyhunghd6/gmind/reference/FastCode` with `.venv` set up
- A target repository (default: `~/duyhunghd6/music-theory`)
- LLM API key configured (for agent-based query in Go; for LLM query enhancement in Python)

---

## Step 1: Build the Go binary

// turbo

```bash
cd /Users/steve/duyhunghd6/fastcode-cli && go mod tidy && go build -o fastcode ./cmd/fastcode/
```

## Step 2: Ensure prompt debug interceptors are active

The Go and Python codebases must both support the `FASTCODE_DEBUG_PROMPT_FILE` environment variable.
When this environment variable is set to a file path, the codebases should:

1. Serialize the LLM request (messages, model, temperature, tools/functions) to the provided file path.
2. Return a dummy response (like "DEBUG_PROMPT_WRITTEN") instead of making the actual API call.

## Step 3: Run the comparison script

The custom comparison script prepares identical test conditions and executes the queries while injecting the interceptor hooks. It outputs a summary of the captured prompts.

// turbo-all

```bash
/Users/steve/duyhunghd6/fastcode-cli/scripts/e2e-query-compare.sh /Users/steve/duyhunghd6/music-theory "how is audio played?"
```

## Step 4: Evaluate the Prompts Structure

The script dumps the initial segment of the system messages directly in the terminal, and tells you where the full files are saved (e.g., `/tmp/fastcode_go_prompt.json` and `/tmp/fastcode_py_prompt.json`).

Compare the output JSON payloads across these dimensions:

| Dimension            | Question to Evaluate                                                                                        |
| -------------------- | ----------------------------------------------------------------------------------------------------------- |
| **System Prompt**    | Are the instructions basically equivalent? Is Go missing any crucial behavioral guardrails found in Python? |
| **User Query**       | Is the query mapped identically?                                                                            |
| **Tool definitions** | Are the functions/schema signatures provided to the LLM (if any) identical in intent and arguments?         |
| **Context**          | Are any preliminary BM25 search results or repo overviews presented identically?                            |

### Pass Criteria

- ✅ Both systems successfully intercept and dump JSON objects before hitting the LLM API.
- ✅ The core system instruction structure is equivalent.
- ✅ Necessary capabilities (tools for retrieving specific files, definitions, directories) match the Python implementation logic.
