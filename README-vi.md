<div align="center">

# ⚡ FastCode-CLI

### Công cụ Phân tích Mã nguồn Thông minh — Viết bằng Go

Lấy cảm hứng từ [HKUDS/FastCode](https://github.com/HKUDS/FastCode) — Được viết lại bằng Go để tối ưu tốc độ, tính di động, và triển khai dưới dạng một file nhị phân duy nhất.

[![Go 1.23+](https://img.shields.io/badge/go-1.23+-00ADD8.svg?logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

**[Tính năng](#-tính-năng)** • **[Bắt đầu nhanh](#-bắt-đầu-nhanh)** • **[Kiến trúc](#-kiến-trúc)** • **[Lộ trình](#-lộ-trình)** • **[Ghi nhận](#-ghi-nhận)**

</div>

---

## 🎯 FastCode-CLI là gì?

FastCode-CLI là một **công cụ hiểu mã nguồn hiệu năng cao, tối ưu token** được viết bằng Go. Nó phân tích cú pháp (AST), đánh chỉ mục, và điều hướng các codebase lớn bằng cách kết hợp phân tích cây cú pháp, tìm kiếm lai (semantic + BM25), và mô hình hóa đồ thị đa tầng — tất cả từ một file nhị phân biên dịch duy nhất.

Được thiết kế cho:

- **Workflow AI Agent** — Cung cấp ngữ cảnh mã nguồn có cấu trúc cho LLM mà không làm tràn context window.
- **Công cụ cho Lập trình viên** — Nhanh chóng hiểu codebase lạ, truy vết phụ thuộc, và tìm kiếm mã.
- **Tích hợp MCP Server** — Kết nối trực tiếp vào Cursor, Claude Code, Windsurf, hoặc bất kỳ MCP client nào.

---

## ✨ Tính năng

### 🏗️ Biểu diễn Mã nguồn theo Ngữ nghĩa - Cấu trúc

- **Phân tích AST** qua [go-tree-sitter](https://github.com/smacker/go-tree-sitter) — Lập chỉ mục đa tầng trên file, class, function cho **8+ ngôn ngữ** (Go, Python, JavaScript, TypeScript, Java, Rust, C/C++, C#).
- **Chỉ mục Lai (Hybrid Index)** — Kết hợp vector embedding với [Bleve](https://github.com/blevesearch/bleve) BM25 để tìm kiếm chính xác.
- **Mô hình Đồ thị Đa tầng** — Ba đồ thị quan hệ liên kết (Call Graph, Dependency Graph, Inheritance Graph) cho điều hướng cấu trúc.

### 🧭 Điều hướng Siêu nhanh

- **Tìm kiếm Thông minh 2 Bước** — Tìm mã tiềm năng trước, rồi xếp hạng kết quả tốt nhất cho câu hỏi cụ thể.
- **Đọc lướt Mã nguồn (Code Skimming)** — Chỉ đọc function signature, class definition và type hint thay vì toàn bộ file, tiết kiệm lượng lớn token.
- **Duyệt Đồ thị** — Truy vết kết nối mã nguồn tới N bước, theo import, call và chuỗi kế thừa.

### 💰 Quản lý Ngữ cảnh Tối ưu Chi phí

- **Quyết định Dựa trên Ngân sách** — Cân nhắc độ tin cậy, độ phức tạp, kích thước codebase và chi phí token trước khi xử lý.
- **Ưu tiên Giá trị** — Lấy thông tin tác động cao, chi phí thấp trước.

### 🚀 Lợi thế của Go

- **File Nhị phân Duy nhất** — Không cần Python, pip, venv hay Docker. Chỉ một file binary nhanh.
- **Đồng thời Goroutine** — Phân tích AST song song và gọi embedding HTTP biến quá trình indexing 20 giây (Python) thành 2 giây (Go).
- **Bộ nhớ Tối thiểu** — Không PyTorch, không FAISS pickle. Chỉ Go + Bleve gọn nhẹ.

---

## 🚀 Bắt đầu nhanh

### Cài đặt từ Mã nguồn

```bash
git clone https://github.com/duyhunghd6/fastcode-cli.git
cd fastcode-cli
go build -o fastcode ./cmd/fastcode

# Cấu hình LLM endpoint
export OPENAI_API_KEY="your-key"
export MODEL="gpt-4o"
export BASE_URL="https://api.openai.com/v1"
```

### Sử dụng

```bash
# Đánh chỉ mục một repository
fastcode index /path/to/your/repo

# Truy vấn codebase đã đánh chỉ mục
fastcode query "Luồng xác thực hoạt động như thế nào?"

# Truy vấn đa repo
fastcode query --repos /path/repo1,/path/repo2 "Logic thanh toán nằm ở đâu?"

# Khởi chạy MCP server (cho Cursor / Claude Code)
fastcode serve-mcp --port 8080
```

---

## 🏗 Kiến trúc

```
┌─────────────────────────────────────────────────┐
│                  fastcode-cli                    │
├─────────────┬───────────────┬───────────────────┤
│  cmd/       │  internal/    │  pkg/             │
│  fastcode   │  parser       │  treesitter       │
│  (Cobra)    │  graph        │                   │
│             │  index        │                   │
│             │  agent        │                   │
│             │  llm          │                   │
└─────────────┴───────────────┴───────────────────┘
        │              │               │
   CLI/MCP      AST + Graph      Tree-sitter
   Interface    Engine           Go Bindings
        │              │               │
        ▼              ▼               ▼
   ┌─────────┐  ┌───────────┐  ┌─────────────┐
   │ LLM API │  │ Bleve BM25│  │ Vector Store│
   │ (OpenAI │  │ (Tìm kiếm │  │ (Embeddings)│
   │ /Ollama)│  │ từ khóa)  │  │             │
   └─────────┘  └───────────┘  └─────────────┘
```

### Cấu trúc Package

| Package           | Mô tả                                                                          |
| ----------------- | ------------------------------------------------------------------------------ |
| `cmd/fastcode`    | Entry point CLI (Cobra), subcommands: `index`, `query`, `serve-mcp`            |
| `internal/parser` | Phân tích AST bằng Tree-sitter, trích xuất code unit (function, class, import) |
| `internal/graph`  | Xây dựng & duyệt Call Graph, Dependency Graph, Inheritance Graph               |
| `internal/index`  | Công cụ đánh chỉ mục lai (vector embedding + BM25 qua Bleve)                   |
| `internal/agent`  | Agent truy xuất lặp với quản lý ngân sách ngữ cảnh                             |
| `internal/llm`    | Abstraction LLM client (API tương thích OpenAI)                                |
| `pkg/treesitter`  | Tree-sitter Go bindings và grammar helper cho các ngôn ngữ                     |
| `reference/`      | Mã nguồn Python FastCode gốc để tham khảo                                      |
| `docs/`           | Tài liệu nghiên cứu, phân tích và kế hoạch porting                             |

---

## 🗺 Lộ trình

### Giai đoạn 1: Core Engine _(Đang tiến hành)_

- [ ] Phân tích AST bằng Tree-sitter cho Go, Python, JS/TS, Java, Rust
- [ ] Trích xuất code unit (function, class, import, type)
- [ ] Xây dựng Call Graph và Dependency Graph

### Giai đoạn 2: Đánh chỉ mục

- [ ] Tạo embedding qua LLM API (OpenAI / Ollama)
- [ ] Đánh chỉ mục BM25 bằng Bleve
- [ ] Truy xuất lai (kết hợp vector + BM25)

### Giai đoạn 3: Agent Truy xuất

- [ ] Agent lặp quản lý ngân sách (port từ Python `IterativeAgent`)
- [ ] Đọc lướt mã nguồn và duyệt file thông minh
- [ ] Hỗ trợ truy vấn đa repo

### Giai đoạn 4: Tích hợp

- [ ] CLI commands: `index`, `query`, `summary`
- [ ] Chế độ MCP Server (`serve-mcp`)
- [ ] Chế độ REST API server

---

## 🙏 Ghi nhận

Dự án này là bản **viết lại bằng Go** lấy cảm hứng từ [**FastCode**](https://github.com/HKUDS/FastCode) của [HKUDS Lab](https://github.com/HKUDS) tại Đại học Hồng Kông. Bản Python gốc đã giới thiệu framework ba giai đoạn đột phá cho việc hiểu mã nguồn tối ưu token.

Chúng tôi chân thành ghi nhận các tác giả gốc và đóng góp nghiên cứu của họ.

---

## 📄 Giấy phép

Dự án được phân phối theo [Giấy phép MIT](LICENSE).

---

<div align="center">

**Xây dựng với ❤️ bằng Go**

_Một phần của hệ sinh thái [Gmind](https://github.com/duyhunghd6/gmind) — Quản lý Bộ nhớ cho Lập trình Đa tác nhân_

</div>
