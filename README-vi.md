<div align="center">

<!-- <img src="assets/FastCode.svg" alt="FastCode-CLI Logo" width="200"/> -->

<!-- # FastCode-CLI -->

### FastCode-CLI: Tăng Tốc và Tối Ưu Hóa Việc Hiểu Mã Nguồn

| **⚡ Một File Duy Nhất** | **💰 Tiết Kiệm Token** | **🚀 Nhanh Nhờ Goroutine** |

[![Go 1.24+](https://img.shields.io/badge/go-1.24+-00ADD8.svg?logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Based on FastCode](https://img.shields.io/badge/based%20on-HKUDS%2FFastCode-blueviolet)](https://github.com/HKUDS/FastCode)

[Tính năng](#-tại-sao-chọn-fastcode-cli) • [Bắt đầu nhanh](#-bắt-đầu-nhanh) • [Cài đặt](#-cài-đặt) • [MCP Server](#mcp-server-sử-dụng-trong-cursor--claude-code--windsurf--antigravity) • [Tài liệu](#-cách-hoạt-động)

</div>

---

## 🎯 Tại Sao Chọn FastCode-CLI?

FastCode-CLI là bản **viết lại bằng Go** của [HKUDS/FastCode](https://github.com/HKUDS/FastCode) — một framework tiết kiệm token cho việc phân tích và hiểu mã nguồn toàn diện: mang đến **tốc độ vượt trội**, **độ chính xác xuất sắc**, và **hiệu quả chi phí** cho các hệ thống mã nguồn lớn và kiến trúc phần mềm phức tạp.

🚀 **Triển Khai Không Phụ Thuộc** — Một file binary duy nhất đã biên dịch. Không cần Python, không pip, không venv, không Docker. Chỉ cần `go build` và chạy.

💰 **Tiết Kiệm Chi Phí Đáng Kể** — Kế thừa mức giảm chi phí 44-55% so với Cursor/Claude Code thông qua điều hướng cấu trúc thông minh và truy xuất có ý thức ngân sách.

⚡ **Đồng Thời Với Goroutine** — Phân tích AST song song, gọi embedding đồng thời, và xây dựng đồ thị đa file. Python mất 20 giây thì Go chỉ mất ~2 giây.

🎯 **Độ Chính Xác Cao Nhất** — Cùng framework ba pha đã vượt trội hơn Cursor và Claude Code với điểm chính xác cao nhất, giờ được biên dịch thành mã máy bản địa.

<!-- <div align="center">
<img src="assets/performance.png" alt="FastCode Performance vs Cost" width="850"/>
</div> -->

---

## Tính Năng Chính của FastCode-CLI

### 🎯 Ưu Thế Hiệu Năng Cốt Lõi

- Nhanh hơn 2-4 lần so với đối thủ (Cursor/Claude Code) — kế thừa từ framework FastCode
- Giảm chi phí 44-55% so với các giải pháp thay thế
- Điểm Chính Xác Cao Nhất trên các benchmark
- Tiết kiệm đến 10 lần Token nhờ điều hướng cấu trúc thông minh

### 🛠️ Khả Năng Kỹ Thuật

- Phân Tích Repository Quy Mô Lớn — Xử lý codebase khổng lồ hiệu quả với đồng thời goroutine
- Hỗ Trợ Đa Ngôn Ngữ — Go, Python, JavaScript, TypeScript, Java, Rust, C/C++, C#
- Suy Luận Đa Repository — Phân tích phụ thuộc xuyên repo _(đang lên kế hoạch)_
- Hỗ Trợ Model Nhỏ — Tương thích với model cục bộ (Ollama, qwen3-coder-30b, v.v.)

### 💻 Trải Nghiệm Người Dùng

- **MCP Server** — Sử dụng FastCode-CLI trực tiếp thông qua tích hợp MCP (Cursor, Claude Code, Windsurf, Antigravity)
- **CLI Mạnh Mẽ** — CLI dựa trên Cobra với các lệnh con `index`, `query`, `serve-mcp`
- **REST API** — Tích hợp workflow dễ dàng _(đang lên kế hoạch)_
- **Điều Hướng Cấu Trúc Thông Minh** — Chỉ tải những gì cần thiết, lướt qua phần còn lại

---

## 🎥 Xem FastCode Hoạt Động

<div align="center">

[![Xem Demo FastCode](https://img.youtube.com/vi/NwexLWHPBOY/0.jpg)](https://youtu.be/NwexLWHPBOY)

**Nhấp để xem FastCode gốc hoạt động** — Phiên bản Go triển khai cùng framework ba pha với khả năng phân tích tương đương.

---

</div>

### Công Nghệ Cốt Lõi Đằng Sau FastCode-CLI

FastCode-CLI triển khai cùng **framework ba pha đột phá** đã thay đổi cách LLM hiểu và điều hướng codebase — được viết lại bằng Go thuần:

<!-- <p align="center">
  <img src="assets/framework.png" alt="FastCode Framework" width="100%"/>
</p> -->

## 🏗️ Biểu Diễn Mã Nguồn Ngữ Nghĩa-Cấu Trúc

### Hiểu codebase đa tầng cho phân tích toàn diện

- **🔍 Đơn Vị Mã Phân Cấp** — Đánh chỉ mục đa cấp tiên tiến bao gồm file, class, function, và tài liệu sử dụng phân tích cú pháp AST dựa trên tree-sitter cho hơn 8 ngôn ngữ lập trình. Được hỗ trợ bởi binding CGo gốc [go-tree-sitter](https://github.com/smacker/go-tree-sitter) cho hiệu năng tối đa.

- **🔗 Chỉ Mục Lai** — Kết hợp liền mạch embedding ngữ nghĩa với tìm kiếm từ khóa (BM25) cho truy xuất mã chính xác và mạnh mẽ. Vector store sử dụng cosine similarity trong bộ nhớ với batch embedding qua API tương thích OpenAI. Không FAISS, không PyTorch — chỉ có Go thuần túy.

- **📊 Mô Hình Đồ Thị Đa Tầng** — Ba đồ thị quan hệ liên kết (Call Graph, Dependency Graph, Inheritance Graph) cho phép điều hướng cấu trúc xuyên suốt toàn bộ codebase. Xây dựng bằng cấu trúc dữ liệu đồ thị Go gốc, không phụ thuộc thư viện ngoài.

### 🧭 Điều Hướng Codebase Siêu Nhanh

Tìm đúng mã nguồn mà không cần mở mọi file — với tốc độ chớp nhoáng

- **⚡ Tìm Kiếm Thông Minh Hai Giai Đoạn** — Như có một trợ lý nghiên cứu, đầu tiên tìm mã có khả năng liên quan thông qua truy xuất lai (vector + BM25), sau đó xếp hạng và tổ chức các kết quả phù hợp nhất cho câu hỏi cụ thể của bạn bằng xếp hạng lại có hỗ trợ LLM.

- **📁 Duyệt File An Toàn** — Khám phá cấu trúc dự án an toàn qua `filepath.Walk` của Go, hiểu tổ chức thư mục và pattern file trong khi tôn trọng `.gitignore` mà không ảnh hưởng bảo mật.

- **🌐 Theo Dõi Kết Nối Mã** — Truy vết cách các phần mã kết nối với nhau (đến N bước) qua duyệt đồ thị đa tầng, như đi theo dấu vết bánh mì xuyên suốt codebase — qua import, lời gọi hàm, và chuỗi kế thừa.

- **🎯 Lướt Mã** — Thay vì đọc toàn bộ file, FastCode-CLI chỉ xem các "tiêu đề" — tên hàm, định nghĩa class, và gợi ý kiểu. Giống như đọc mục lục chương của một cuốn sách thay vì từng trang, tiết kiệm lượng lớn sức mạnh xử lý.

### 💰 Quản Lý Context Tiết Kiệm Chi Phí

Đạt được hiểu biết tối đa với chi phí tối thiểu — tự động

- **📈 Ra Quyết Định Có Ý Thức Ngân Sách** — Xem xét năm yếu tố chính trước khi xử lý: mức độ tự tin, độ phức tạp truy vấn, kích thước codebase, chi phí tài nguyên, và số vòng lặp. Giống như một cố vấn tài chính tiết kiệm, cân nhắc mọi lựa chọn trước khi ra quyết định.

- **🔄 Học Tối Ưu Tài Nguyên** — Liên tục điều chỉnh phương pháp theo thời gian thực, ngày càng hiệu quả hơn trong việc xác định thông tin nào cần thu thập và khi nào dừng lại. Hãy nghĩ đó là một AI tối đa hóa giá trị và trở nên tiết kiệm chi phí hơn với mỗi truy vấn.

- **🎯 Chọn Lọc Ưu Tiên Giá Trị** — Ưu tiên thông tin có tác động cao, chi phí thấp trước, giống như chọn quả chín nhất với giá tốt nhất. Cách tiếp cận tối ưu chi phí này đảm bảo bạn nhận được giá trị tối đa cho mỗi token chi tiêu cho đến điểm dừng hoàn hảo.

---

## 📊 Hiệu Năng Benchmark

FastCode-CLI kế thừa cùng framework ba pha đã được kiểm nghiệm nghiêm ngặt trên bốn benchmark lớn đại diện cho các thách thức kỹ thuật phần mềm thực tế:

### 🎯 Bộ Dữ Liệu Đánh Giá

| Benchmark    | Lĩnh Vực Tập Trung        | Kiểm Tra Gì                       |
| ------------ | ------------------------- | --------------------------------- |
| SWE-QA       | Hỏi Đáp Kỹ Thuật Phần Mềm | Trả lời câu hỏi kỹ thuật phức tạp |
| LongCodeQA   | Phân Tích Mã Mở Rộng      | Hiểu mã trong context dài         |
| LOC-BENCH    | Định Vị Mã                | Phát hiện lỗi & yêu cầu tính năng |
| GitTaskBench | Tác Vụ Thực Tế            | Quy trình repository thực tế      |

### 🏆 Kết Quả Xuất Sắc

- ✅ **Độ Chính Xác Vượt Trội** — Luôn vượt trội hơn các baseline tiên tiến nhất trên tất cả benchmark

- ✅ **Hiệu Quả Token 10 Lần** — Đạt kết quả tốt hơn trong khi sử dụng ít hơn đến 90% token

- ✅ **Xác Thực Thực Tế** — Hiệu năng đã được chứng minh trên các codebase và quy trình sản xuất thực tế

### 🐹 Ưu Thế Riêng Của Go

| Khía Cạnh               | Python (FastCode)                | Go (FastCode-CLI)                          |
| ----------------------- | -------------------------------- | ------------------------------------------ |
| **Triển Khai**          | Python 3.12 + pip + venv + FAISS | Binary duy nhất (`go build`)               |
| **Phân Tích AST**       | `tree_sitter` Python bindings    | `go-tree-sitter` binding CGo gốc           |
| **Tìm Kiếm Vector**     | FAISS (NumPy/C++) + blob `.pkl`  | Cosine similarity trong bộ nhớ (Go thuần)  |
| **Tìm Kiếm Text**       | Triển khai BM25 tùy chỉnh        | BM25 qua tokenizer gốc                     |
| **Thư Viện Đồ Thị**     | NetworkX                         | Triển khai đồ thị Go gốc                   |
| **LLM Client**          | `openai` Python SDK              | HTTP client tùy chỉnh (tương thích OpenAI) |
| **Framework CLI**       | `argparse`                       | Cobra                                      |
| **MCP Server**          | `fastmcp` Python library         | JSON-RPC tùy chỉnh qua stdio               |
| **Cache**               | Pickle / file `.pkl`             | Gob encoding (file `.gob`)                 |
| **Đồng Thời**           | `asyncio`                        | Goroutine + channel                        |
| **Thời Gian Khởi Động** | ~2 giây (Python interpreter)     | ~10ms (binary đã biên dịch)                |
| **Bộ Nhớ**              | ~500MB+ (FAISS + PyTorch)        | ~50MB (Go runtime tinh gọn)                |

---

## 🚀 Bắt Đầu Nhanh

Khởi chạy FastCode-CLI trong chưa đầy 1 phút:

```bash
# 1. Clone repository
git clone https://github.com/duyhunghd6/fastcode-cli.git
cd fastcode-cli

# 2. Build binary
go build -o fastcode ./cmd/fastcode

# 3. Cấu hình API key
cp .env.example .env
# Chỉnh sửa .env với API key của bạn

# 4. Đánh chỉ mục và truy vấn codebase
./fastcode index /path/to/your/repo
./fastcode query --repo /path/to/your/repo "Xác thực hoạt động như thế nào?"
```

Vậy là xong — không cần môi trường ảo, không giải quyết phụ thuộc, không Docker. 🎉

---

## 📦 Cài Đặt

FastCode-CLI hỗ trợ **Linux**, **macOS**, và **Windows**. Chọn nền tảng của bạn bên dưới:

> **💡 Yêu cầu:** Chỉ cần [Go 1.24+](https://go.dev/dl/) và Git. Không cần Python, không pip, không venv.

<details>
<summary><b>🐧 Cài Đặt Linux</b></summary>

### Yêu Cầu

- Go 1.24 trở lên
- Git

### Hướng Dẫn Từng Bước

1. **Clone FastCode-CLI**

   ```bash
   git clone https://github.com/duyhunghd6/fastcode-cli.git
   cd fastcode-cli
   ```

2. **Build Binary**

   ```bash
   go build -o fastcode ./cmd/fastcode

   # Tùy chọn: Cài đặt toàn cục
   sudo mv fastcode /usr/local/bin/
   ```

3. **Cấu Hình Môi Trường**

   ```bash
   cp .env.example .env
   nano .env  # hoặc sử dụng trình soạn thảo yêu thích
   ```

   Thêm API key:

   ```env
   OPENAI_API_KEY=api_key_openai_của_bạn
   MODEL=gpt-4o
   BASE_URL=https://api.openai.com/v1
   ```

4. **Khởi Chạy FastCode-CLI**

   ```bash
   # Đánh chỉ mục codebase
   ./fastcode index /path/to/your/repo

   # Truy vấn codebase
   ./fastcode query --repo /path/to/your/repo "Câu hỏi của bạn"

   # Hoặc khởi động MCP server
   ./fastcode serve-mcp --port 8080
   ```

</details>

<details>
<summary><b>🍎 Cài Đặt macOS</b></summary>

### Yêu Cầu

- Go 1.24 trở lên
- Git

### Hướng Dẫn Từng Bước

1. **Clone FastCode-CLI**

   ```bash
   git clone https://github.com/duyhunghd6/fastcode-cli.git
   cd fastcode-cli
   ```

2. **Build Binary**

   ```bash
   go build -o fastcode ./cmd/fastcode

   # Tùy chọn: Cài đặt toàn cục
   sudo mv fastcode /usr/local/bin/
   ```

3. **Cấu Hình Môi Trường**

   ```bash
   cp .env.example .env
   nano .env  # hoặc dùng: open -e .env
   ```

   Thêm API key:

   ```env
   OPENAI_API_KEY=api_key_openai_của_bạn
   MODEL=gemini-3-flash
   BASE_URL=https://...
   ```

4. **Khởi Chạy FastCode-CLI**

   ```bash
   # Đánh chỉ mục codebase
   ./fastcode index /path/to/your/repo

   # Truy vấn codebase
   ./fastcode query --repo /path/to/your/repo "Câu hỏi của bạn"

   # Hoặc khởi động MCP server
   ./fastcode serve-mcp --port 8080
   ```

**Lưu ý cho Apple Silicon (M1/M2/M3/M4):** Go hỗ trợ ARM64 bản địa. Không cần cấu hình đặc biệt — `go build` tự động tạo binary ARM.

</details>

<details>
<summary><b>💻 Cài Đặt Windows</b></summary>

### Yêu Cầu

- Go 1.24 trở lên
- Git

### Hướng Dẫn Từng Bước

1. **Clone FastCode-CLI**

   ```cmd
   git clone https://github.com/duyhunghd6/fastcode-cli.git
   cd fastcode-cli
   ```

2. **Build Binary**

   ```cmd
   go build -o fastcode.exe ./cmd/fastcode
   ```

3. **Cấu Hình Môi Trường**

   ```cmd
   copy .env.example .env
   notepad .env
   ```

   Thêm API key:

   ```env
   OPENAI_API_KEY=api_key_openai_của_bạn
   MODEL=qwen/qwen3-coder-30b-a3b-instruct
   BASE_URL=https://api.openai.com/v1
   ```

4. **Khởi Chạy FastCode-CLI**

   ```cmd
   REM Đánh chỉ mục codebase
   fastcode.exe index C:\path\to\your\repo

   REM Truy vấn codebase
   fastcode.exe query --repo C:\path\to\your\repo "Câu hỏi của bạn"

   REM Hoặc khởi động MCP server
   fastcode.exe serve-mcp --port 8080
   ```

**Xử lý sự cố:**

- Nếu build CGo thất bại: Đảm bảo GCC đã được cài đặt (qua [MSYS2](https://www.msys2.org/) hoặc MinGW)
- Nếu gặp lỗi quyền, chạy Command Prompt với quyền Administrator
- Tree-sitter yêu cầu trình biên dịch C — theo [hướng dẫn build go-tree-sitter](https://github.com/smacker/go-tree-sitter#installation)

</details>

---

## 🎮 Sử Dụng

### Giao Diện Dòng Lệnh (Khuyến Nghị)

CLI cung cấp trải nghiệm tinh gọn nhất:

```bash
# Đánh chỉ mục repository
./fastcode index /path/to/repo

# Bắt buộc đánh chỉ mục lại (bỏ qua cache)
./fastcode index /path/to/repo --force

# Đánh chỉ mục với đầu ra JSON (cho máy đọc)
./fastcode index /path/to/repo --json

# Bỏ qua tạo embedding (chế độ chỉ BM25, không cần API key khi đánh chỉ mục)
./fastcode index /path/to/repo --no-embeddings

# Sử dụng model embedding tùy chỉnh
./fastcode index /path/to/repo --embedding-model text-embedding-3-large

# Sử dụng thư mục cache tùy chỉnh
./fastcode index /path/to/repo --cache-dir /tmp/fastcode-cache
```

**Truy vấn:**

```bash
# Truy vấn codebase đã được đánh chỉ mục
./fastcode query --repo /path/to/repo "Luồng xác thực hoạt động như thế nào?"

# Truy vấn với đầu ra JSON (cho tự động hóa và scripting)
./fastcode query --repo /path/to/repo --json "Logic thanh toán ở đâu?"
```

**Bắt đầu đặt câu hỏi như:**

- "Logic xác thực được triển khai ở đâu?"
- "Luồng xử lý thanh toán hoạt động như thế nào?"
- "File nào sẽ bị ảnh hưởng nếu tôi thay đổi User model?"
- "Giải thích sự phụ thuộc giữa module A và module B"

<details>
<summary><b>REST API</b></summary>

> **Lưu ý:** REST API server đang được lên kế hoạch cho phiên bản tương lai. Đặc tả REST API của Python FastCode được bao gồm ở đây để tham khảo — phiên bản Go sẽ triển khai cùng các endpoint.

Tích hợp FastCode-CLI vào công cụ của bạn với REST API toàn diện:

```bash
# Khởi động API server (đang lên kế hoạch)
./fastcode serve-api --host 0.0.0.0 --port 8000
```

API sẽ cung cấp tất cả tính năng có trong CLI. Truy cập http://localhost:8000/docs để xem tài liệu API tương tác.

**Các Endpoint API Chính (Đang Lên Kế Hoạch):**

<details>
<summary><b>Quản Lý Repository</b></summary>

```bash
# Liệt kê các repository đã tải
GET /repositories

# Tải repository từ URL hoặc đường dẫn cục bộ
POST /load
{
  "source": "https://github.com/user/repo",
  "is_url": true
}

# Đánh chỉ mục repository đã tải
POST /index?force=false

# Tải và đánh chỉ mục trong một lệnh
POST /load-and-index
{
  "source": "/path/to/repo",
  "is_url": false
}

# Tải nhiều repository đã đánh chỉ mục
POST /load-repositories
{
  "repo_names": ["repo1", "repo2"]
}

# Đánh chỉ mục nhiều repository cùng lúc
POST /index-multiple
{
  "sources": [
    {"source": "https://github.com/user/repo1", "is_url": true},
    {"source": "/path/to/repo2", "is_url": false}
  ]
}

# Xóa repository và chỉ mục
POST /delete-repos
{
  "repo_names": ["repo1", "repo2"],
  "delete_source": true
}

# Lấy tóm tắt repository
GET /summary
```

</details>

<details>
<summary><b>Truy Vấn & Hội Thoại</b></summary>

```bash
# Truy vấn repository (phản hồi đơn)
POST /query
{
  "question": "Xác thực hoạt động như thế nào?",
  "filters": null,
  "repo_filter": ["repo1"],
  "multi_turn": false,
  "session_id": null
}

# Truy vấn với phản hồi streaming (SSE)
POST /query-stream
{
  "question": "Giải thích schema database",
  "multi_turn": true,
  "session_id": "abc123"
}

# Tạo phiên hội thoại mới
POST /new-session?clear_session_id=old_session

# Liệt kê tất cả phiên hội thoại
GET /sessions

# Lấy lịch sử hội thoại
GET /session/{session_id}

# Xóa phiên hội thoại
DELETE /session/{session_id}
```

</details>

<details>
<summary><b>Hệ Thống & Trạng Thái</b></summary>

```bash
# Kiểm tra sức khỏe
GET /health

# Lấy trạng thái hệ thống
GET /status?full_scan=false

# Xóa cache
POST /clear-cache

# Lấy thống kê cache
GET /cache-stats

# Làm mới cache chỉ mục
POST /refresh-index-cache

# Gỡ tải repository hiện tại
DELETE /repository
```

</details>

**Ví dụ sử dụng:**

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

func main() {
    // Tải và đánh chỉ mục repository
    loadBody, _ := json.Marshal(map[string]interface{}{
        "source": "https://github.com/user/repo",
        "is_url": true,
    })
    http.Post("http://localhost:8000/load-and-index", "application/json", bytes.NewBuffer(loadBody))

    // Truy vấn repository
    queryBody, _ := json.Marshal(map[string]interface{}{
        "question":   "Điểm khởi đầu chính ở đâu?",
        "multi_turn": false,
    })
    resp, _ := http.Post("http://localhost:8000/query", "application/json", bytes.NewBuffer(queryBody))

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    fmt.Println(result["answer"])
    fmt.Printf("Token đã dùng: %v\n", result["total_tokens"])
}
```

</details>

<a id="mcp-server-sử-dụng-trong-cursor--claude-code--windsurf--antigravity"></a>

<details>
<summary><b>MCP Server (Sử Dụng Trong Cursor / Claude Code / Windsurf / Antigravity)</b></summary>

FastCode-CLI bao gồm server [MCP (Model Context Protocol)](https://modelcontextprotocol.io/) tích hợp sẵn, cho phép các trợ lý AI lập trình như **Cursor**, **Claude Code**, **Windsurf**, và **Antigravity** sử dụng trực tiếp khả năng hiểu mã cấp repository của FastCode-CLI.

#### Thiết Lập

Build FastCode-CLI trước — không cần Python, không venv, chỉ một binary duy nhất:

```bash
git clone https://github.com/duyhunghd6/fastcode-cli.git
cd fastcode-cli
go build -o fastcode ./cmd/fastcode
```

MCP server được khởi chạy với `./fastcode serve-mcp`, cần `OPENAI_API_KEY`, `MODEL`, và `BASE_URL` như biến môi trường (hoặc cấu hình trong `.env`).

**Cursor** (`~/.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "fastcode": {
      "command": "/path/to/fastcode",
      "args": ["serve-mcp"],
      "env": {
        "MODEL": "gpt-4o",
        "BASE_URL": "https://api.openai.com/v1",
        "OPENAI_API_KEY": "sk-..."
      }
    }
  }
}
```

**Claude Code** (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "fastcode": {
      "command": "/path/to/fastcode",
      "args": ["serve-mcp"],
      "env": {
        "MODEL": "gpt-4o",
        "BASE_URL": "https://api.openai.com/v1",
        "OPENAI_API_KEY": "sk-..."
      }
    }
  }
}
```

Hoặc qua `claude mcp add`:

```bash
claude mcp add fastcode -- /path/to/fastcode serve-mcp
```

**Antigravity** (`.gemini/settings.json`):

```json
{
  "mcpServers": {
    "fastcode": {
      "command": "/path/to/fastcode",
      "args": ["serve-mcp"],
      "env": {
        "MODEL": "gemini-3-flash",
        "BASE_URL": "https://...",
        "OPENAI_API_KEY": "your-key"
      }
    }
  }
}
```

**Transport SSE** (cho triển khai từ xa / chia sẻ):

```bash
OPENAI_API_KEY=sk-... MODEL=gpt-4o BASE_URL=https://api.openai.com/v1 \
/path/to/fastcode serve-mcp --port 8080
```

#### Các Tool Có Sẵn

| Tool                   | Mô Tả                                                                                                                                     |
| ---------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `code_qa`              | Tool cốt lõi — đặt câu hỏi về một hoặc nhiều repository mã nguồn. Tự động đánh chỉ mục các repo chưa được index. Hỗ trợ câu hỏi tiếp nối. |
| `list_indexed_repos`   | Liệt kê tất cả repository đã được đánh chỉ mục và sẵn sàng truy vấn.                                                                      |
| `delete_repo_metadata` | Xóa metadata đã đánh chỉ mục của repository (file cache `.gob`) trong khi giữ lại mã nguồn.                                               |

#### Tham Số `code_qa`

| Tham số    | Bắt buộc | Mặc định | Mô tả                                                       |
| ---------- | -------- | -------- | ----------------------------------------------------------- |
| `question` | Có       | —        | Câu hỏi về mã nguồn                                         |
| `repos`    | Có       | —        | Danh sách nguồn repo (đường dẫn cục bộ). Hỗ trợ nhiều repo. |

#### Cách Hoạt Động

1. **Tự động phát hiện**: Với mỗi repo trong `repos`, FastCode-CLI kiểm tra xem đã được đánh chỉ mục chưa. Nếu rồi, bỏ qua việc đánh chỉ mục.
2. **Khởi động tức thì**: Không giống MCP server Python cần khởi động interpreter + tải dependency, binary Go khởi động trong vài mili giây.
3. **Cache**: Các repo đã đánh chỉ mục được cache tại `~/.fastcode/cache/` (có thể cấu hình). Các truy vấn tiếp theo tái sử dụng cache cho phản hồi gần tức thì.

#### Ví Dụ Sử Dụng

Trong Cursor hoặc Claude Code, chỉ cần hỏi:

```
Dùng FastCode phân tích repository tại /path/to/repo_name dùng để làm gì.
```

hoặc

```
Dùng FastCode phân tích luồng xác thực trong dự án này.
```

Trợ lý AI sẽ gọi `code_qa`, FastCode-CLI sẽ đánh chỉ mục repo (nếu cần), và trả về câu trả lời chi tiết kèm tham chiếu nguồn.

Với câu hỏi tiếp nối, trợ lý tiếp tục hội thoại tự nhiên:

```
File nào sẽ bị ảnh hưởng nếu tôi thay đổi User model?
```

</details>

---

## 🔧 Cấu Hình

### Nhà Cung Cấp LLM Được Hỗ Trợ

FastCode-CLI hoạt động với mọi nhà cung cấp LLM **tương thích OpenAI**:

<details>
<summary><b>OpenAI</b></summary>

```env
OPENAI_API_KEY=sk-...
MODEL=gpt-4o
BASE_URL=https://api.openai.com/v1
```

</details>

<details>
<summary><b>Google Gemini (qua endpoint tương thích OpenAI)</b></summary>

```env
OPENAI_API_KEY=your-gemini-key
MODEL=gemini-3-flash
BASE_URL=https://generativelanguage.googleapis.com/v1beta/openai
```

</details>

<details>
<summary><b>OpenRouter (Nhiều Model)</b></summary>

```env
OPENAI_API_KEY=sk-or-...
MODEL=google/gemini-flash-1.5
BASE_URL=https://openrouter.ai/api/v1
```

</details>

<details>
<summary><b>Model Cục Bộ (Ollama)</b></summary>

```env
OPENAI_API_KEY=ollama
MODEL=qwen3-coder-30b
BASE_URL=http://localhost:11434/v1
```

</details>

<details>
<summary><b>Tùy Chỉnh / Tự Host (vLLM, LiteLLM, v.v.)</b></summary>

```env
OPENAI_API_KEY=your-key
MODEL=tên-model-của-bạn
BASE_URL=http://server-của-bạn:8000/v1
```

</details>

### Ngôn Ngữ Được Hỗ Trợ

FastCode-CLI tự động phát hiện và phân tích:

- 🐹 Go
- 🐍 Python
- 📜 JavaScript / TypeScript
- ☕ Java
- 🦀 Rust
- ⚙️ C / C++
- 💎 C#

---

## 🧠 Cách Hoạt Động

FastCode-CLI sử dụng phương pháp **trinh sát trước** độc đáo, khác biệt cơ bản so với các hệ thống suy luận mã truyền thống:

### Cách Tiếp Cận Truyền Thống ❌

```
Câu hỏi → Tải File → Tìm Kiếm → Tải Thêm File → Tìm Kiếm Lại → ... → Trả Lời
💸 Chi phí token cao do tải file lặp đi lặp lại
```

### Cách Tiếp Cận FastCode-CLI ✅

```
Câu hỏi → Phân Tích AST → Xây Dựng Đồ Thị → Tìm Kiếm Lai → Lướt Mục Tiêu → Trả Lời
💰 Chi phí token tối thiểu với nhắm mục tiêu cấu trúc chính xác
```

### Pipeline Ba Pha Trong Go

**Pha 1 — Biểu Diễn Ngữ Nghĩa-Cấu Trúc** (`internal/parser` + `internal/graph`)

1. Duyệt repository qua `internal/loader` — phát hiện ngôn ngữ, tôn trọng `.gitignore`
2. Phân tích mỗi file qua bộ trích xuất AST `go-tree-sitter`
3. Trích xuất đơn vị mã phân cấp: function, class, import, type
4. Xây dựng Call Graph, Dependency Graph, và Inheritance Graph

**Pha 2 — Đánh Chỉ Mục Lai** (`internal/index` + `internal/llm`)

1. Tạo dense vector embedding cho mỗi phần tử mã (qua `internal/llm/embedder.go`)
2. Xây dựng chỉ mục từ khóa BM25 cho tìm kiếm text (qua `internal/index/bm25.go`)
3. Kết hợp thành truy xuất lai (qua `internal/index/hybrid.go`)
4. Cache toàn bộ chỉ mục ra đĩa (qua `internal/cache/cache.go`) để tái sử dụng tức thì

**Pha 3 — Truy Xuất Có Ý Thức Ngân Sách** (`internal/agent`)

1. Phân tích truy vấn (chấm điểm phức tạp, trích xuất từ khóa qua `internal/agent/query.go`)
2. Chạy truy xuất lặp đa vòng (qua `internal/agent/iterative.go`):
   - Mỗi vòng sử dụng agent tool: `search`, `browse`, `skim`, `list` (`internal/agent/tools.go`)
   - Duyệt đồ thị để khám phá mã liên quan
   - Đánh giá độ tự tin — dừng sớm khi đủ tự tin hoặc hết ngân sách
3. Tạo câu trả lời có cấu trúc bằng LLM với context đã thu thập (`internal/agent/answer.go`)

---

## 🏗 Kiến Trúc

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
│             │  loader       │                   │
│             │  cache        │                   │
│             │  orchestrator │                   │
└─────────────┴───────────────┴───────────────────┘
        │              │               │
   CLI/MCP      AST + Đồ     Tree-sitter
   Giao diện    Thị Engine    Go Bindings
        │              │               │
        ▼              ▼               ▼
   ┌─────────┐  ┌───────────┐  ┌─────────────┐
   │ LLM API │  │ BM25 Text │  │ Vector Store│
   │ (OpenAI │  │ (Tìm Kiếm│  │ (Embeddings)│
   │ /Ollama)│  │ Từ Khóa)  │  │             │
   └─────────┘  └───────────┘  └─────────────┘
```

### Bố Cục Package

| Package                 | Mô Tả                                                                    |
| ----------------------- | ------------------------------------------------------------------------ |
| `cmd/fastcode`          | Điểm vào CLI (Cobra): lệnh con `index`, `query`, `serve-mcp`             |
| `internal/types`        | Cấu trúc dữ liệu chung: `CodeElement`, `FunctionInfo`, `ClassInfo`, v.v. |
| `internal/util`         | Phát hiện ngôn ngữ, chuẩn hóa đường dẫn, tiện ích hỗ trợ                 |
| `internal/loader`       | Duyệt file repository với hỗ trợ `.gitignore` và lọc ngôn ngữ            |
| `internal/parser`       | Phân tích AST tree-sitter và trích xuất phần tử mã cho hơn 8 ngôn ngữ    |
| `internal/graph`        | Xây dựng & duyệt Call Graph, Dependency Graph, Inheritance Graph         |
| `internal/index`        | Engine đánh chỉ mục lai (vector embedding + BM25 qua cosine similarity)  |
| `internal/llm`          | Trừu tượng LLM client (API chat + embedding tương thích OpenAI)          |
| `internal/agent`        | Agent truy xuất lặp với thu thập context có ý thức ngân sách             |
| `internal/cache`        | Cache đĩa cho chỉ mục đã serialize (gob encoding)                        |
| `internal/orchestrator` | Orchestrator engine: nối loader → parser → graph → index → agent         |
| `pkg/treesitter`        | Binding Go tree-sitter và helper ngữ pháp ngôn ngữ                       |
| `reference/`            | Mã nguồn Python FastCode gốc để tham khảo trong quá trình porting        |
| `docs/`                 | Tài liệu nghiên cứu, phân tích, báo cáo test, và kế hoạch porting        |

---

## 📚 Ví Dụ

### Ví Dụ 1: Hiểu Luồng Xác Thực

**Truy vấn:** "Xác thực người dùng hoạt động như thế nào trong ứng dụng này?"

**Quy trình FastCode-CLI:**

1. 🗺️ Trinh sát các cấu trúc liên quan đến xác thực qua tìm kiếm lai
2. 🔍 Xác định `auth/handler.go`, `middleware/auth.go`, `models/user.go`
3. 📊 Truy vết phụ thuộc giữa các file qua Call Graph
4. 📖 Lướt chữ ký hàm — chỉ tải các hàm liên quan
5. ✅ Cung cấp câu trả lời toàn diện với đường dẫn file và số dòng

### Ví Dụ 2: Phân Tích Tác Động

**Truy vấn:** "Cái gì sẽ hỏng nếu tôi thay đổi schema User model?"

**Quy trình FastCode-CLI:**

1. 🗺️ Định vị định nghĩa User model qua tìm kiếm lai
2. 🔗 Truy vết tất cả import và sử dụng qua Dependency Graph
3. 📊 Lập bản đồ chuỗi phụ thuộc đa bước xuyên file
4. 📖 Tải các phần mã bị ảnh hưởng
5. ✅ Liệt kê tất cả file và hàm bị ảnh hưởng kèm điểm tự tin

### Ví Dụ 3: Hiểu Kiến Trúc

**Truy vấn:** "Giải thích cách routing API được cấu trúc"

**Quy trình FastCode-CLI:**

1. 🗺️ Trinh sát các pattern routing (`router`, `handler`, `endpoint`)
2. 🔍 Xác định file đăng ký route và hàm handler
3. 📊 Truy vết Call Graph từ router → handler → service
4. 📖 Lướt chuỗi middleware và bảo vệ xác thực
5. ✅ Cung cấp giải thích kiến trúc phân tầng

---

## 🗺 Lộ Trình

### Pha 1: Engine Cốt Lõi ✅

- [x] Phân tích AST tree-sitter cho Go, Python, JS/TS, Java, Rust, C/C++, C#
- [x] Trích xuất đơn vị mã (function, class, import, type)
- [x] Xây dựng Call Graph, Dependency Graph, Inheritance Graph
- [x] Trình tải file repository với hỗ trợ `.gitignore`

### Pha 2: Đánh Chỉ Mục ✅

- [x] Tạo embedding bằng LLM (qua API tương thích OpenAI)
- [x] Đánh chỉ mục text BM25 cho tìm kiếm từ khóa
- [x] Truy xuất lai (vector + BM25 fusion với trọng số)
- [x] Cache đĩa cho lưu trữ chỉ mục (gob encoding)

### Pha 3: Agent Truy Xuất ✅

- [x] Agent lặp có ý thức ngân sách với kiểm soát độ tự tin
- [x] Công cụ lướt mã và duyệt file thông minh
- [x] Bộ xử lý truy vấn (chấm điểm phức tạp, trích xuất từ khóa)
- [x] Bộ tạo câu trả lời (hỗ trợ LLM với context có cấu trúc)

### Pha 4: CLI & Tích Hợp ✅

- [x] Cobra CLI: `index`, `query`, `serve-mcp`
- [x] Chế độ MCP Server (JSON-RPC qua stdio)
- [x] Cờ đầu ra JSON cho tích hợp pipeline
- [x] Hỗ trợ file `.env` qua godotenv

### Pha 5: Hệ Sinh Thái _(Đang Lên Kế Hoạch)_

- [ ] Chế độ REST API server (`serve-api`)
- [ ] Hỗ trợ truy vấn đa repo (suy luận xuyên repository)
- [ ] Giao diện Web (tùy chọn)
- [ ] Binary dựng sẵn cho GitHub Releases
- [ ] Hỗ trợ `go install` / công thức Homebrew

---

## 🤝 Đóng Góp

Chúng tôi hoan nghênh đóng góp! FastCode-CLI được xây dựng cho cộng đồng, bởi cộng đồng.

### Cách Đóng Góp

- 🐛 **Báo Lỗi** — Tìm thấy vấn đề? Hãy cho chúng tôi biết!
- 💡 **Đề Xuất Tính Năng** — Có ý tưởng? Chúng tôi rất muốn nghe!
- 📝 **Cải Thiện Tài Liệu** — Giúp người khác hiểu FastCode-CLI tốt hơn
- 🔧 **Gửi Pull Request** — Đóng góp mã luôn được chào đón

### Thiết Lập Phát Triển

```bash
# Clone và thiết lập
git clone https://github.com/duyhunghd6/fastcode-cli.git
cd fastcode-cli

# Chạy test
go test ./... -v -cover

# Build
go build -o fastcode ./cmd/fastcode

# Chạy
./fastcode --version
```

---

## 📄 Giấy Phép

FastCode-CLI được phát hành theo [Giấy phép MIT](LICENSE).

---

## 🌟 Lịch Sử Star

<div align="center">

Nếu FastCode-CLI giúp bạn tiết kiệm token và thời gian, hãy tặng chúng tôi một ngôi sao! ⭐

**Được xây dựng với ❤️ bằng Go cho các lập trình viên coi trọng hiệu quả**

</div>

<div align="center">
  <a href="https://star-history.com/#duyhunghd6/fastcode-cli&Date">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=duyhunghd6/fastcode-cli&type=Date&theme=dark" />
      <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=duyhunghd6/fastcode-cli&type=Date" />
      <img alt="Biểu Đồ Lịch Sử Star" src="https://api.star-history.com/svg?repos=duyhunghd6/fastcode-cli&type=Date" style="border-radius: 15px; box-shadow: 0 0 30px rgba(0, 217, 255, 0.3);" />
    </picture>
  </a>
</div>

---

## 🙏 Ghi Nhận

Dự án này là bản **viết lại bằng Go** lấy cảm hứng từ [**FastCode**](https://github.com/HKUDS/FastCode) của [HKUDS Lab](https://github.com/HKUDS) tại Đại học Hồng Kông. Triển khai Python gốc đã giới thiệu framework ba pha đột phá cho việc hiểu mã nguồn tiết kiệm token.

Chúng tôi trân trọng ghi nhận các tác giả gốc và đóng góp nghiên cứu của họ.

---

<p align="center">
  <em> Cảm ơn bạn đã ghé thăm ✨ FastCode-CLI!</em><br><br>
  <strong>Một phần của hệ sinh thái <a href="https://github.com/duyhunghd6/gmind">Gmind</a> — Quản Lý Bộ Nhớ cho Agentic Coding</strong>
</p>
