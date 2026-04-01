package world

import "sync"

// ChunkSize 默认分块大小。
const ChunkSize = 16

// Chunk 世界分块，包含固定大小的方块网格。
type Chunk struct {
	X      int                                  `json:"x"`
	Z      int                                  `json:"z"`
	Blocks [ChunkSize][256][ChunkSize]BlockType `json:"-"`
	mu     sync.RWMutex
}

// NewChunk 创建一个新的空分块。
func NewChunk(x, z int) *Chunk {
	return &Chunk{X: x, Z: z}
}

// GetBlock 获取分块内指定位置的方块类型。
func (c *Chunk) GetBlock(localX, y, localZ int) BlockType {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isValidLocal(localX, y, localZ) {
		return BlockAir
	}
	return c.Blocks[localX][y][localZ]
}

// SetBlock 设置分块内指定位置的方块类型。
func (c *Chunk) SetBlock(localX, y, localZ int, blockType BlockType) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isValidLocal(localX, y, localZ) {
		c.Blocks[localX][y][localZ] = blockType
	}
}

// isValidLocal 检查分块内坐标是否有效。
func (c *Chunk) isValidLocal(x, y, z int) bool {
	return x >= 0 && x < ChunkSize &&
		y >= 0 && y < 256 &&
		z >= 0 && z < ChunkSize
}

// ToBlockList 将分块数据转为方块列表（用于序列化传输）。
func (c *Chunk) ToBlockList() []Block {
	c.mu.RLock()
	defer c.mu.RUnlock()

	blocks := make([]Block, 0)
	for x := 0; x < ChunkSize; x++ {
		for y := 0; y < 256; y++ {
			for z := 0; z < ChunkSize; z++ {
				if c.Blocks[x][y][z] != BlockAir && c.Blocks[x][y][z] != "" {
					blocks = append(blocks, Block{
						X:    c.X*ChunkSize + x,
						Y:    y,
						Z:    c.Z*ChunkSize + z,
						Type: c.Blocks[x][y][z],
					})
				}
			}
		}
	}
	return blocks
}
