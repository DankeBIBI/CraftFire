package world

// PhysicsWorld 世界物理引擎封装，用于碰撞检测和重力模拟。
type PhysicsWorld struct {
	Gravity     float64
	MaxVelocity float64
}

// NewPhysicsWorld 创建一个新的世界物理实例。
func NewPhysicsWorld(gravity, maxVelocity float64) *PhysicsWorld {
	return &PhysicsWorld{
		Gravity:     gravity,
		MaxVelocity: maxVelocity,
	}
}

// CheckCollision 检查给定位置是否与方块发生碰撞。
// 简单的 AABB 碰撞检测。
func (pw *PhysicsWorld) CheckCollision(x, y, z float64, chunks map[string]*Chunk) bool {
	// 将浮点坐标转换为体素坐标
	bx := int(x)
	by := int(y)
	bz := int(z)

	// 检查该位置是否有实体方块
	chunkX := bx / ChunkSize
	chunkZ := bz / ChunkSize
	localX := bx % ChunkSize
	localZ := bz % ChunkSize

	key := chunkKey(chunkX, chunkZ)
	if chunk, exists := chunks[key]; exists {
		blockType := chunk.GetBlock(localX, by, localZ)
		return blockType != BlockAir && blockType != ""
	}

	return false
}

// chunkKey 生成分块的键值字符串。
func chunkKey(x, z int) string {
	return string(rune(x)) + "," + string(rune(z))
}
