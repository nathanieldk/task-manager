package handler

import (
	"math"
	"net/http"

	"github.com/nathanieldk/task-manager/internal/dto"
	"github.com/nathanieldk/task-manager/internal/pkg/apperror"
	"github.com/nathanieldk/task-manager/internal/pkg/response"
	"github.com/nathanieldk/task-manager/internal/usecase"

	"github.com/labstack/echo/v4"
)

type TaskHandler struct {
	taskUsecase usecase.TaskUsecase
}

func NewTaskHandler(taskUsecase usecase.TaskUsecase) *TaskHandler {
	return &TaskHandler{taskUsecase: taskUsecase}
}

// Create handles task creation.
// POST /tasks
func (h *TaskHandler) Create(c echo.Context) error {
	userID := c.Get("user_id").(string)
	idempotencyKey := c.Request().Header.Get("Idempotency-Key")

	var req dto.CreateTaskRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, apperror.ErrBadRequest("INVALID_REQUEST", "Invalid request body"))
	}

	if err := c.Validate(&req); err != nil {
		return response.Error(c, apperror.ErrValidation(err.Error()))
	}

	task, fromCache, err := h.taskUsecase.Create(c.Request().Context(), userID, req, idempotencyKey)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr)
		}
		return response.Error(c, apperror.ErrInternal(err))
	}

	taskResp := dto.TaskFromEntity(task)

	if fromCache {
		// Return the cached response — same status code as the original
		return response.Success(c, http.StatusCreated, taskResp)
	}

	return response.Success(c, http.StatusCreated, taskResp)
}

// List handles task listing with filters and pagination.
// GET /tasks
func (h *TaskHandler) List(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var query dto.TaskListQuery
	if err := c.Bind(&query); err != nil {
		return response.Error(c, apperror.ErrBadRequest("INVALID_REQUEST", "Invalid query parameters"))
	}

	tasks, total, err := h.taskUsecase.List(c.Request().Context(), userID, query)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr)
		}
		return response.Error(c, apperror.ErrInternal(err))
	}

	query.Normalize()

	var taskResponses []dto.TaskResponse
	for _, t := range tasks {
		taskResponses = append(taskResponses, dto.TaskFromEntity(t))
	}
	if taskResponses == nil {
		taskResponses = []dto.TaskResponse{}
	}

	pagination := dto.PaginationMeta{
		CurrentPage: query.Page,
		Limit:       query.Limit,
		TotalItems:  total,
		TotalPages:  int64(math.Ceil(float64(total) / float64(query.Limit))),
	}

	return response.Paginated(c, http.StatusOK, taskResponses, pagination)
}

// GetByID handles getting a single task.
// GET /tasks/:id
func (h *TaskHandler) GetByID(c echo.Context) error {
	userID := c.Get("user_id").(string)
	taskID := c.Param("id")

	task, err := h.taskUsecase.GetByID(c.Request().Context(), userID, taskID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr)
		}
		return response.Error(c, apperror.ErrInternal(err))
	}

	return response.Success(c, http.StatusOK, dto.TaskFromEntity(task))
}

// Update handles task updates.
// PUT /tasks/:id
func (h *TaskHandler) Update(c echo.Context) error {
	userID := c.Get("user_id").(string)
	taskID := c.Param("id")

	var req dto.UpdateTaskRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, apperror.ErrBadRequest("INVALID_REQUEST", "Invalid request body"))
	}

	task, err := h.taskUsecase.Update(c.Request().Context(), userID, taskID, req)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr)
		}
		return response.Error(c, apperror.ErrInternal(err))
	}

	return response.Success(c, http.StatusOK, dto.TaskFromEntity(task))
}

// Delete handles task deletion.
// DELETE /tasks/:id
func (h *TaskHandler) Delete(c echo.Context) error {
	userID := c.Get("user_id").(string)
	taskID := c.Param("id")

	if err := h.taskUsecase.Delete(c.Request().Context(), userID, taskID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr)
		}
		return response.Error(c, apperror.ErrInternal(err))
	}

	return response.Success(c, http.StatusOK, map[string]string{"message": "Task deleted successfully"})
}

// Assign handles task assignment to a team member.
// POST /tasks/:id/assign
func (h *TaskHandler) Assign(c echo.Context) error {
	userID := c.Get("user_id").(string)
	taskID := c.Param("id")

	var req dto.AssignTaskRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, apperror.ErrBadRequest("INVALID_REQUEST", "Invalid request body"))
	}

	if err := c.Validate(&req); err != nil {
		return response.Error(c, apperror.ErrValidation(err.Error()))
	}

	if err := h.taskUsecase.Assign(c.Request().Context(), userID, taskID, req); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr)
		}
		return response.Error(c, apperror.ErrInternal(err))
	}

	return response.Success(c, http.StatusOK, map[string]string{"message": "Task assigned successfully"})
}
