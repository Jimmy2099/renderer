package generate_train_dataset

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	exts           string
	dir            string
	output         string
	excludeDirs    string
	includeHidden  bool
	headerFormat   string
	footerFormat   string
	maxChars       int
	maxTokens      int
	keepBlankLines bool
)

func init() {
	flag.StringVar(&exts, "exts", ".go,.c,.cpp,.h,.hpp,.cc", "Comma-separated list of file extensions\n(default: .go,.c,.cpp,.h,.hpp,.cc)")
	flag.StringVar(&dir, "dir", ".", "Root directory to scan\n(default: current directory)")
	flag.StringVar(&output, "output", "result.c", "Output filename\n(default: result.c)")
	flag.StringVar(&excludeDirs, "exclude-dirs", ".git,vendor,node_modules", "Directories to exclude\n(comma-separated, default: .git,vendor,node_modules)")
	flag.BoolVar(&includeHidden, "include-hidden", false, "Include hidden files (starting with .)\n(default: false)")
	flag.StringVar(&headerFormat, "header", "----- BEGIN %s -----", "Header format string\n(use %s for filename)")
	flag.StringVar(&footerFormat, "footer", "----- END %s -----", "Footer format string\n(use %s for filename)")
	flag.IntVar(&maxChars, "max-chars", 0, "Max characters per output file\n(0 = unlimited, default: 0)")
	flag.IntVar(&maxTokens, "max-tokens", 0, "Max tokens per output file\n(0 = unlimited, default: 0)")
	flag.BoolVar(&keepBlankLines, "keep-blank-lines", false, "Preserve blank lines\n(default: false)")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n  %s [options]\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  ./file-merger -max-chars 10000 -max-tokens 2500")
		fmt.Println("  ./file-merger -keep-blank-lines -output combined.txt")
		fmt.Println("  ./file-merger -exts .go,.c -dir ./src -exclude-dirs test")
	}
}

func PreProcess() {
	flag.Parse()

	extList := parseExtensions()
	excludeList := parseExcludeDirs()

	var files []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if contains(excludeList, filepath.Base(path)) {
				return filepath.SkipDir
			}
			return nil
		}

		if !includeHidden && strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		ext := filepath.Ext(path)
		if contains(extList, ext) {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		exitWithError("Directory traversal error: %v", err)
	}

	if len(files) == 0 {
		exitWithError("No matching files found")
	}

	sortFiles(files)

	if err := processFiles(files); err != nil {
		exitWithError("Processing failed: %v", err)
	}

	fmt.Printf("Successfully processed %d files\n", len(files))
}

func parseExtensions() []string {
	extList := strings.Split(exts, ",")
	for i := range extList {
		ext := strings.TrimSpace(extList[i])
		if ext != "" && !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		extList[i] = ext
	}
	return extList
}

func parseExcludeDirs() []string {
	return strings.Split(excludeDirs, ",")
}

func contains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}

func sortFiles(files []string) {
	sort.Slice(files, func(i, j int) bool {
		baseI, baseJ := filepath.Base(files[i]), filepath.Base(files[j])
		if baseI == baseJ {
			return files[i] < files[j]
		}
		return baseI < baseJ
	})
}

func processFiles(files []string) error {
	var (
		currentFile   *os.File
		currentChars  int
		currentTokens int
		fileIndex     = 1
	)

	defer func() {
		if currentFile != nil {
			currentFile.Close()
		}
	}()

	for _, file := range files {
		block, blockChars, blockTokens, err := buildFileBlock(file)
		if err != nil {
			fmt.Printf("Skipping %s: %v\n", file, err)
			continue
		}

		if isOversized(blockChars, blockTokens) {
			if err := handleOversizeFile(file, block, &fileIndex); err != nil {
				return err
			}
			continue
		}

		if needsNewFile(currentChars, currentTokens, blockChars, blockTokens) {
			if err := rotateFile(&currentFile, &currentChars, &currentTokens, &fileIndex); err != nil {
				return err
			}
		}

		if currentFile == nil {
			if err := createNewFile(&currentFile, fileIndex); err != nil {
				return err
			}
		}

		if _, err := currentFile.Write(block); err != nil {
			return fmt.Errorf("write error: %v", err)
		}

		currentChars += blockChars
		currentTokens += blockTokens
	}
	return nil
}

func buildFileBlock(path string) ([]byte, int, int, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, 0, 0, err
	}

	processed := processContent(content)
	// charCount := len(processed)
	tokenCount := len(strings.Fields(processed))

	filename := filepath.Base(path)
	relPath, _ := filepath.Rel(dir, path)

	header := fmt.Sprintf(headerFormat, filename)
	footer := fmt.Sprintf(footerFormat, filename)

	block := fmt.Sprintf("// Source: %s\n%s\n%s\n%s\n\n",
		relPath, header, processed, footer)
	return []byte(block), len(block), tokenCount, nil
}

func processContent(content []byte) string {
	var builder strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(string(content)))

	for scanner.Scan() {
		line := scanner.Text()
		if keepBlankLines || strings.TrimSpace(line) != "" {
			builder.WriteString(line + "\n")
		}
	}

	result := builder.String()
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}
	return result
}

func isOversized(chars, tokens int) bool {
	return (maxChars > 0 && chars > maxChars) ||
		(maxTokens > 0 && tokens > maxTokens)
}

func handleOversizeFile(path string, content []byte, index *int) error {
	fmt.Printf("File %s exceeds limits (chars: %d, tokens: %d), saving separately\n",
		path, len(content), len(strings.Fields(string(content))))

	newFile := generateFilename(*index)
	*index++
	return os.WriteFile(newFile, content, 0644)
}

func needsNewFile(currentChars, currentTokens, addChars, addTokens int) bool {
	if maxChars == 0 && maxTokens == 0 {
		return false
	}

	charExceeded := maxChars > 0 && (currentChars+addChars) > maxChars
	tokenExceeded := maxTokens > 0 && (currentTokens+addTokens) > maxTokens
	return charExceeded || tokenExceeded
}

func rotateFile(file **os.File, chars, tokens *int, index *int) error {
	if *file != nil {
		if err := (*file).Close(); err != nil {
			return err
		}
		*file = nil
	}
	*chars = 0
	*tokens = 0
	*index++
	return nil
}

func createNewFile(file **os.File, index int) error {
	name := generateFilename(index)
	f, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("create file failed: %v", err)
	}
	*file = f
	return nil
}

func generateFilename(index int) string {
	if index == 1 && maxChars == 0 && maxTokens == 0 {
		return output
	}

	ext := filepath.Ext(output)
	base := strings.TrimSuffix(output, ext)
	return fmt.Sprintf("%s-%d%s", base, index, ext)
}

func exitWithError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
