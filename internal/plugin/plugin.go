package plugin

import (
	"context"
	"sort"
	"sync"
)

// Info is public metadata for admin UI.
type Info struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	Builtin     bool   `json:"builtin"`
}

// Plugin is a Go-native extension. Xiuno PHP plugins are NOT supported.
type Plugin interface {
	Meta() Info
	// OnEnable is called when the plugin is turned on (may be at boot).
	OnEnable(ctx context.Context) error
	// OnDisable is called when the plugin is turned off.
	OnDisable(ctx context.Context) error
}

// Hook names used by core. Plugins register handlers via Manager.On.
const (
	HookPageRender   = "page.render"   // *PageRenderEvent
	HookThreadViewed = "thread.viewed" // *ThreadViewedEvent
	HookReplyCreated = "reply.created" // *ReplyCreatedEvent
	HookBoot         = "app.boot"      // nil
)

// PageRenderEvent is fired before writing a public HTML page.
type PageRenderEvent struct {
	Template string
	// ExtraCSS / ExtraJS are appended by plugins (public URL paths or inline markers).
	ExtraCSS []string
	ExtraJS  []string
	// FooterHTML is injected before </body> via layout if supported.
	FooterHTML string
	Params     map[string]any
}

type ThreadViewedEvent struct {
	Tid uint
	Uid uint
}

type ReplyCreatedEvent struct {
	Tid uint
	Pid uint
	Uid uint
}

type Handler func(ctx context.Context, event any) error

type Manager struct {
	mu       sync.RWMutex
	plugins  map[string]Plugin
	enabled  map[string]bool
	handlers map[string][]namedHandler
}

type namedHandler struct {
	pluginID string
	fn       Handler
}

var global = &Manager{
	plugins:  map[string]Plugin{},
	enabled:  map[string]bool{},
	handlers: map[string][]namedHandler{},
}

func Global() *Manager { return global }

// Register adds a builtin plugin (call from init of builtin packages).
func Register(p Plugin) {
	meta := p.Meta()
	global.mu.Lock()
	defer global.mu.Unlock()
	global.plugins[meta.ID] = p
	// default disabled unless set later
	if _, ok := global.enabled[meta.ID]; !ok {
		global.enabled[meta.ID] = false
	}
}

// On registers a hook handler for a plugin id.
func On(pluginID, hook string, fn Handler) {
	global.mu.Lock()
	defer global.mu.Unlock()
	global.handlers[hook] = append(global.handlers[hook], namedHandler{pluginID: pluginID, fn: fn})
}

// Init applies persisted enable flags and enables plugins.
func Init(ctx context.Context, enabled map[string]bool) error {
	m := global
	m.mu.Lock()
	if enabled != nil {
		for id, on := range enabled {
			m.enabled[id] = on
		}
	}
	// copy enable set for unlock
	toEnable := make([]Plugin, 0)
	for id, p := range m.plugins {
		if m.enabled[id] {
			toEnable = append(toEnable, p)
		}
	}
	m.mu.Unlock()
	for _, p := range toEnable {
		if err := p.OnEnable(ctx); err != nil {
			return err
		}
	}
	return m.Fire(ctx, HookBoot, nil)
}

func (m *Manager) List() []Info {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Info, 0, len(m.plugins))
	for id, p := range m.plugins {
		info := p.Meta()
		info.ID = id
		info.Enabled = m.enabled[id]
		info.Builtin = true
		out = append(out, info)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (m *Manager) EnabledMap() map[string]bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]bool, len(m.enabled))
	for k, v := range m.enabled {
		out[k] = v
	}
	return out
}

func (m *Manager) SetEnabled(ctx context.Context, id string, on bool) error {
	m.mu.Lock()
	p, ok := m.plugins[id]
	if !ok {
		m.mu.Unlock()
		return errNotFound
	}
	prev := m.enabled[id]
	m.enabled[id] = on
	m.mu.Unlock()
	if prev == on {
		return nil
	}
	if on {
		return p.OnEnable(ctx)
	}
	return p.OnDisable(ctx)
}

// Fire runs handlers for enabled plugins only.
func (m *Manager) Fire(ctx context.Context, hook string, event any) error {
	m.mu.RLock()
	list := append([]namedHandler(nil), m.handlers[hook]...)
	enabled := make(map[string]bool, len(m.enabled))
	for k, v := range m.enabled {
		enabled[k] = v
	}
	m.mu.RUnlock()
	for _, h := range list {
		if !enabled[h.pluginID] {
			continue
		}
		if err := h.fn(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

type notFoundError string

func (e notFoundError) Error() string { return string(e) }

const errNotFound = notFoundError("plugin not found")
