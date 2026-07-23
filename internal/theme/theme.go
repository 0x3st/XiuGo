package theme

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	KVActiveTheme = "xiugo_theme"
	DefaultID     = "default"
)

// Info describes a theme package under resource/themes/<id>/.
type Info struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Author      string `json:"author"`
	Description string `json:"description"`
	// CSS is URL path served from /themes/<id>/...
	CSS []string `json:"css"`
	// Dir absolute path on disk
	Dir string `json:"-"`
}

type Manager struct {
	mu       sync.RWMutex
	root     string
	themes   map[string]Info
	activeID string
}

var global = &Manager{themes: map[string]Info{}}

func Global() *Manager { return global }

func (m *Manager) Root() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.root
}

// Init scans resource/themes and sets active theme from kv loader.
func Init(ctx context.Context, themesRoot string, loadActive func(context.Context) (string, error)) error {
	if themesRoot == "" {
		themesRoot = "resource/themes"
	}
	if abs, err := filepath.Abs(themesRoot); err == nil {
		themesRoot = abs
	}
	m := global
	m.mu.Lock()
	defer m.mu.Unlock()
	m.root = themesRoot
	m.themes = map[string]Info{}

	entries, err := os.ReadDir(themesRoot)
	if err != nil {
		if os.IsNotExist(err) {
			m.activeID = DefaultID
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		id := entry.Name()
		dir := filepath.Join(themesRoot, id)
		info := Info{
			ID: id, Name: id, Version: "1.0.0", Dir: dir,
			CSS: []string{},
		}
		if raw, err := os.ReadFile(filepath.Join(dir, "theme.json")); err == nil {
			_ = json.Unmarshal(raw, &info)
			info.ID = id
			info.Dir = dir
		}
		// default css path if theme.css exists
		if _, err := os.Stat(filepath.Join(dir, "css", "theme.css")); err == nil {
			if len(info.CSS) == 0 {
				info.CSS = []string{"css/theme.css"}
			}
		}
		m.themes[id] = info
	}
	// ensure default exists in map even if empty folder
	if _, ok := m.themes[DefaultID]; !ok {
		m.themes[DefaultID] = Info{ID: DefaultID, Name: "Default", Version: "5.0.0", Description: "Built-in Xiuno-like default", Dir: filepath.Join(themesRoot, DefaultID)}
	}
	active := DefaultID
	if loadActive != nil {
		if id, err := loadActive(ctx); err == nil && strings.TrimSpace(id) != "" {
			active = strings.TrimSpace(id)
		}
	}
	if _, ok := m.themes[active]; !ok {
		active = DefaultID
	}
	m.activeID = active
	g.Log().Infof(ctx, "theme: loaded %d theme(s), active=%s", len(m.themes), active)
	return nil
}

func (m *Manager) List() []Info {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Info, 0, len(m.themes))
	for _, t := range m.themes {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (m *Manager) Active() Info {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if t, ok := m.themes[m.activeID]; ok {
		return t
	}
	return m.themes[DefaultID]
}

func (m *Manager) ActiveID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeID
}

func (m *Manager) SetActive(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	id = strings.TrimSpace(id)
	if id == "" {
		id = DefaultID
	}
	if _, ok := m.themes[id]; !ok {
		return os.ErrNotExist
	}
	m.activeID = id
	return nil
}

// Stylesheets returns public URLs for the active theme CSS.
func (m *Manager) Stylesheets() []string {
	t := m.Active()
	if t.ID == "" || t.ID == DefaultID && len(t.CSS) == 0 {
		// default may ship empty css (uses stock bootstrap-bbs)
		urls := make([]string, 0, len(t.CSS))
		for _, rel := range t.CSS {
			urls = append(urls, "/themes/"+t.ID+"/"+strings.TrimPrefix(rel, "/"))
		}
		return urls
	}
	urls := make([]string, 0, len(t.CSS))
	for _, rel := range t.CSS {
		urls = append(urls, "/themes/"+t.ID+"/"+strings.TrimPrefix(rel, "/"))
	}
	return urls
}

// ResolveTemplate returns an absolute template path if the active theme
// overrides it under templates/<rel>, otherwise returns rel unchanged.
// rel is like "pages/home.html".
func (m *Manager) ResolveTemplate(rel string) string {
	rel = strings.TrimPrefix(filepath.ToSlash(rel), "/")
	t := m.Active()
	if t.ID == "" || t.ID == DefaultID || t.Dir == "" {
		// Still allow default theme overrides if present.
	}
	if t.Dir == "" {
		return rel
	}
	candidate := filepath.Join(t.Dir, "templates", filepath.FromSlash(rel))
	if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
		return candidate
	}
	return rel
}

// HasTemplateOverride reports whether active theme ships templates/<rel>.
func (m *Manager) HasTemplateOverride(rel string) bool {
	resolved := m.ResolveTemplate(rel)
	return resolved != rel && filepath.IsAbs(resolved)
}
