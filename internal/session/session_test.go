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
