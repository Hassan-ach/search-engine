package handlers

import (
	"fmt"
	"net/http"

	"query-engine/internal/service"

	"github.com/labstack/echo/v5"
	"query-engine/view/result"
)

type SearchingHandler struct {
	Ranker  service.Ranker
	Speller service.Speller
}

func NewSearchHandler(ranker service.Ranker, speller service.Speller) *SearchingHandler {
	return &SearchingHandler{
		ranker,
		speller,
	}
}

func (h SearchingHandler) Handle(c *echo.Context) error {
	query := c.QueryParam("query")

	fmt.Printf("query: %s\n", query)

	sugs := h.Speller.GetSuggestions(query)

	pages, err := h.Ranker.Rank(sugs)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprint("err: %w", err))
	}

	return render(c, result.Show(pages))
}
