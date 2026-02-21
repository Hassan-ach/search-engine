package internal

import "github.com/google/uuid"

type Data struct {
	Pages []*Page
	Idf   map[string]float64
}

type Page struct {
	URLID       uuid.UUID
	URL         string
	PRScore     float64
	Words       map[string]int
	GlobalScore float64
}

// // Mappers
type PageMapper struct {
	index     map[uuid.UUID]int
	nextIndex int
}

func (m PageMapper) GetIndex(id uuid.UUID) (int, bool) {
	idx, ok := m.index[id]

	return idx, ok
}

func (m *PageMapper) MapUUID(id uuid.UUID) {
	_, ok := m.index[id]
	if !ok {
		m.index[id] = m.nextIndex
		m.nextIndex++
	}
}

func NewPageMapper() *PageMapper {
	return &PageMapper{
		index:     make(map[uuid.UUID]int),
		nextIndex: 0,
	}
}

type WordMapper struct {
	index     map[string]int
	nextIndex int
}

func (m WordMapper) GetIndex(word string) (int, bool) {
	idx, ok := m.index[word]
	return idx, ok
}

func (m *WordMapper) MapWord(word string) {
	_, ok := m.index[word]
	if !ok {
		m.index[word] = m.nextIndex
		m.nextIndex++
	}
}

func NewWordMapper() *WordMapper {
	return &WordMapper{
		index:     make(map[string]int),
		nextIndex: 0,
	}
}
