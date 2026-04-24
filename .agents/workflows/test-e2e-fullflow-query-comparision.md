---
description: Run E2E full-flow comparison between FastCode Go and Python by logging ALL multi-step LLM interactions (requests + responses)
---

# E2E Full-Flow Query Comparison Workflow: Go vs Python

This workflow compares the **complete iterative agent loop** of FastCode Go vs Python. Unlike the single-prompt comparison (`test-e2e-query-comparison`), this captures **every** LLM call across multiple rounds — including tool call decisions and their results. Both systems run against a real LLM and produce numbered request/response JSON files for pair-by-pair analysis.

> ⚠️ **This workflow consumes LLM tokens** — both Go and Python run full query sessions.

## Prerequisites

- Go 1.21+ installed
- FastCode-CLI Go binary buildable at `~/duyhunghd6/fastcode-cli`
- Python reference at `~/duyhunghd6/gmind/reference/FastCode` with `.venv` set up
- A target repository (default: `~/duyhunghd6/music-theory`)
- **LLM API key configured** (both Go and Python need working LLM access)

---

## Step 1: Build the Go binary

// turbo

```bash
cd /Users/steve/duyhunghd6/fastcode-cli && go mod tidy && go build -o fastcode ./cmd/fastcode/
```

## Step 2: Ensure full-flow interceptors are active

Both Go and Python codebases support the `FASTCODE_DEBUG_PROMPT_DIR` environment variable.
When set to a directory path, the codebases will:

1. For **each** LLM call, write the request to `{dir}/call_{N}_request.json`
2. **Proceed** with the actual LLM API call (NOT abort)
3. Write the response to `{dir}/call_{N}_response.json`
4. Increment N for the next call

This differs from `FASTCODE_DEBUG_PROMPT_FILE` which aborts after the first call.

## Step 3: Run the full-flow comparison script

// turbo-all

```bash
chmod +x /Users/steve/duyhunghd6/fastcode-cli/scripts/e2e-fullflow-query-compare.sh
```

```bash
/Users/steve/duyhunghd6/fastcode-cli/scripts/e2e-fullflow-query-compare.sh /Users/steve/duyhunghd6/music-theory "how is audio played?"
```

## Step 4: Evaluate Each Call Pair

The script outputs a pair-by-pair comparison table and tells you where the full files are saved (`/tmp/fastcode_go_fullflow/` and `/tmp/fastcode_py_fullflow/`).

For each call pair, compare across these dimensions:

| Dimension                | Question to Evaluate                                                        |
| ------------------------ | --------------------------------------------------------------------------- |
| **Number of Calls**      | Does Go make the same number of LLM calls as Python? If not, why?           |
| **System Prompt**        | Is the system prompt equivalent in each round?                              |
| **User Query / Context** | Is the query and gathered context passed identically at each round?         |
| **Tool definitions**     | Are the tools/functions provided to the LLM identical in intent and schema? |
| **Tool Call Decisions**  | Does the LLM make the same tool call decisions in both systems?             |
| **Tool Call Results**    | Do the tool executions return equivalent results?                           |
| **Confidence Scores**    | Do confidence assessments track similarly across rounds?                    |
| **Final Answer**         | Are the final answers comparable in quality and content?                    |

### Deep Comparison Commands

To diff a specific call pair:

```bash
# Compare request structures
diff <(python3 -m json.tool /tmp/fastcode_go_fullflow/call_001_request.json) \
     <(python3 -m json.tool /tmp/fastcode_py_fullflow/call_001_request.json)

# Compare response content
diff <(python3 -m json.tool /tmp/fastcode_go_fullflow/call_001_response.json) \
     <(python3 -m json.tool /tmp/fastcode_py_fullflow/call_001_response.json)
```

### Pass Criteria

- ✅ Both systems successfully capture all LLM calls (request + response pairs) in numbered JSON files.
- ✅ The number of LLM rounds is similar (±1 round is acceptable due to LLM non-determinism).
- ✅ System prompts and tool definitions are structurally equivalent.
- ✅ Tool call patterns follow the same strategy (search → browse → refine).
- ✅ Final answers address the same aspects of the query.

### Analysis Recommendations

When pairs diverge, investigate:

1. **Code comparison**: Compare the corresponding Go and Python agent code to identify logic differences.
2. **Prompt engineering**: Check if system prompts or round prompts differ in ways that change LLM behavior.
3. **Tool schema**: Verify tool definitions match between Go and Python.
4. **Context assembly**: Check how gathered elements are formatted into the prompt context.
