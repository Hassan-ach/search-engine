package handlers

import (
	"github.com/labstack/echo/v5"
	"query-engine/view/home"
)

type HomeHandler struct{}

func (h HomeHandler) Handle(c *echo.Context) error {
	return render(c, home.Show())
}
