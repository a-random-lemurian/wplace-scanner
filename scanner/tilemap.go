package scanner

import (
	"fmt"
	"image"
	"sync"
)

type TileMap struct {
	m  map[int]map[int]*WplaceTile
	mu sync.Mutex
}

func (t *TileMap) Add(w *WplaceTile) {
	t.mu.Lock()
	defer t.mu.Unlock()

	x, ok := t.m[w.Coords.X]
	if !ok {
		x = make(map[int]*WplaceTile, 0)
		t.m[w.Coords.X] = x
	}
	y, ok := t.m[w.Coords.Y]
	if !ok {
		x = make(map[int]*WplaceTile, 0)
		t.m[w.Coords.Y] = y
	}
	t.m[w.Coords.X][w.Coords.Y] = w
}

func (t *TileMap) StitchTiles() (image.Image, error) {
	return nil, fmt.Errorf("unimplemented")
}

type TileMapIteratorFunc func(x, y int, tile *WplaceTile)

func (t *TileMap) Iterate(it TileMapIteratorFunc) {

}

func NewTileMap() *TileMap {
	t := &TileMap{}
	t.m = make(map[int]map[int]*WplaceTile)
	return t
}
