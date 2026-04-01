package world

import (
	"math"
	"math/rand"
)

// Generator 世界生成器，根据种子生成体素地形。
type Generator struct {
	Seed int64
	rng  *rand.Rand
}

// NewGenerator 创建一个新的世界生成器。
func NewGenerator(seed int64) *Generator {
	return &Generator{
		Seed: seed,
		rng:  rand.New(rand.NewSource(seed)),
	}
}

// GenerateChunk 根据分块坐标生成地形。
// 使用简单的 Perlin 噪声模拟生成地形高度。
func (g *Generator) GenerateChunk(chunkX, chunkZ int) *Chunk {
	chunk := NewChunk(chunkX, chunkZ)

	for x := 0; x < ChunkSize; x++ {
		for z := 0; z < ChunkSize; z++ {
			worldX := float64(chunkX*ChunkSize + x)
			worldZ := float64(chunkZ*ChunkSize + z)

			// 简单的高度生成（使用正弦波模拟起伏地形）
			height := int(10 + 5*math.Sin(worldX*0.1)*math.Cos(worldZ*0.1))

			for y := 0; y < height; y++ {
				if y == 0 {
					chunk.SetBlock(x, y, z, BlockStone)
				} else if y < height-3 {
					chunk.SetBlock(x, y, z, BlockStone)
				} else if y < height-1 {
					chunk.SetBlock(x, y, z, BlockDirt)
				} else {
					chunk.SetBlock(x, y, z, BlockGrass)
				}
			}
		}
	}

	return chunk
}
