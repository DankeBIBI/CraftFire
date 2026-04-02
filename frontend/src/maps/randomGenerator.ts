import type { BlockData, BlockType } from '@/types/game'

interface Hill {
  cx: number
  cz: number
  radius: number
  height: number
}

function clamp(value: number, min: number, max: number): number {
  return Math.max(min, Math.min(max, value))
}

function randInt(min: number, max: number): number {
  return Math.floor(Math.random() * (max - min + 1)) + min
}

function chance(probability: number): boolean {
  return Math.random() < probability
}

export function generateRandomMapBlocks(): BlockData[] {
  const blocks: BlockData[] = []
  const occupied = new Set<string>()

  const xMin = -40
  const xMax = 39
  const zMin = -40
  const zMax = 39

  const hills: Hill[] = Array.from({ length: randInt(8, 14) }, () => ({
    cx: randInt(xMin + 6, xMax - 6),
    cz: randInt(zMin + 6, zMax - 6),
    radius: randInt(8, 18),
    height: randInt(3, 8),
  }))

  const heightMap = new Map<string, number>()

  const setBlock = (x: number, y: number, z: number, type: BlockType) => {
    const key = `${x},${y},${z}`
    if (occupied.has(key)) return
    occupied.add(key)
    blocks.push({ x, y, z, type })
  }

  const getHeight = (x: number, z: number): number => {
    const key = `${x},${z}`
    const cached = heightMap.get(key)
    if (cached !== undefined) return cached

    const wave = Math.sin(x * 0.14) * 1.8 + Math.cos(z * 0.16) * 1.4
    const micro = Math.sin((x + z) * 0.23) * 0.8
    let height = 3 + Math.floor(wave + micro)

    for (const hill of hills) {
      const dx = x - hill.cx
      const dz = z - hill.cz
      const distance = Math.sqrt(dx * dx + dz * dz)
      if (distance > hill.radius) continue
      const ratio = 1 - distance / hill.radius
      height += Math.floor(hill.height * ratio * ratio)
    }

    if (chance(0.07)) {
      height += randInt(-1, 2)
    }

    const finalHeight = clamp(height, 2, 15)
    heightMap.set(key, finalHeight)
    return finalHeight
  }

  for (let x = xMin; x <= xMax; x++) {
    for (let z = zMin; z <= zMax; z++) {
      const h = getHeight(x, z)
      for (let y = 0; y < h; y++) {
        const type: BlockType = y < h - 2 ? 'stone' : 'dirt'
        setBlock(x, y, z, type)
      }
      setBlock(x, h, z, 'grass')

      if (h <= 3 && chance(0.05)) {
        setBlock(x, h + 1, z, 'water')
      }
    }
  }

  const spawnPlateauX = randInt(-4, 4)
  const spawnPlateauZ = randInt(-4, 4)
  for (let x = spawnPlateauX - 3; x <= spawnPlateauX + 3; x++) {
    for (let z = spawnPlateauZ - 3; z <= spawnPlateauZ + 3; z++) {
      const h = getHeight(x, z)
      setBlock(x, h + 1, z, 'grass')
    }
  }

  const tryPlaceTree = (x: number, z: number) => {
    const baseH = getHeight(x, z)
    if (baseH < 4 || baseH > 12) return

    const neighbors = [
      getHeight(x + 1, z),
      getHeight(x - 1, z),
      getHeight(x, z + 1),
      getHeight(x, z - 1),
    ]
    if (neighbors.some((n) => Math.abs(n - baseH) > 2)) return

    const trunkHeight = randInt(4, 7)
    for (let y = 1; y <= trunkHeight; y++) {
      setBlock(x, baseH + y, z, 'wood')
    }

    const canopyY = baseH + trunkHeight
    for (let dx = -2; dx <= 2; dx++) {
      for (let dz = -2; dz <= 2; dz++) {
        for (let dy = -1; dy <= 1; dy++) {
          const manhattan = Math.abs(dx) + Math.abs(dz) + Math.abs(dy)
          if (manhattan > 4) continue
          if (chance(0.18)) continue
          setBlock(x + dx, canopyY + dy, z + dz, 'grass')
        }
      }
    }
  }

  for (let i = 0; i < 170; i++) {
    const x = randInt(xMin + 3, xMax - 3)
    const z = randInt(zMin + 3, zMax - 3)
    if (!chance(0.33)) continue
    tryPlaceTree(x, z)
  }

  const placeBuilding = (cx: number, cz: number, w: number, d: number, h: number) => {
    const x1 = cx - Math.floor(w / 2)
    const x2 = cx + Math.floor(w / 2)
    const z1 = cz - Math.floor(d / 2)
    const z2 = cz + Math.floor(d / 2)

    const floorY = Math.max(
      2,
      Math.floor((
        getHeight(x1, z1) +
        getHeight(x1, z2) +
        getHeight(x2, z1) +
        getHeight(x2, z2)
      ) / 4),
    )

    for (let x = x1; x <= x2; x++) {
      for (let z = z1; z <= z2; z++) {
        setBlock(x, floorY, z, chance(0.5) ? 'stoneDark' : 'sandDark')
      }
    }

    const wallType: BlockType = chance(0.5) ? 'stone' : 'wood'
    const roofType: BlockType = chance(0.5) ? 'stoneDark' : 'crate'

    const doorX = randInt(x1 + 1, x2 - 1)
    const doorOnNorth = chance(0.5)

    for (let y = floorY + 1; y <= floorY + h; y++) {
      for (let x = x1; x <= x2; x++) {
        if (!(doorOnNorth && x >= doorX && x <= doorX + 1 && y <= floorY + 2)) {
          setBlock(x, y, z1, wallType)
        }
        if (!(!doorOnNorth && x >= doorX && x <= doorX + 1 && y <= floorY + 2)) {
          setBlock(x, y, z2, wallType)
        }
      }
      for (let z = z1; z <= z2; z++) {
        setBlock(x1, y, z, wallType)
        setBlock(x2, y, z, wallType)
      }
    }

    for (let x = x1; x <= x2; x++) {
      for (let z = z1; z <= z2; z++) {
        setBlock(x, floorY + h + 1, z, roofType)
      }
    }

    const windowY = floorY + 2
    setBlock(x1, windowY, cz, 'glass')
    setBlock(x2, windowY, cz, 'glass')
  }

  for (let i = 0; i < randInt(7, 12); i++) {
    const cx = randInt(xMin + 10, xMax - 10)
    const cz = randInt(zMin + 10, zMax - 10)
    const width = randInt(6, 12)
    const depth = randInt(6, 12)
    const height = randInt(4, 7)
    placeBuilding(cx, cz, width, depth, height)
  }

  for (let i = 0; i < 80; i++) {
    const x = randInt(xMin + 2, xMax - 2)
    const z = randInt(zMin + 2, zMax - 2)
    const base = getHeight(x, z)
    const h = randInt(1, 4)
    const type: BlockType = chance(0.5) ? 'stone' : 'crate'
    for (let y = 1; y <= h; y++) {
      setBlock(x, base + y, z, type)
      if (chance(0.35)) setBlock(x + 1, base + y, z, type)
      if (chance(0.35)) setBlock(x, base + y, z + 1, type)
    }
  }

  return blocks
}
