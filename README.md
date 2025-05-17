# easyPages 頁面生成器

這是一個使用 Golang 開發的簡單頁面生成器，可以將 Markdown 檔案轉換為 HTML 頁面並嵌入預定義的布局模板。

## 特點

- 將 Markdown 轉換為 HTML
- 自動套用版面設計 (layout.html)
- 批量處理所有 pages 資料夾中的 Markdown 檔案
- 輸出生成到 dist 資料夾
- 支援在 Markdown 中嵌入 HTML 代碼 (如 `<script>` 標籤)
- 正確處理 Markdown 語法元素 (如標題、列表、連結等)
- 使用 TOML 配置文件管理設置

## 安裝前置條件

在使用 easyPages 前，您需要安裝 Go 環境：

1. 前往 [Go 官方網站](https://golang.org/dl/) 下載適合您作業系統的安裝包
2. 按照官方指南完成安裝
3. 驗證安裝成功：
   ```bash
   go version
   ```
   應顯示已安裝的 Go 版本

## 使用方法

1. 將 Markdown 檔案放入 `pages` 資料夾
2. 根據需要修改 `layout.html` 布局模板
3. 根據需要修改 `config.toml` 配置文件
4. 執行程式：
   ```bash
   # 使用 Go 命令
   go run main.go
   ```
5. 生成的 HTML 檔案將保存在 `dist` 資料夾中

### TOML 配置

程式使用 `config.toml` 文件進行配置：

```toml
# 基本設置
[general]
pagesDir = "pages"     # Markdown 檔案所在的目錄
outputDir = "dist"     # 輸出 HTML 檔案的目錄
layoutFile = "layout.html"  # 佈局模板檔案
author = "easyPages Team"   # 頁面作者

# 監視模式設置
[watch]
enabled = false        # 是否啟用監視模式
delay = 5              # 監視間隔 (秒)
```

### 命令行參數

除了使用配置文件，程式仍支援命令行參數來覆蓋配置：

- `-pages` - Markdown 檔案所在的目錄 (預設: "pages")
- `-output` - 輸出 HTML 檔案的目錄 (預設: "dist")
- `-layout` - 佈局模板檔案 (預設: "layout.html")
- `-author` - 頁面作者 (預設: "Unknown")
- `-watch` - 啟用監視模式，檔案變更時自動重新生成 (預設: false)
- `-delay` - 監視模式下的檢查間隔秒數 (預設: 5)

範例：

```bash
# 使用不同的目錄和作者名稱
go run main.go -pages content -output public -author "easyPages Team"

# 啟用監視模式，每3秒檢查一次變更
go run main.go -watch -delay 3
```

## Markdown 中使用 HTML

您可以在 Markdown 文件中直接編寫 HTML 代碼，頁面生成器將保留這些 HTML 標籤。例如：

```markdown
# 我的頁面

這是一些 Markdown 文本。

<script>
    document.addEventListener('DOMContentLoaded', function() {
        const title = document.querySelector('h1');
        title.style.color = 'blue';
    });
</script>
```

## 布局模板

布局模板使用 Go 的 `html/template` 包，您可以在模板中使用以下變量：

- `{{.Title}}` - 頁面標題（基於檔案名）
- `{{.Content}}` - 頁面內容（從 Markdown 轉換後的 HTML）
- `{{.GeneratedAt}}` - 頁面生成時間
- `{{.Author}}` - 頁面作者（由命令行參數 `-author` 指定）

## 項目結構

```text
easyPages/
├── dist/           # 生成的 HTML 文件目錄
├── pages/          # Markdown 源文件目錄
├── layout.html     # HTML 模板文件
├── main.go         # 頁面生成器主程式
├── go.mod          # Go 模組定義
└── README.md       # 項目文檔
```

## 依賴項

- [github.com/russross/blackfriday/v2](https://github.com/russross/blackfriday) - Markdown 到 HTML 的轉換庫