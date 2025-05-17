package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/russross/blackfriday/v2"
)

// 頁面結構
type Page struct {
	Title       string
	Content     template.HTML
	GeneratedAt string // 添加生成日期和時間
	Author      string // 添加作者資訊
}

// TOML 配置結構
type TomlConfig struct {
	General struct {
		PagesDir   string `toml:"pagesDir"`
		OutputDir  string `toml:"outputDir"`
		LayoutFile string `toml:"layoutFile"`
		Author     string `toml:"author"`
	} `toml:"general"`
	Watch struct {
		Enabled bool `toml:"enabled"`
		Delay   int  `toml:"delay"`
	} `toml:"watch"`
}

// 應用程式配置
type Config struct {
	PagesDir   string // Markdown 檔案所在的目錄
	OutputDir  string // 輸出 HTML 檔案的目錄
	LayoutFile string // 佈局模板檔案
	Author     string // 頁面作者
	WatchMode  bool   // 是否啟用監視模式
	WatchDelay int    // 監視模式下的延遲秒數
}

// 處理 Markdown 並保留 HTML 標籤
func processMarkdownWithHTML(mdContent []byte) []byte {
	// 使用正則表達式找出所有 HTML 標籤
	// 匹配HTML標籤
	htmlTagsRegex := regexp.MustCompile(`<[a-zA-Z][^>]*>[\s\S]*?</[a-zA-Z][^>]*>|<[a-zA-Z][^>/]*/>|<[a-zA-Z][^>]*>`)

	// 找到所有 HTML 標籤
	htmlTags := htmlTagsRegex.FindAllIndex(mdContent, -1)
	// 設置 Blackfriday 擴展選項，確保列表和其他 Markdown 元素正確轉換
	extensions := blackfriday.CommonExtensions | blackfriday.HardLineBreak

	// 如果沒有 HTML 標籤，直接使用 Blackfriday 處理整個內容
	if len(htmlTags) == 0 {
		return blackfriday.Run(mdContent, blackfriday.WithExtensions(extensions))
	}

	// 將內容拆分為 HTML 標籤和 Markdown
	var processedContent bytes.Buffer
	lastIndex := 0

	for _, tagIndices := range htmlTags {
		start, end := tagIndices[0], tagIndices[1]

		// 處理 HTML 標籤前的 Markdown 內容
		if start > lastIndex {
			markdownPart := mdContent[lastIndex:start]
			// 使用擴展選項處理 Markdown
			htmlPart := blackfriday.Run(markdownPart, blackfriday.WithExtensions(extensions))
			processedContent.Write(htmlPart)
		}

		// 保留原始 HTML 標籤
		htmlTag := mdContent[start:end]
		processedContent.Write(htmlTag)

		lastIndex = end
	}

	// 處理最後一個 HTML 標籤後的 Markdown 內容
	if lastIndex < len(mdContent) {
		markdownPart := mdContent[lastIndex:]
		// 使用擴展選項處理最後的 Markdown 部分
		htmlPart := blackfriday.Run(markdownPart, blackfriday.WithExtensions(extensions))
		processedContent.Write(htmlPart)
	}

	return processedContent.Bytes()
}

// 處理 Markdown 檔案並生成 HTML
func processMarkdownFiles(config Config) error {
	// 讀取布局模板
	layoutContent, err := os.ReadFile(config.LayoutFile)
	if err != nil {
		return fmt.Errorf("讀取布局文件失敗: %v", err)
	}

	// 解析模板
	tmpl, err := template.New("layout").Parse(string(layoutContent))
	if err != nil {
		return fmt.Errorf("解析模板失敗: %v", err)
	}

	// 處理所有 Markdown 文件
	err = filepath.Walk(config.PagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳過目錄
		if info.IsDir() {
			return nil
		}

		// 只處理 Markdown 文件
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			return nil
		}

		fmt.Printf("處理文件: %s\n", path)
		// 讀取 Markdown 內容
		mdContent, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("讀取文件 %s 失敗: %v\n", path, err)
			return nil
		}

		// 預處理 Markdown 內容，確保列表格式正確
		mdStr := string(mdContent)
		// 確保破折號列表項前後有足夠的空行
		mdStr = strings.ReplaceAll(mdStr, "\n- ", "\n\n- ")
		mdContent = []byte(mdStr)

		// 處理 Markdown 並保留 HTML 標籤
		htmlContent := processMarkdownWithHTML(mdContent)

		// 從文件名提取標題
		baseName := filepath.Base(path)
		title := strings.TrimSuffix(baseName, filepath.Ext(baseName))
		// 首字母大寫 (使用 strings.ToUpper 替代已棄用的 strings.Title)
		if len(title) > 0 {
			title = strings.ToUpper(title[:1]) + title[1:]
		}

		// 創建頁面
		page := Page{
			Title:       title,
			Content:     template.HTML(htmlContent),
			GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
			Author:      config.Author,
		}

		// 渲染 HTML
		var htmlBuffer bytes.Buffer
		if err := tmpl.Execute(&htmlBuffer, page); err != nil {
			fmt.Printf("渲染模板失敗: %v\n", err)
			return nil
		}
		// 寫入輸出文件
		outputPath := filepath.Join(config.OutputDir, title+".html")
		err = os.WriteFile(outputPath, htmlBuffer.Bytes(), 0644)
		if err != nil {
			fmt.Printf("寫入輸出文件失敗: %v\n", err)
			return nil
		}

		fmt.Printf("生成了HTML文件: %s\n", outputPath)
		return nil
	})

	return err
}

// 複製非 Markdown 檔案到輸出目錄
func copyNonMarkdownFiles(config Config) error {
	fmt.Println("複製非 Markdown 檔案到輸出目錄...")

	// 處理所有非 Markdown 文件
	err := filepath.Walk(config.PagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳過目錄
		if info.IsDir() {
			return nil
		}

		// 跳過 Markdown 文件
		if strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			return nil
		}

		// 計算相對路徑
		relPath, err := filepath.Rel(config.PagesDir, path)
		if err != nil {
			fmt.Printf("計算相對路徑失敗: %v\n", err)
			return nil
		}

		// 構建目標路徑
		destPath := filepath.Join(config.OutputDir, relPath)

		// 確保目標目錄存在
		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			fmt.Printf("創建目標目錄失敗: %v\n", err)
			return nil
		}

		// 讀取源文件
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("讀取文件 %s 失敗: %v\n", path, err)
			return nil
		}

		// 寫入目標文件
		if err := os.WriteFile(destPath, data, 0644); err != nil {
			fmt.Printf("寫入文件 %s 失敗: %v\n", destPath, err)
			return nil
		}

		fmt.Printf("複製了文件: %s -> %s\n", path, destPath)
		return nil
	})

	return err
}

// 監視文件變更並在變更時重新生成頁面
func watchForChanges(config Config) {
	fmt.Printf("監視模式已啟用，監視間隔: %d 秒\n", config.WatchDelay)
	fmt.Println("按 Ctrl+C 停止...")

	// 儲存檔案的最後修改時間
	fileModTimes := make(map[string]time.Time)
	// 初始化檔案修改時間
	updateModTimes := func() {
		err := filepath.Walk(config.PagesDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				// 記錄所有檔案的修改時間，包括非 Markdown 檔案
				fileModTimes[path] = info.ModTime()
			}
			return nil
		})
		if err != nil {
			fmt.Printf("遍歷文件失敗: %v\n", err)
		}

		// 也檢查佈局文件
		if info, err := os.Stat(config.LayoutFile); err == nil {
			fileModTimes[config.LayoutFile] = info.ModTime()
		}
	}

	// 初始化
	updateModTimes()

	// 每隔指定的秒數檢查一次檔案變更
	for {
		time.Sleep(time.Duration(config.WatchDelay) * time.Second)

		filesChanged := false
		// 檢查所有檔案
		err := filepath.Walk(config.PagesDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				lastMod, exists := fileModTimes[path]
				if !exists || info.ModTime().After(lastMod) {
					filesChanged = true
					fileModTimes[path] = info.ModTime()
					fmt.Printf("檢測到文件變更: %s\n", path)
				}
			}
			return nil
		})
		if err != nil {
			fmt.Printf("遍歷文件失敗: %v\n", err)
		}

		// 檢查佈局文件
		if info, err := os.Stat(config.LayoutFile); err == nil {
			lastMod, exists := fileModTimes[config.LayoutFile]
			if !exists || info.ModTime().After(lastMod) {
				filesChanged = true
				fileModTimes[config.LayoutFile] = info.ModTime()
				fmt.Printf("檢測到佈局文件變更: %s\n", config.LayoutFile)
			}
		}
		// 如果有檔案變更，重新生成頁面
		if filesChanged {
			fmt.Println("重新生成所有頁面...")
			// 先處理 Markdown 檔案
			if err := processMarkdownFiles(config); err != nil {
				fmt.Printf("處理 Markdown 檔案失敗: %v\n", err)
			}

			// 然後複製非 Markdown 檔案
			if err := copyNonMarkdownFiles(config); err != nil {
				fmt.Printf("複製非 Markdown 檔案失敗: %v\n", err)
			} else {
				fmt.Println("所有頁面已更新!")
			}
		}
	}
}

func main() {
	// 定義命令行參數 (保留以便兼容舊模式)
	configPath := flag.String("config", "config.toml", "配置文件路徑")
	var cliConfig Config
	flag.StringVar(&cliConfig.PagesDir, "pages", "", "Markdown 檔案所在的目錄")
	flag.StringVar(&cliConfig.OutputDir, "output", "", "輸出 HTML 檔案的目錄")
	flag.StringVar(&cliConfig.LayoutFile, "layout", "", "佈局模板檔案")
	flag.StringVar(&cliConfig.Author, "author", "", "頁面作者")
	flag.BoolVar(&cliConfig.WatchMode, "watch", false, "是否啟用監視模式")
	flag.IntVar(&cliConfig.WatchDelay, "delay", 0, "監視模式下的延遲秒數")
	flag.Parse()

	// 創建默認配置
	config := Config{
		PagesDir:   "pages",
		OutputDir:  "dist",
		LayoutFile: "layout.html",
		Author:     "easyPages Team",
		WatchMode:  false,
		WatchDelay: 5,
	}

	// 嘗試從 TOML 文件加載配置
	if _, err := os.Stat(*configPath); err == nil {
		var tomlConfig TomlConfig
		if _, err := toml.DecodeFile(*configPath, &tomlConfig); err != nil {
			fmt.Printf("無法解析配置文件: %v\n", err)
		} else {
			// 從 TOML 配置更新配置
			config.PagesDir = tomlConfig.General.PagesDir
			config.OutputDir = tomlConfig.General.OutputDir
			config.LayoutFile = tomlConfig.General.LayoutFile
			config.Author = tomlConfig.General.Author
			config.WatchMode = tomlConfig.Watch.Enabled
			config.WatchDelay = tomlConfig.Watch.Delay
			fmt.Println("已從配置文件載入設定")
		}
	} else {
		fmt.Printf("配置文件不存在，使用默認配置: %s\n", *configPath)
	}

	// 命令行參數覆蓋配置文件設置
	if cliConfig.PagesDir != "" {
		config.PagesDir = cliConfig.PagesDir
	}

	if cliConfig.OutputDir != "" {
		config.OutputDir = cliConfig.OutputDir
	}

	if cliConfig.LayoutFile != "" {
		config.LayoutFile = cliConfig.LayoutFile
	}

	if cliConfig.Author != "" {
		config.Author = cliConfig.Author
	}

	if flag.Lookup("watch").Value.String() == "true" {
		config.WatchMode = true
	}

	if cliConfig.WatchDelay > 0 {
		config.WatchDelay = cliConfig.WatchDelay
	}

	// 顯示當前配置
	fmt.Println("頁面目錄:", config.PagesDir)
	fmt.Println("輸出目錄:", config.OutputDir)
	fmt.Println("佈局檔案:", config.LayoutFile)
	fmt.Println("作者:", config.Author)
	fmt.Println("監視模式:", config.WatchMode)
	if config.WatchMode {
		fmt.Println("監視間隔:", config.WatchDelay, "秒")
	}
	fmt.Println()

	// 確保輸出目錄存在
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		fmt.Printf("創建輸出目錄失敗: %v\n", err)
		return
	}
	if config.WatchMode {
		// 先處理一次
		if err := processMarkdownFiles(config); err != nil {
			fmt.Printf("處理 Markdown 檔案失敗: %v\n", err)
			return
		}
		if err := copyNonMarkdownFiles(config); err != nil {
			fmt.Printf("複製非 Markdown 檔案失敗: %v\n", err)
			return
		}
		// 然後啟動監視
		watchForChanges(config)
	} else {
		if err := processMarkdownFiles(config); err != nil {
			fmt.Printf("處理 Markdown 檔案失敗: %v\n", err)
			return
		}
		if err := copyNonMarkdownFiles(config); err != nil {
			fmt.Printf("複製非 Markdown 檔案失敗: %v\n", err)
			return
		}
		fmt.Println("完成! 所有頁面都已生成。")
	}
}
