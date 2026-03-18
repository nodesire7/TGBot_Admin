package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tgbot/admin/config"
)

// MarketPlugin represents a plugin from the marketplace
type MarketPlugin struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Author      string                 `json:"author"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Icon        string                 `json:"icon"`
	Keywords    []string               `json:"keywords"`
	Hooks       []string               `json:"hooks"`
	Permissions []string               `json:"permissions"`
	Path        string                 `json:"path"`
	Commands    []map[string]string    `json:"commands,omitempty"`
	Manifest    map[string]interface{} `json:"manifest,omitempty"`
	Code        string                 `json:"code,omitempty"`
}

// MarketIndex represents the marketplace index
type MarketIndex struct {
	Version     string          `json:"version"`
	GeneratedAt string          `json:"generated_at"`
	Plugins     []MarketPlugin  `json:"plugins"`
	Categories  []MarketCategory `json:"categories"`
}

// MarketCategory represents a plugin category
type MarketCategory struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// MarketplaceRepository represents a plugin repository
type MarketplaceRepository struct {
	URL     string `json:"url"`
	Name    string `json:"name"`
	Branch  string `json:"branch"`
	Enabled bool   `json:"enabled"`
}

// Default repositories
var defaultRepositories = []MarketplaceRepository{
	{
		URL:     "https://raw.githubusercontent.com/nodesire7/TGBot_Plugins/main",
		Name:    "Official TGBot Plugins",
		Branch:  "main",
		Enabled: true,
	},
}

// cache for marketplace index
var marketCache *MarketIndex
var marketCacheTime time.Time
var cacheDuration = 5 * time.Minute

// GetMarketPlugins returns list of available plugins from marketplace
func GetMarketPlugins(c *gin.Context) {
	// Get query parameters
	category := c.Query("category")
	search := c.Query("search")
	repoURL := c.Query("repo")

	// Fetch index
	index, err := fetchMarketIndex(repoURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch marketplace index: " + err.Error()})
		return
	}

	// Filter plugins
	plugins := index.Plugins
	if category != "" {
		plugins = filterByCategory(plugins, category)
	}
	if search != "" {
		plugins = filterBySearch(plugins, search)
	}

	c.JSON(http.StatusOK, gin.H{
		"plugins":    plugins,
		"categories": index.Categories,
		"total":      len(plugins),
	})
}

// GetMarketPlugin returns details of a specific plugin
func GetMarketPlugin(c *gin.Context) {
	pluginID := c.Param("id")
	repoURL := c.Query("repo")

	index, err := fetchMarketIndex(repoURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch marketplace index"})
		return
	}

	// Find plugin
	var plugin *MarketPlugin
	for i := range index.Plugins {
		if index.Plugins[i].ID == pluginID {
			plugin = &index.Plugins[i]
			break
		}
	}

	if plugin == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found"})
		return
	}

	// Fetch full manifest
	manifest, err := fetchPluginManifest(repoURL, plugin.Path)
	if err == nil {
		plugin.Manifest = manifest
	}

	c.JSON(http.StatusOK, plugin)
}

// GetMarketPluginCode returns the source code of a plugin
func GetMarketPluginCode(c *gin.Context) {
	pluginID := c.Param("id")
	repoURL := c.Query("repo")

	index, err := fetchMarketIndex(repoURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch marketplace index"})
		return
	}

	// Find plugin
	var pluginPath string
	for _, p := range index.Plugins {
		if p.ID == pluginID {
			pluginPath = p.Path
			break
		}
	}

	if pluginPath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found"})
		return
	}

	// Fetch code
	code, err := fetchPluginCode(repoURL, pluginPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plugin code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plugin_id": pluginID,
		"code":      code,
	})
}

// InstallFromMarket installs a plugin from the marketplace
func InstallFromMarket(c *gin.Context) {
	var req struct {
		PluginID string `json:"plugin_id" binding:"required"`
		RepoURL  string `json:"repo_url"`
		BotID    int64  `json:"bot_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch index
	index, err := fetchMarketIndex(req.RepoURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch marketplace index"})
		return
	}

	// Find plugin
	var plugin *MarketPlugin
	for i := range index.Plugins {
		if index.Plugins[i].ID == req.PluginID {
			plugin = &index.Plugins[i]
			break
		}
	}

	if plugin == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found in marketplace"})
		return
	}

	// Fetch manifest and code
	manifest, err := fetchPluginManifest(req.RepoURL, plugin.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plugin manifest"})
		return
	}

	code, err := fetchPluginCode(req.RepoURL, plugin.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plugin code"})
		return
	}

	// Insert plugin into database
	var pluginDBID int64
	err = config.GetDB().QueryRow(context.Background(), `
		INSERT INTO plugins (plugin_id, name, version, author, description, main_file, manifest, source, is_system)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'github', false)
		ON CONFLICT (plugin_id) DO UPDATE SET
			name = EXCLUDED.name,
			version = EXCLUDED.version,
			author = EXCLUDED.author,
			description = EXCLUDED.description,
			main_file = EXCLUDED.main_file,
			manifest = EXCLUDED.manifest,
			source = EXCLUDED.source,
			updated_at = NOW()
		RETURNING id
	`, plugin.ID, plugin.Name, plugin.Version, plugin.Author, plugin.Description,
		code, manifest).Scan(&pluginDBID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save plugin: " + err.Error()})
		return
	}

	// If bot_id provided, enable plugin for that bot
	if req.BotID > 0 {
		_, err = config.GetDB().Exec(context.Background(), `
			INSERT INTO bot_plugins (bot_id, plugin_id, is_enabled, config)
			VALUES ($1, $2, true, '{}')
			ON CONFLICT (bot_id, plugin_id) DO UPDATE SET is_enabled = true
		`, req.BotID, plugin.ID)
		if err != nil {
			// Log but don't fail
			fmt.Printf("Warning: Failed to enable plugin for bot: %v\n", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Plugin installed successfully",
		"plugin_id": plugin.ID,
		"id":        pluginDBID,
	})
}

// GetMarketRepositories returns list of configured repositories
func GetMarketRepositories(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"repositories": defaultRepositories,
	})
}

// GetMarketCategories returns available categories
func GetMarketCategories(c *gin.Context) {
	index, err := fetchMarketIndex(c.Query("repo"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch marketplace index"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": index.Categories,
	})
}

// fetchMarketIndex fetches the marketplace index from remote or cache
func fetchMarketIndex(repoURL string) (*MarketIndex, error) {
	// Check cache
	if marketCache != nil && time.Since(marketCacheTime) < cacheDuration {
		return marketCache, nil
	}

	// Use default repo if not specified
	if repoURL == "" && len(defaultRepositories) > 0 {
		repoURL = defaultRepositories[0].URL
	}

	// Fetch index.json
	indexURL := repoURL + "/index.json"
	resp, err := http.Get(indexURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch index: status %d", resp.StatusCode)
	}

	var index MarketIndex
	if err := json.NewDecoder(resp.Body).Decode(&index); err != nil {
		return nil, err
	}

	// Update cache
	marketCache = &index
	marketCacheTime = time.Now()

	return &index, nil
}

// fetchPluginManifest fetches a plugin's manifest
func fetchPluginManifest(repoURL, pluginPath string) (map[string]interface{}, error) {
	if repoURL == "" && len(defaultRepositories) > 0 {
		repoURL = defaultRepositories[0].URL
	}

	manifestURL := repoURL + "/" + pluginPath + "/manifest.json"
	resp, err := http.Get(manifestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch manifest: status %d", resp.StatusCode)
	}

	var manifest map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, err
	}

	return manifest, nil
}

// fetchPluginCode fetches a plugin's main code
func fetchPluginCode(repoURL, pluginPath string) (string, error) {
	if repoURL == "" && len(defaultRepositories) > 0 {
		repoURL = defaultRepositories[0].URL
	}

	codeURL := repoURL + "/" + pluginPath + "/main.py"
	resp, err := http.Get(codeURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch code: status %d", resp.StatusCode)
	}

	// Read body as string
	buf := make([]byte, 1024*1024) // 1MB max
	n, err := resp.Body.Read(buf)
	if err != nil && err.Error() != "EOF" {
		return "", err
	}

	return string(buf[:n]), nil
}

// filterByCategory filters plugins by category
func filterByCategory(plugins []MarketPlugin, category string) []MarketPlugin {
	var result []MarketPlugin
	for _, p := range plugins {
		if p.Category == category {
			result = append(result, p)
		}
	}
	return result
}

// filterBySearch filters plugins by search term
func filterBySearch(plugins []MarketPlugin, search string) []MarketPlugin {
	var result []MarketPlugin
	searchLower := search
	for _, p := range plugins {
		// Search in name, description, keywords
		if containsIgnoreCase(p.Name, searchLower) ||
			containsIgnoreCase(p.Description, searchLower) ||
			containsString(p.Keywords, searchLower) {
			result = append(result, p)
		}
	}
	return result
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsIgnoreCase(s[1:], substr) || containsIgnoreCase(s[:len(s)-1], substr))
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
