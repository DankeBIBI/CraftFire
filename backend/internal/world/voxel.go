// Package world 提供体素世界相关的数据结构和操作。
package world

// BlockType 方块类型。
type BlockType string

const (
	BlockStone BlockType = "stone"
	BlockWood  BlockType = "wood"
	BlockGlass BlockType = "glass"
	BlockDirt  BlockType = "dirt"
	BlockGrass BlockType = "grass"
	BlockSand  BlockType = "sand"
	BlockWater BlockType = "water"
	BlockAir   BlockType = "air"
)

// Block 单个方块数据。
type Block struct {
	X        int       `json:"x"`
	Y        int       `json:"y"`
	Z        int       `json:"z"`
	Type     BlockType `json:"type"`
	Metadata int       `json:"metadata,omitempty"`
}

// VoxelCoord 体素坐标。
type VoxelCoord struct {
	X int
	Y int
	Z int
}

// IsValidBlockType 检查方块类型是否有效。
func IsValidBlockType(t string) bool {
	switch BlockType(t) {
	case BlockStone, BlockWood, BlockGlass, BlockDirt, BlockGrass, BlockSand, BlockWater:
		return true
	default:
		return false
	}
}
