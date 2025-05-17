# 使用頁面生成器

這個文檔提供了如何使用 easyPages 頁面生成器的指南。

## 創建新頁面

1. 在 `pages` 目錄中創建一個新的 `.md` 文件
2. 使用 Markdown 格式編寫內容
3. 運行 `go run main.go` 生成 HTML 頁面

## Markdown 語法範例

### 標題

```markdown
# 一級標題
## 二級標題
### 三級標題
```

### 列表

```markdown
- 項目 1
- 項目 2
- 項目 3

1. 第一項
2. 第二項
3. 第三項
```

### 強調

```markdown
*斜體* 或 _斜體_
**粗體** 或 __粗體__
```

### 連結和圖片

```markdown
[連結文字](https://example.com)
![圖片說明](image.jpg)
```

### 在 Markdown 中使用 HTML

您可以在 Markdown 文件中直接使用 HTML 標籤，例如：

```markdown
# 我的頁面

這是一些 Markdown 文本。

<div style="color: red;">
  這是一個紅色的 div 區塊
</div>

<script>
  // 這是一段 JavaScript 代碼
  document.addEventListener('DOMContentLoaded', function() {
    const title = document.querySelector('h1');
    title.style.color = 'blue';
  });
</script>
```

## 布局模板

布局模板 (`layout.html`) 定義了生成的 HTML 頁面的整體結構。您可以根據需要修改此文件以更改頁面的外觀。

模板中可用的變量：

- `{{.Title}}` - 頁面標題（基於檔案名）
- `{{.Content}}` - 從 Markdown 轉換後的 HTML 內容

例如，您可以添加 CSS 樣式表或 JavaScript 文件：

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
    <link rel="stylesheet" href="styles.css">
</head>
<body>
    <header>
        <h1>{{.Title}}</h1>
    </header>
    <main>
        {{.Content}}
    </main>    <footer>
        &copy; easyPages
    </footer>
    <script src="script.js"></script>
</body>
</html>
```
