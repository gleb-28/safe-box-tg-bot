package session

import (
	"strings"
	"sync"
	"testing"

	"safeboxtgbot/models"
)

type testLogger struct {
	mu     sync.Mutex
	debugs []string
	infos  []string
	errors []string
}

func (l *testLogger) Debug(message string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.debugs = append(l.debugs, message)
}

func (l *testLogger) Info(message string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.infos = append(l.infos, message)
}

func (l *testLogger) Error(message string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.errors = append(l.errors, message)
}

func (l *testLogger) countDebugContains(substr string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	count := 0
	for _, msg := range l.debugs {
		if strings.Contains(msg, substr) {
			count++
		}
	}
	return count
}

func (l *testLogger) countInfoContains(substr string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	count := 0
	for _, msg := range l.infos {
		if strings.Contains(msg, substr) {
			count++
		}
	}
	return count
}

func TestStoreGet_CacheMissCreatesSession(t *testing.T) {
	logger := &testLogger{}
	store := NewStore(logger)

	sess := store.Get(42)
	if sess == nil || sess.User == nil {
		t.Fatalf("expected session and user to be initialized")
	}

	if got := logger.countDebugContains("cache miss"); got != 1 {
		t.Fatalf("expected 1 cache miss log, got %d", got)
	}
	if got := logger.countInfoContains("New session created"); got != 1 {
		t.Fatalf("expected 1 session created log, got %d", got)
	}
	if got := logger.countDebugContains("cache hit"); got != 0 {
		t.Fatalf("expected 0 cache hit logs, got %d", got)
	}
}

func TestStoreGet_CacheHitOnSecondCall(t *testing.T) {
	logger := &testLogger{}
	store := NewStore(logger)

	_ = store.Get(7)
	_ = store.Get(7)

	if got := logger.countDebugContains("cache miss"); got != 1 {
		t.Fatalf("expected 1 cache miss log, got %d", got)
	}
	if got := logger.countDebugContains("cache hit"); got != 1 {
		t.Fatalf("expected 1 cache hit log, got %d", got)
	}
}

func TestStoreUpdateUser_SetsLoadedFlag(t *testing.T) {
	logger := &testLogger{}
	store := NewStore(logger)

	_ = store.Get(9)
	store.UpdateUser(9, &models.User{TelegramID: 9})

	if !store.IsUserLoaded(9) {
		t.Fatalf("expected UserIsLoaded to be true")
	}
	if got := logger.countDebugContains("Session user updated"); got != 1 {
		t.Fatalf("expected 1 session user updated log, got %d", got)
	}
}

func TestStoreIsUserLoaded_DefaultFalse(t *testing.T) {
	logger := &testLogger{}
	store := NewStore(logger)

	if store.IsUserLoaded(100) {
		t.Fatalf("expected UserIsLoaded to be false for new session")
	}
}

func TestStoreUpdate_DoesNothingWithoutSession(t *testing.T) {
	logger := &testLogger{}
	store := NewStore(logger)

	updated := false
	store.Update(5, func(s *Session) {
		updated = true
	})

	if updated {
		t.Fatalf("expected Update to not run without existing session")
	}
}

func TestStoreUpdate_ModifiesSession(t *testing.T) {
	logger := &testLogger{}
	store := NewStore(logger)

	sess := store.Get(11)
	store.Update(11, func(s *Session) {
		s.User = &models.User{TelegramID: 123}
	})

	if sess.User.TelegramID != 123 {
		t.Fatalf("expected Update to modify session user")
	}
}

func TestStoreGetUser_ReturnsUpdatedUser(t *testing.T) {
	logger := &testLogger{}
	store := NewStore(logger)

	_ = store.Get(21)
	store.UpdateUser(21, &models.User{TelegramID: 555})

	user := store.GetUser(21)
	if user.TelegramID != 555 {
		t.Fatalf("expected GetUser to return updated user")
	}
}

func TestStoreItems_DefaultState(t *testing.T) {
	logger := &testLogger{}
	store := NewStore(logger)

	sess := store.Get(1)
	if sess.Items.ItemsLoaded {
		t.Fatalf("expected ItemsLoaded to be false for new session")
	}
	if len(sess.Items.ItemList) != 0 {
		t.Fatalf("expected ItemList to be empty for new session")
	}
	if sess.Items.EditingItemName != "" {
		t.Fatalf("expected EditingItemName to be empty for new session")
	}
}

func TestStoreSetItemList_MarksLoaded(t *testing.T) {
	logger := &testLogger{}
	store := NewStore(logger)

	_ = store.Get(2)
	items := []models.Item{
		{Name: "item-a"},
		{Name: "item-b"},
	}
	store.SetItemList(2, items)

	if !store.IsItemsLoaded(2) {
		t.Fatalf("expected ItemsLoaded to be true after SetItemList")
	}
	got := store.GetItemList(2)
	if len(got) != len(items) {
		t.Fatalf("expected ItemList length %d, got %d", len(items), len(got))
	}
	if got[0].Name != items[0].Name {
		t.Fatalf("expected first item name %q, got %q", items[0].Name, got[0].Name)
	}
}

func TestStoreEditingItemName(t *testing.T) {
	logger := &testLogger{}
	store := NewStore(logger)

	_ = store.Get(3)
	store.SetEditingItemName(3, "item-99")
	if got := store.GetEditingItemName(3); got != "item-99" {
		t.Fatalf("expected EditingItemName \"item-99\", got %q", got)
	}
	store.ClearEditingItemName(3)
	if got := store.GetEditingItemName(3); got != "" {
		t.Fatalf("expected EditingItemName empty after Clear, got %q", got)
	}
}

func TestStoreUpdateUser_DoesNotAlterItemsState(t *testing.T) {
	logger := &testLogger{}
	store := NewStore(logger)

	_ = store.Get(4)
	items := []models.Item{{Name: "item-a"}}
	store.SetItemList(4, items)
	store.SetEditingItemName(4, "item-7")

	store.UpdateUser(4, &models.User{TelegramID: 4})

	if !store.IsItemsLoaded(4) {
		t.Fatalf("expected ItemsLoaded to remain true after UpdateUser")
	}
	got := store.GetItemList(4)
	if len(got) != 1 || got[0].Name != "item-a" {
		t.Fatalf("expected ItemList to remain unchanged after UpdateUser")
	}
	if gotName := store.GetEditingItemName(4); gotName != "item-7" {
		t.Fatalf("expected EditingItemName to remain \"item-7\" after UpdateUser, got %q", gotName)
	}
}

func TestStoreUpdateUser_DoesNotMarkItemsLoaded(t *testing.T) {
	logger := &testLogger{}
	store := NewStore(logger)

	_ = store.Get(5)
	store.UpdateUser(5, &models.User{TelegramID: 5})

	if store.IsItemsLoaded(5) {
		t.Fatalf("expected ItemsLoaded to remain false after UpdateUser")
	}
}
