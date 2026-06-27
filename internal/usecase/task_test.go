package usecase

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/nathanieldk/task-manager/internal/dto"
	"github.com/nathanieldk/task-manager/internal/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// --- Mock Repositories ---

type mockTaskRepo struct {
	mu      sync.Mutex
	tasks   map[string]*entity.Task
	created int64
}

func newMockTaskRepo() *mockTaskRepo {
	return &mockTaskRepo{tasks: make(map[string]*entity.Task)}
}

func (m *mockTaskRepo) Create(ctx context.Context, task *entity.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks[task.ID] = task
	atomic.AddInt64(&m.created, 1)
	return nil
}

func (m *mockTaskRepo) FindByID(ctx context.Context, id string) (*entity.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	task, ok := m.tasks[id]
	if !ok {
		return nil, nil
	}
	return task, nil
}

func (m *mockTaskRepo) FindAll(ctx context.Context, creatorID string, status string, title string, limit, offset int) ([]*entity.Task, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*entity.Task
	for _, t := range m.tasks {
		if t.CreatorID == creatorID {
			result = append(result, t)
		}
	}
	return result, int64(len(result)), nil
}

func (m *mockTaskRepo) Update(ctx context.Context, task *entity.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskRepo) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tasks, id)
	return nil
}

func (m *mockTaskRepo) UpdateAssignee(ctx context.Context, taskID string, assigneeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if task, ok := m.tasks[taskID]; ok {
		task.AssigneeID = &assigneeID
	}
	return nil
}

func (m *mockTaskRepo) CreatedCount() int64 {
	return atomic.LoadInt64(&m.created)
}

// --- Mock Task Log Repository ---

type mockTaskLogRepo struct{}

func (m *mockTaskLogRepo) Create(ctx context.Context, log *entity.TaskLog) error {
	return nil
}

// --- Mock User Repository ---

type mockUserRepo struct {
	users map[string]*entity.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*entity.User)}
}

func (m *mockUserRepo) Create(ctx context.Context, user *entity.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*entity.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, nil
	}
	return user, nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, nil
}

// --- Mock Transaction Manager ---

type mockTxManager struct{}

func (m *mockTxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

// --- Mock Idempotency Usecase ---

type mockIdempotencyUsecase struct {
	mu    sync.Mutex
	store map[string]*entity.Task
	locks map[string]bool
}

func newMockIdempotencyUsecase() *mockIdempotencyUsecase {
	return &mockIdempotencyUsecase{
		store: make(map[string]*entity.Task),
		locks: make(map[string]bool),
	}
}

func (m *mockIdempotencyUsecase) Acquire(ctx context.Context, key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.locks[key] {
		return false, nil
	}
	m.locks[key] = true
	return true, nil
}

func (m *mockIdempotencyUsecase) Get(ctx context.Context, key string) (*entity.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	task, ok := m.store[key]
	if !ok {
		return nil, nil
	}
	return task, nil
}

func (m *mockIdempotencyUsecase) Store(ctx context.Context, key string, task *entity.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[key] = task
	return nil
}

// --- Helper ---

func newTestTaskUsecase() (*taskUsecase, *mockTaskRepo, *mockIdempotencyUsecase) {
	taskRepo := newMockTaskRepo()
	idempotency := newMockIdempotencyUsecase()
	logger, _ := zap.NewDevelopment()

	uc := &taskUsecase{
		taskRepo:    taskRepo,
		taskLogRepo: &mockTaskLogRepo{},
		userRepo:    newMockUserRepo(),
		txManager:   &mockTxManager{},
		idempotency: idempotency,
		logger:      logger,
	}

	return uc, taskRepo, idempotency
}

// ============================================================
// Test: Sequential Idempotency (no race condition)
// ============================================================
// 1. First request with a new Idempotency-Key → creates a task (201)
// 2. Second request with the same key → returns the cached task, no new creation

func TestIdempotency_Sequential(t *testing.T) {
	uc, taskRepo, _ := newTestTaskUsecase()
	ctx := context.Background()

	req := dto.CreateTaskRequest{
		Title:       "Test Task",
		Description: "Test Description",
	}
	idempotencyKey := "test-key-sequential"
	userID := "user-123"

	// First request — should create a new task
	task1, fromCache1, err := uc.Create(ctx, userID, req, idempotencyKey)
	require.NoError(t, err)
	assert.False(t, fromCache1, "First request should NOT be from cache")
	assert.NotNil(t, task1)
	assert.Equal(t, "Test Task", task1.Title)
	assert.Equal(t, int64(1), taskRepo.CreatedCount(), "Exactly 1 task should be created")

	// Second request with the same key — should return cached response
	task2, fromCache2, err := uc.Create(ctx, userID, req, idempotencyKey)
	require.NoError(t, err)
	assert.True(t, fromCache2, "Second request SHOULD be from cache")
	assert.NotNil(t, task2)
	assert.Equal(t, task1.ID, task2.ID, "Should return the same task ID")
	assert.Equal(t, task1.Title, task2.Title, "Should return the same task title")
	assert.Equal(t, int64(1), taskRepo.CreatedCount(), "Still exactly 1 task — no duplicate")
}

// ============================================================
// Test: Concurrent Idempotency (race condition)
// ============================================================
// Send N goroutines simultaneously with the same Idempotency-Key.
// Exactly 1 task must be created — no duplicates.

func TestIdempotency_ConcurrentDuplicate(t *testing.T) {
	uc, taskRepo, _ := newTestTaskUsecase()
	ctx := context.Background()

	req := dto.CreateTaskRequest{
		Title:       "Concurrent Task",
		Description: "Testing concurrent idempotency",
	}
	idempotencyKey := "test-key-concurrent"
	userID := "user-456"

	const goroutines = 20

	var wg sync.WaitGroup
	wg.Add(goroutines)

	results := make([]*entity.Task, goroutines)
	errors := make([]error, goroutines)

	// Launch all goroutines simultaneously
	start := make(chan struct{})
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			<-start // Wait for the starting signal
			task, _, err := uc.Create(ctx, userID, req, idempotencyKey)
			results[idx] = task
			errors[idx] = err
		}(i)
	}

	// Fire!
	close(start)
	wg.Wait()

	// Verify: exactly 1 task should have been created in the repository
	assert.Equal(t, int64(1), taskRepo.CreatedCount(),
		"Exactly 1 task should be created in the database despite %d concurrent requests", goroutines)

	// Verify: all successful responses should have the same task ID
	var firstTaskID string
	successCount := 0
	for i := 0; i < goroutines; i++ {
		if errors[i] != nil {
			// Some goroutines may get a "conflict" error while waiting — that's acceptable
			continue
		}
		successCount++
		if results[i] != nil {
			if firstTaskID == "" {
				firstTaskID = results[i].ID
			} else {
				assert.Equal(t, firstTaskID, results[i].ID,
					"All successful responses must return the same task ID")
			}
		}
	}

	assert.GreaterOrEqual(t, successCount, 1, "At least one request should succeed")
	t.Logf("Results: %d/%d succeeded, %d tasks created", successCount, goroutines, taskRepo.CreatedCount())
}

// ============================================================
// Test: Create without idempotency key works normally
// ============================================================

func TestCreate_WithoutIdempotencyKey(t *testing.T) {
	uc, taskRepo, _ := newTestTaskUsecase()
	ctx := context.Background()

	req := dto.CreateTaskRequest{
		Title:       "No Idempotency Task",
		Description: "Test without idempotency",
	}
	userID := "user-789"

	// Request without idempotency key
	task, fromCache, err := uc.Create(ctx, userID, req, "")
	require.NoError(t, err)
	assert.False(t, fromCache)
	assert.NotNil(t, task)
	assert.Equal(t, "No Idempotency Task", task.Title)
	assert.Equal(t, int64(1), taskRepo.CreatedCount())

	// Second request without idempotency key — should create a new task
	task2, fromCache2, err := uc.Create(ctx, userID, req, "")
	require.NoError(t, err)
	assert.False(t, fromCache2)
	assert.NotNil(t, task2)
	assert.NotEqual(t, task.ID, task2.ID, "Should create a different task")
	assert.Equal(t, int64(2), taskRepo.CreatedCount())
}

// ============================================================
// Test: Different idempotency keys create different tasks
// ============================================================

func TestIdempotency_DifferentKeys(t *testing.T) {
	uc, taskRepo, _ := newTestTaskUsecase()
	ctx := context.Background()

	req := dto.CreateTaskRequest{
		Title:       "Task A",
		Description: "First task",
	}
	userID := "user-101"

	task1, _, err := uc.Create(ctx, userID, req, "key-1")
	require.NoError(t, err)
	assert.NotNil(t, task1)

	req.Title = "Task B"
	task2, _, err := uc.Create(ctx, userID, req, "key-2")
	require.NoError(t, err)
	assert.NotNil(t, task2)

	assert.NotEqual(t, task1.ID, task2.ID, "Different keys should create different tasks")
	assert.Equal(t, int64(2), taskRepo.CreatedCount(), "Should have 2 tasks")
}

// ============================================================
// Test: GetByID — ownership check
// ============================================================

func TestGetByID_OwnershipCheck(t *testing.T) {
	uc, _, _ := newTestTaskUsecase()
	ctx := context.Background()

	// Create a task
	req := dto.CreateTaskRequest{Title: "Private Task", Description: "Secret"}
	task, _, err := uc.Create(ctx, "owner-user", req, "")
	require.NoError(t, err)

	// Owner can access
	result, err := uc.GetByID(ctx, "owner-user", task.ID)
	require.NoError(t, err)
	assert.Equal(t, task.ID, result.ID)

	// Non-owner cannot access
	_, err = uc.GetByID(ctx, "other-user", task.ID)
	assert.Error(t, err)
}

// ============================================================
// Test: List — pagination normalization
// ============================================================

func TestList_PaginationDefaults(t *testing.T) {
	uc, _, _ := newTestTaskUsecase()
	ctx := context.Background()

	// Create some tasks
	for i := 0; i < 5; i++ {
		req := dto.CreateTaskRequest{Title: "Task", Description: ""}
		_, _, err := uc.Create(ctx, "user-list", req, "")
		require.NoError(t, err)
	}

	// Query with invalid pagination — should normalize
	query := dto.TaskListQuery{Page: 0, Limit: -1}
	tasks, total, err := uc.List(ctx, "user-list", query)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, tasks, 5)
}

// ============================================================
// Test: Update — only creator can update
// ============================================================

func TestUpdate_OnlyCreator(t *testing.T) {
	uc, _, _ := newTestTaskUsecase()
	ctx := context.Background()

	task, _, err := uc.Create(ctx, "creator", dto.CreateTaskRequest{Title: "Original"}, "")
	require.NoError(t, err)

	// Creator can update
	newTitle := "Updated"
	updated, err := uc.Update(ctx, "creator", task.ID, dto.UpdateTaskRequest{Title: &newTitle})
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Title)

	// Non-creator cannot update
	_, err = uc.Update(ctx, "other", task.ID, dto.UpdateTaskRequest{Title: &newTitle})
	assert.Error(t, err)
}

// ============================================================
// Test: Delete — only creator can delete
// ============================================================

func TestDelete_OnlyCreator(t *testing.T) {
	uc, _, _ := newTestTaskUsecase()
	ctx := context.Background()

	task, _, err := uc.Create(ctx, "creator", dto.CreateTaskRequest{Title: "To Delete"}, "")
	require.NoError(t, err)

	// Non-creator cannot delete
	err = uc.Delete(ctx, "other", task.ID)
	assert.Error(t, err)

	// Creator can delete
	err = uc.Delete(ctx, "creator", task.ID)
	require.NoError(t, err)

	// Task should be gone
	_, err = uc.GetByID(ctx, "creator", task.ID)
	assert.Error(t, err)
}
