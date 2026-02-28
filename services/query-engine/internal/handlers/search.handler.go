package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"query-engine/internal/service"

	"query-engine/view/page/result"

	"github.com/labstack/echo/v5"
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
	pageNum := getPageNum(c)

	fmt.Printf("query: %s\n", query)

	sugs := h.Speller.GetSuggestions(query)

	pages, err := h.Ranker.Rank(sugs, pageNum)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprint("err: %w", err))
	}

	return render(c, result.Show(pages))
}

func getPageNum(c *echo.Context) int {
	page := c.QueryParam("page")
	pageNum := 1

	if page != "" {
		v, err := strconv.Atoi(page)
		if err == nil {
			pageNum = v
		}
	}
	return pageNum
}
