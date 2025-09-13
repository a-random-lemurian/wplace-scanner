package scanner

import (
	"image"
	"image/draw"
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

func (t *TileMap) StitchTiles(tileBbox *TileBoundingBox) (image.Image, error) {
	xSize := tileBbox.maxX - tileBbox.minX
	ySize := tileBbox.maxY - tileBbox.minY
	tileSize := 1000 // todo: do not hardcode this

	rectangle := image.Rect(0, 0, (xSize+1)*tileSize, (ySize+1)*tileSize)
	dst := image.NewRGBA(rectangle)
	xZero := 0
	yZero := 0

	for y := tileBbox.minY; y <= tileBbox.maxY; y++ {
		for x := tileBbox.minX; x <= tileBbox.maxX; x++ {
			img := t.m[x][y].Image
			pos := image.Pt(xZero*tileSize, yZero*tileSize)
			r := image.Rectangle{Min: pos, Max: pos.Add(img.Bounds().Size())}
			draw.Draw(dst, r, img, image.Point{}, draw.Over)
			xZero++
		}
		yZero++
		xZero = 0
	}

	return dst, nil
}

type TileMapIteratorFunc func(x, y int, tile *WplaceTile)

func (t *TileMap) Iterate(it TileMapIteratorFunc) {

}

func NewTileMap() *TileMap {
	t := &TileMap{}
	t.m = make(map[int]map[int]*WplaceTile)
	return t
}
