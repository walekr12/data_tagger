package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/nfnt/resize"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	_ "image/gif"
	_ "image/png"
)

// App struct
type App struct {
	ctx          context.Context
	datasetPath  string
	items        []DatasetItem
	tagFrequency map[string]int
	thumbnailDir string
}

// DatasetItem represents a single image/video with its tags
type DatasetItem struct {
	ID            string   `json:"id"`
	MediaPath     string   `json:"mediaPath"`
	TxtPath       string   `json:"txtPath"`
	Tags          []string `json:"tags"`
	RawTags       string   `json:"rawTags"`
	ThumbnailPath string   `json:"thumbnailPath"`
	ThumbnailData string   `json:"thumbnailData"`
	IsVideo       bool     `json:"isVideo"`
	Selected      bool     `json:"selected"`
	Modified      bool     `json:"modified"`
}

// TagInfo represents a tag with its frequency
type TagInfo struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// ScanResult represents the result of scanning a folder
type ScanResult struct {
	Success     bool          `json:"success"`
	Message     string        `json:"message"`
	Items       []DatasetItem `json:"items"`
	Tags        []TagInfo     `json:"tags"`
	TotalItems  int           `json:"totalItems"`
	TotalImages int           `json:"totalImages"`
	TotalVideos int           `json:"totalVideos"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		items:        make([]DatasetItem, 0),
		tagFrequency: make(map[string]int),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// Create temp dir for thumbnails
	a.thumbnailDir = filepath.Join(os.TempDir(), "dataset-tagger-thumbnails")
	os.MkdirAll(a.thumbnailDir, 0755)
}

// SelectFolder opens a folder dialog
func (a *App) SelectFolder() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择数据集文件夹",
	})
	if err != nil {
		return "", err
	}
	return dir, nil
}

// ScanFolder scans the selected folder for image/video + txt pairs
func (a *App) ScanFolder(folderPath string) ScanResult {
	a.datasetPath = folderPath
	a.items = make([]DatasetItem, 0)
	a.tagFrequency = make(map[string]int)

	imageExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".bmp": true}
	videoExts := map[string]bool{".mp4": true, ".avi": true, ".mov": true, ".mkv": true, ".webm": true, ".flv": true}

	mediaFiles := make(map[string]string)
	txtFiles := make(map[string]string)

	totalImages := 0
	totalVideos := 0

	// Walk through directory
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		baseName := strings.TrimSuffix(filepath.Base(path), ext)
		dir := filepath.Dir(path)
		key := filepath.Join(dir, baseName)

		if ext == ".txt" {
			txtFiles[key] = path
		} else if imageExts[ext] {
			mediaFiles[key] = path
		} else if videoExts[ext] {
			mediaFiles[key] = path
		}

		return nil
	})

	if err != nil {
		return ScanResult{Success: false, Message: fmt.Sprintf("扫描失败: %v", err)}
	}

	// Match media files with txt files
	for key, mediaPath := range mediaFiles {
		txtPath, hasTxt := txtFiles[key]

		ext := strings.ToLower(filepath.Ext(mediaPath))
		isVideo := videoExts[ext]

		if isVideo {
			totalVideos++
		} else {
			totalImages++
		}

		item := DatasetItem{
			ID:        key,
			MediaPath: mediaPath,
			TxtPath:   txtPath,
			IsVideo:   isVideo,
			Tags:      []string{},
			RawTags:   "",
		}

		// Read tags from txt file
		if hasTxt {
			content, err := os.ReadFile(txtPath)
			if err == nil {
				item.RawTags = string(content)
				tags := a.parseTags(string(content))
				item.Tags = tags

				// Update frequency
				for _, tag := range tags {
					a.tagFrequency[tag]++
				}
			}
		}

		a.items = append(a.items, item)
	}

	// Sort tags by frequency - only include tags that appear in more than 1 file (common tags)
	tagInfos := make([]TagInfo, 0, len(a.tagFrequency))
	for tag, count := range a.tagFrequency {
		// 只统计出现在2个以上文件中的共同标签
		if count >= 2 {
			tagInfos = append(tagInfos, TagInfo{Tag: tag, Count: count})
		}
	}
	sort.Slice(tagInfos, func(i, j int) bool {
		return tagInfos[i].Count > tagInfos[j].Count
	})

	// 限制返回前100个最常见标签
	if len(tagInfos) > 100 {
		tagInfos = tagInfos[:100]
	}

	return ScanResult{
		Success:     true,
		Message:     fmt.Sprintf("成功扫描 %d 个文件", len(a.items)),
		Items:       a.items,
		Tags:        tagInfos,
		TotalItems:  len(a.items),
		TotalImages: totalImages,
		TotalVideos: totalVideos,
	}
}

// parseTags splits tag string into individual tags
func (a *App) parseTags(content string) []string {
	// Common separators: comma, newline
	content = strings.ReplaceAll(content, "\r\n", ",")
	content = strings.ReplaceAll(content, "\n", ",")

	parts := strings.Split(content, ",")
	tags := make([]string, 0)

	for _, p := range parts {
		tag := strings.TrimSpace(p)
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	return tags
}

// GetThumbnail generates and returns thumbnail as base64
func (a *App) GetThumbnail(mediaPath string, isVideo bool) string {
	cacheKey := base64.StdEncoding.EncodeToString([]byte(mediaPath))
	cachePath := filepath.Join(a.thumbnailDir, cacheKey+".jpg")

	// Check cache
	if _, err := os.Stat(cachePath); err == nil {
		data, err := os.ReadFile(cachePath)
		if err == nil {
			return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(data)
		}
	}

	var thumbData []byte

	if isVideo {
		thumbData = a.generateVideoThumbnail(mediaPath, cachePath)
	} else {
		thumbData = a.generateImageThumbnail(mediaPath, cachePath)
	}

	if thumbData == nil {
		return ""
	}

	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(thumbData)
}

// generateImageThumbnail creates a thumbnail for an image
func (a *App) generateImageThumbnail(imagePath, cachePath string) []byte {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil
	}

	// Resize to max 300px width
	thumb := resize.Thumbnail(300, 300, img, resize.Lanczos3)

	// Save to cache
	out, err := os.Create(cachePath)
	if err != nil {
		return nil
	}
	defer out.Close()

	jpeg.Encode(out, thumb, &jpeg.Options{Quality: 85})

	// Read back
	data, _ := os.ReadFile(cachePath)
	return data
}

// generateVideoThumbnail extracts middle frame from video using ffmpeg
func (a *App) generateVideoThumbnail(videoPath, cachePath string) []byte {
	// First get video duration
	durationCmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath)

	durationOutput, err := durationCmd.Output()
	if err != nil {
		// Fallback: extract at 1 second
		a.extractFrameAt(videoPath, cachePath, "1")
	} else {
		// Parse duration and get middle point
		durationStr := strings.TrimSpace(string(durationOutput))
		var duration float64
		fmt.Sscanf(durationStr, "%f", &duration)
		middleTime := fmt.Sprintf("%.2f", duration/2)
		a.extractFrameAt(videoPath, cachePath, middleTime)
	}

	data, _ := os.ReadFile(cachePath)
	return data
}

// extractFrameAt extracts a frame at specified time
func (a *App) extractFrameAt(videoPath, outputPath, timePos string) error {
	cmd := exec.Command("ffmpeg",
		"-ss", timePos,
		"-i", videoPath,
		"-vframes", "1",
		"-vf", "scale=300:-1",
		"-y",
		outputPath)

	return cmd.Run()
}

// SaveTags saves tags for a specific item
func (a *App) SaveTags(itemID string, tags string) error {
	for i, item := range a.items {
		if item.ID == itemID {
			txtPath := item.TxtPath
			if txtPath == "" {
				// Create new txt file
				txtPath = strings.TrimSuffix(item.MediaPath, filepath.Ext(item.MediaPath)) + ".txt"
				a.items[i].TxtPath = txtPath
			}

			err := os.WriteFile(txtPath, []byte(tags), 0644)
			if err != nil {
				return err
			}

			a.items[i].RawTags = tags
			a.items[i].Tags = a.parseTags(tags)
			return nil
		}
	}
	return fmt.Errorf("item not found: %s", itemID)
}

// SaveAllChanges saves all modified items
func (a *App) SaveAllChanges(items []DatasetItem) error {
	for _, item := range items {
		if item.Modified {
			err := a.SaveTags(item.ID, item.RawTags)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// BatchAddTag adds a tag to multiple items
func (a *App) BatchAddTag(itemIDs []string, tag string, position string) error {
	for _, id := range itemIDs {
		for i, item := range a.items {
			if item.ID == id {
				var newTags string
				if position == "prepend" {
					if item.RawTags == "" {
						newTags = tag
					} else {
						newTags = tag + ", " + item.RawTags
					}
				} else {
					if item.RawTags == "" {
						newTags = tag
					} else {
						newTags = item.RawTags + ", " + tag
					}
				}
				a.items[i].RawTags = newTags
				a.items[i].Tags = a.parseTags(newTags)
				a.items[i].Modified = true
			}
		}
	}
	return nil
}

// BatchRemoveTag removes a tag from multiple items
func (a *App) BatchRemoveTag(itemIDs []string, tag string, useRegex bool) error {
	for _, id := range itemIDs {
		for i, item := range a.items {
			if item.ID == id {
				var newTags []string
				for _, t := range item.Tags {
					if useRegex {
						matched, _ := regexp.MatchString(tag, t)
						if !matched {
							newTags = append(newTags, t)
						}
					} else {
						if t != tag {
							newTags = append(newTags, t)
						}
					}
				}
				a.items[i].RawTags = strings.Join(newTags, ", ")
				a.items[i].Tags = newTags
				a.items[i].Modified = true
			}
		}
	}
	return nil
}

// BatchReplaceTag replaces a tag in multiple items
func (a *App) BatchReplaceTag(itemIDs []string, oldTag string, newTag string, useRegex bool) error {
	for _, id := range itemIDs {
		for i, item := range a.items {
			if item.ID == id {
				var newTags []string
				for _, t := range item.Tags {
					if useRegex {
						re, err := regexp.Compile(oldTag)
						if err == nil {
							newTags = append(newTags, re.ReplaceAllString(t, newTag))
						} else {
							newTags = append(newTags, t)
						}
					} else {
						if t == oldTag {
							newTags = append(newTags, newTag)
						} else {
							newTags = append(newTags, t)
						}
					}
				}
				a.items[i].RawTags = strings.Join(newTags, ", ")
				a.items[i].Tags = newTags
				a.items[i].Modified = true
			}
		}
	}
	return nil
}

// GetItems returns all items
func (a *App) GetItems() []DatasetItem {
	return a.items
}

// FilterByTag returns items containing a specific tag
func (a *App) FilterByTag(tag string) []DatasetItem {
	result := make([]DatasetItem, 0)
	for _, item := range a.items {
		for _, t := range item.Tags {
			if t == tag {
				result = append(result, item)
				break
			}
		}
	}
	return result
}

// GetItemByID returns a single item by ID
func (a *App) GetItemByID(id string) *DatasetItem {
	for _, item := range a.items {
		if item.ID == id {
			return &item
		}
	}
	return nil
}

// ReadMediaFile reads and returns media file as base64 (for full preview)
func (a *App) ReadMediaFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	ext := strings.ToLower(filepath.Ext(path))
	mimeType := "image/jpeg"

	switch ext {
	case ".png":
		mimeType = "image/png"
	case ".gif":
		mimeType = "image/gif"
	case ".webp":
		mimeType = "image/webp"
	case ".mp4":
		mimeType = "video/mp4"
	case ".webm":
		mimeType = "video/webm"
	case ".mov":
		mimeType = "video/quicktime"
	}

	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data))
}

// ReadTextFile reads a text file and returns its content
func (a *App) ReadTextFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteTextFile writes content to a text file
func (a *App) WriteTextFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// GetPagedItems returns items for pagination
func (a *App) GetPagedItems(page int, pageSize int) ([]DatasetItem, int) {
	total := len(a.items)
	start := (page - 1) * pageSize
	if start >= total {
		return []DatasetItem{}, total
	}

	end := start + pageSize
	if end > total {
		end = total
	}

	return a.items[start:end], total
}

// OpenInExplorer opens the file location in explorer
func (a *App) OpenInExplorer(path string) error {
	dir := filepath.Dir(path)
	cmd := exec.Command("explorer", "/select,", path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		cmd = exec.Command("explorer", dir)
	}
	return cmd.Start()
}

// StreamFile streams large files efficiently
func (a *App) StreamFile(path string, offset int64, length int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	file.Seek(offset, io.SeekStart)
	data := make([]byte, length)
	n, err := file.Read(data)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return data[:n], nil
}
