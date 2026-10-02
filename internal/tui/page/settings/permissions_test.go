package settings

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/vesvai/vesvai/internal/core/config"
	"github.com/vesvai/vesvai/internal/core/event"
	"github.com/vesvai/vesvai/internal/core/logger"
	"github.com/vesvai/vesvai/internal/llm"
	"github.com/vesvai/vesvai/internal/tui/components"

	_ "github.com/vesvai/vesvai/internal/decision/providers"
)

type permDiscardHandler struct{}

func (permDiscardHandler) Write(logger.Record) error { return nil }
func (permDiscardHandler) Close() error              { return nil }

type permMockLLM struct {
	name   string
	models []llm.Model
}

func (p *permMockLLM) Name() string { return p.name }
func (p *permMockLLM) Chat(context.Context, *llm.Request) (*llm.Response, error) {
	return nil, errors.New("not used")
}
func (p *permMockLLM) ChatStream(context.Context, *llm.Request, llm.StreamHandler) error {
	return errors.New("not used")
}
func (p *permMockLLM) ListModels(context.Context) ([]llm.Model, error) {
	return p.models, nil
}

func newPermLLM(t *testing.T, provider, model string) *llm.Manager {
	t.Helper()
	bus := event.New()
	mgr := llm.NewManager(bus, logger.New(logger.LevelError, permDiscardHandler{}), nil)
	if err := mgr.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(mgr.Shutdown)
	llm.RegisterProvider(provider, func(cfg config.LLMConfig) (llm.Provider, error) {
		return &permMockLLM{name: cfg.Provider, models: []llm.Model{{ID: model}}}, nil
	})
	bus.Publish(event.TopicAppMounted, &config.Config{
		Providers: []config.LLMConfig{{Provider: provider}},
	})
	mgr.WaitUntilReady()
	return mgr
}

func permSettings(t *testing.T, cfg *config.Config, mgr *llm.Manager) *Settings {
	t.Helper()
	s := New(Deps{Config: cfg, LLM: mgr})
	s.tab = tabPermissions
	s.HandleKey(tcell.NewEventKey(tcell.KeyDown, 0, 0))
	return s
}

func TestPermissionsJudgeNavigation(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	s := permSettings(t, config.DefaultConfig(), nil)
	pt := s.permissions

	s.HandleKey(tcell.NewEventKey(tcell.KeyDown, 0, 0))
	if pt.focus != permFocusModel {
		t.Fatalf("focus = %v, want model", pt.focus)
	}
	s.HandleKey(tcell.NewEventKey(tcell.KeyDown, 0, 0))
	if pt.focus != permFocusThreshold {
		t.Fatalf("focus = %v, want threshold", pt.focus)
	}
	s.HandleKey(tcell.NewEventKey(tcell.KeyDown, 0, 0))
	if pt.focus != permFocusTools {
		t.Fatalf("focus = %v, want tools", pt.focus)
	}
	s.HandleKey(tcell.NewEventKey(tcell.KeyUp, 0, 0))
	if pt.focus != permFocusThreshold {
		t.Fatalf("focus = %v, want threshold", pt.focus)
	}
	s.HandleKey(tcell.NewEventKey(tcell.KeyUp, 0, 0))
	if pt.focus != permFocusModel {
		t.Fatalf("focus = %v, want model", pt.focus)
	}
	s.HandleKey(tcell.NewEventKey(tcell.KeyUp, 0, 0))
	if pt.focus != permFocusPresets {
		t.Fatalf("focus = %v, want presets", pt.focus)
	}
}

func TestPermissionsThresholdAdjustAndClamp(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	thr := 0.6
	cfg := config.DefaultConfig()
	cfg.Providers = []config.LLMConfig{{Provider: "openrouter", APIKey: "sk-or-test"}}
	cfg.Permission.JudgeProvider = "openrouter"
	cfg.Permission.JudgeModel = "typesafe/jev-1.13"
	cfg.Permission.JudgeThreshold = &thr
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}

	s := permSettings(t, cfg, nil)
	pt := s.permissions
	pt.loadIfNeeded()

	s.HandleKey(tcell.NewEventKey(tcell.KeyDown, 0, 0))
	s.HandleKey(tcell.NewEventKey(tcell.KeyDown, 0, 0))
	if pt.focus != permFocusThreshold {
		t.Fatalf("focus = %v, want threshold", pt.focus)
	}

	s.HandleKey(tcell.NewEventKey(tcell.KeyRight, 0, 0))
	if pt.judgeThreshold != 0.65 {
		t.Fatalf("threshold = %v, want 0.65", pt.judgeThreshold)
	}
	s.HandleKey(tcell.NewEventKey(tcell.KeyLeft, 0, 0))
	s.HandleKey(tcell.NewEventKey(tcell.KeyLeft, 0, 0))
	if pt.judgeThreshold != 0.55 {
		t.Fatalf("threshold = %v, want 0.55", pt.judgeThreshold)
	}

	// clamp low: 0.55 -> 0.05 in ten left presses
	for i := 0; i < 12; i++ {
		s.HandleKey(tcell.NewEventKey(tcell.KeyLeft, 0, 0))
	}
	if pt.judgeThreshold != 0.05 {
		t.Fatalf("threshold = %v, want clamped 0.05", pt.judgeThreshold)
	}

	// clamp high: 0.05 -> 1.0
	for i := 0; i < 25; i++ {
		s.HandleKey(tcell.NewEventKey(tcell.KeyRight, 0, 0))
	}
	if pt.judgeThreshold != 1 {
		t.Fatalf("threshold = %v, want clamped 1.0", pt.judgeThreshold)
	}

	got, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Permission.JudgeThreshold == nil || *got.Permission.JudgeThreshold != 1 {
		t.Fatalf("saved threshold = %+v, want 1", got.Permission.JudgeThreshold)
	}
}

func TestPermissionsThresholdDisabledForLLMJudge(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := config.DefaultConfig()
	cfg.Providers = []config.LLMConfig{{Provider: "llm-prov", APIKey: "sk-llm"}}
	cfg.Permission.JudgeProvider = "llm-prov"
	cfg.Permission.JudgeModel = "chat-model"
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}

	s := permSettings(t, cfg, nil)
	pt := s.permissions
	pt.loadIfNeeded()

	if pt.decisionEnabled() {
		t.Fatal("expected threshold disabled for a non-decision provider")
	}

	s.HandleKey(tcell.NewEventKey(tcell.KeyDown, 0, 0))
	s.HandleKey(tcell.NewEventKey(tcell.KeyDown, 0, 0))
	before := pt.judgeThreshold
	s.HandleKey(tcell.NewEventKey(tcell.KeyRight, 0, 0))
	s.HandleKey(tcell.NewEventKey(tcell.KeyLeft, 0, 0))
	if pt.judgeThreshold != before {
		t.Fatalf("threshold changed despite disabled row: %v -> %v", before, pt.judgeThreshold)
	}
}

func TestPermissionsThresholdDisplay(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	s := permSettings(t, config.DefaultConfig(), nil)
	pt := s.permissions
	pt.loadIfNeeded()
	if got := pt.thresholdDisplay(); got != fmt.Sprintf("%.2f", config.DefaultJudgeThreshold) {
		t.Errorf("display = %q, want default", got)
	}
	pt.judgeThreshold = 0.85
	if got := pt.thresholdDisplay(); got != "0.85" {
		t.Errorf("display = %q, want 0.85", got)
	}
}

func TestPermissionsEnterOpensJudgeModelPicker(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	s := permSettings(t, config.DefaultConfig(), nil)

	s.HandleKey(tcell.NewEventKey(tcell.KeyDown, 0, 0))
	s.HandleKey(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	if s.sub == nil {
		t.Fatal("Enter on Judge model row should open the picker")
	}
	s.HandleKey(tcell.NewEventKey(tcell.KeyEsc, 0, 0))
	if s.sub != nil {
		t.Error("Esc should close the picker")
	}
}

func TestPermissionsPickerListsDecisionsAndLLMs(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := config.DefaultConfig()
	cfg.Providers = []config.LLMConfig{
		{Provider: "openrouter", APIKey: "sk-or-test"},
		{Provider: "llm-prov", APIKey: "sk-llm"},
	}
	mgr := newPermLLM(t, "llm-prov", "chat-model")

	s := permSettings(t, cfg, mgr)
	pt := s.permissions
	pt.loadIfNeeded()

	pt.openJudgeModels()
	modal, ok := s.sub.(*listModal)
	if !ok {
		t.Fatalf("sub = %T, want *listModal", s.sub)
	}
	items := modal.list.Items()
	var labels []string
	for _, it := range items {
		labels = append(labels, it.Label)
	}
	if len(labels) < 4 {
		t.Fatalf("expected decisions + LLMs + headers, got %v", labels)
	}
	if labels[0] != "Decisions" {
		t.Errorf("first item = %q, want Decisions header", labels[0])
	}
	if items[1].Label != "typesafe/jev-1.13" || items[1].Detail != "openrouter · decision" {
		t.Errorf("decision item = %+v", items[1])
	}
	foundLLM := false
	for _, it := range items {
		if it.Label == "chat-model" && it.Detail == "llm-prov" {
			foundLLM = true
		}
	}
	if !foundLLM {
		t.Errorf("LLM model missing from picker: %v", labels)
	}
}

func TestPermissionsSelectDecisionSavesConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := config.DefaultConfig()
	cfg.Providers = []config.LLMConfig{{Provider: "openrouter", APIKey: "sk-or-test"}}
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}

	s := permSettings(t, cfg, nil)
	s.HandleKey(tcell.NewEventKey(tcell.KeyDown, 0, 0))
	s.HandleKey(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	if s.sub == nil {
		t.Fatal("picker did not open")
	}
	s.HandleKey(tcell.NewEventKey(tcell.KeyDown, 0, 0))
	s.HandleKey(tcell.NewEventKey(tcell.KeyEnter, 0, 0))

	got, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Permission == nil || got.Permission.JudgeProvider != "openrouter" {
		t.Fatalf("judge_provider = %+v, want openrouter", got.Permission)
	}
	if got.Permission.JudgeModel != "typesafe/jev-1.13" {
		t.Fatalf("judge_model = %q, want typesafe/jev-1.13", got.Permission.JudgeModel)
	}
	if s.permissions.judgeProvider != "openrouter" || s.permissions.judgeModel != "typesafe/jev-1.13" {
		t.Fatalf("tab state = %q/%q", s.permissions.judgeProvider, s.permissions.judgeModel)
	}
}

func TestPermissionsSelectLLMSavesConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := config.DefaultConfig()
	cfg.Providers = []config.LLMConfig{{Provider: "llm-prov", APIKey: "sk-llm"}}
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}
	mgr := newPermLLM(t, "llm-prov", "chat-model")

	s := permSettings(t, cfg, mgr)
	s.HandleKey(tcell.NewEventKey(tcell.KeyDown, 0, 0))
	s.HandleKey(tcell.NewEventKey(tcell.KeyEnter, 0, 0))

	// 0 = LLM models header, 1 = chat-model
	s.HandleKey(tcell.NewEventKey(tcell.KeyDown, 0, 0))
	s.HandleKey(tcell.NewEventKey(tcell.KeyEnter, 0, 0))

	got, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Permission.JudgeProvider != "llm-prov" || got.Permission.JudgeModel != "chat-model" {
		t.Fatalf("judge = %q/%q, want llm-prov/chat-model", got.Permission.JudgeProvider, got.Permission.JudgeModel)
	}
}

func TestPermissionsJudgeDisplay(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	s := permSettings(t, config.DefaultConfig(), nil)
	pt := s.permissions
	pt.loadIfNeeded()

	if got := pt.judgeDisplay(); got != "—" {
		t.Errorf("empty display = %q", got)
	}
	pt.judgeProvider, pt.judgeModel = "openrouter", ""
	if got := pt.judgeDisplay(); got != "openrouter (default)" {
		t.Errorf("provider-only display = %q", got)
	}
	pt.judgeModel = "typesafe/jev-1.13"
	if got := pt.judgeDisplay(); got != "openrouter/typesafe/jev-1.13" {
		t.Errorf("full display = %q", got)
	}
}

func TestPermissionsApplyLivePreservesJudgeConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	thr := 0.7
	cfg := config.DefaultConfig()
	cfg.Permission.JudgeProvider = "openrouter"
	cfg.Permission.JudgeModel = "typesafe/jev-1.13"
	cfg.Permission.JudgeThreshold = &thr
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}

	s := permSettings(t, cfg, nil)
	s.permissions.applyLive()

	got, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Permission.JudgeProvider != "openrouter" || got.Permission.JudgeModel != "typesafe/jev-1.13" {
		t.Fatalf("judge config lost: %+v", got.Permission)
	}
	if got.Permission.JudgeThreshold == nil || *got.Permission.JudgeThreshold != 0.7 {
		t.Fatalf("threshold lost: %+v", got.Permission.JudgeThreshold)
	}
	if got.Permission.Rules["bash"] != "semi-judge" {
		t.Errorf("rules not applied: %+v", got.Permission.Rules)
	}
}

var _ = components.ListItem{}
