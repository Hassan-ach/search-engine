package handlers

import (
	"github.com/labstack/echo/v5"
	"github.com/Hassan-ach/boogle/services/engine/view/page/home"
)

type HomeHandler struct{}

func (h HomeHandler) Handle(c *echo.Context) error {
	return render(c, home.Show())
}
