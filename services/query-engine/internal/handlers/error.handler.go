package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"query-engine/internal/apperror"
	errorpage "query-engine/view/page/error"

	"github.com/labstack/echo/v5"
)

func HandleError(c *echo.Context, err error) {
	code, message, internalErr := classifyError(err)

	if internalErr != nil {
		c.Logger().
			Error(fmt.Sprintf("[%s] %s → %v", c.Request().Method, c.Request().URL.Path, internalErr.Error()))
	}

	c.Response().WriteHeader(code)
	err = render(c, errorpage.ErrorPage(code, message))

	if err != nil {
		c.Logger().Error("Failed to render error page", "error", err)
	}
}

func classifyError(err error) (int, string, error) {
	if errors.Is(err, echo.ErrNotFound) {
		return http.StatusNotFound, "page not found", nil
	}

	if errors.Is(err, echo.ErrMethodNotAllowed) {
		return http.StatusMethodNotAllowed, "method not allowed", nil
	}

	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		return appErr.Code, appErr.Message, appErr.Err
	}

	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) {
		message := httpErr.Message
		if message == "" {
			message = http.StatusText(httpErr.Code)
		}
		return httpErr.Code, message, nil
	}

	return http.StatusInternalServerError, "internal server error", err
}
