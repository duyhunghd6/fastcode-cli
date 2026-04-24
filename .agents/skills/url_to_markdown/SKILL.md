---
name: URL to Markdown Converter
description: Downloads a web page from a given URL and converts its content into a Markdown format file. Uses Scrapling for fetching and html2text for parsing.
---

# URL to Markdown Converter Skill

This skill allows you to quickly fetch the content of any web page and save it as a readable Markdown file. It's highly useful for extracting documentation, articles, or any text-based content for offline usage or further processing by LLMs.

## Prerequisites

This skill requires the `scrapling` and `html2text` python packages. They should be installed in the environment where this skill is run.

```bash
pip install scrapling html2text
```

## Usage

Use the `run_command` tool to execute the `url_to_markdown.py` script.

### Syntax

```bash
python3 .agents/skills/url_to_markdown/url_to_markdown.py <URL> <OUTPUT_FILE.md>
```

- `<URL>`: The full HTTP/HTTPS URL of the page you want to download.
- `<OUTPUT_FILE.md>`: The absolute or relative path where you want the resulting markdown file to be saved.

### Example

```bash
python3 .agents/skills/url_to_markdown/url_to_markdown.py https://example.com output.md
```

## How it works internally

1. **Fetching**: It uses the robust `scrapling` library (`scrapling.fetchers.Fetcher`) to download the raw HTML of the target URL.
2. **Parsing**: It extracts the `<html>` content block using XPath.
3. **Conversion**: The `html2text` library processes the HTML string, stripping unnecessary tags while preserving formatting elements like headers, lists, links, and text styling, converting it all into valid Markdown.
4. **Saving**: The final markdown string is written to the specified output file with UTF-8 encoding.
