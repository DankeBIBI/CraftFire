/**
 * Wails 服务封装层。
 * 提供从前端调用 Go 后端方法的统一接口。
 * 在 Wails 环境中，这些方法通过 Wails 的 JS 绑定直接调用 Go 函数。
 */

import type { LANServerInfo, RoomConfig } from "@/types/room";
import { logger } from "@/utils/logger";

/** Wails 运行时接口（由 Wails 自动注入） */
declare global {
  interface Window {
    go: {
      main: {
        App: {
          CreateRoom: () => Promise<string>;
          JoinRoom: (roomId: string, ip: string) => Promise<boolean>;
          LeaveRoom: () => Promise<void>;
          FindLANServers: () => Promise<
            Array<{
              roomId: string;
              ip: string;
              playerCount: number;
              maxPlayers: number;
              gameMode: string;
            }>
          >;
          PlaceBlock: (
            x: number,
            y: number,
            z: number,
            blockType: string,
          ) => Promise<void>;
          RemoveBlock: (x: number, y: number, z: number) => Promise<void>;
          VerifyAdminPassword: (
            roomId: string,
            password: string,
          ) => Promise<{ token: string; expiresAt: number }>;
          GetOnlinePlayers: (roomId: string) => Promise<unknown[]>;
          GetPlayerDetails: (
            roomId: string,
            playerId: string,
          ) => Promise<unknown>;
          KickPlayer: (
            roomId: string,
            playerId: string,
            reason: string,
          ) => Promise<void>;
          MutePlayer: (
            roomId: string,
            playerId: string,
            durationSeconds: number,
          ) => Promise<void>;
          GetRoomStats: (roomId: string) => Promise<unknown>;
          SetRoomLocked: (roomId: string, locked: boolean) => Promise<void>;
          IsRoomLocked: (roomId: string) => Promise<boolean>;
          BroadcastToRoom: (roomId: string, message: string) => Promise<void>;
          ChangeGameMode: (roomId: string, mode: string) => Promise<void>;
          HealPlayer: (roomId: string, playerId: string) => Promise<void>;
          TeleportPlayer: (
            roomId: string,
            playerId: string,
            x: number,
            y: number,
            z: number,
          ) => Promise<void>;
          GetRoomConfig: (roomId: string) => Promise<RoomConfig>;
          ImportModel: (filePath: string, roomId: string) => Promise<string>;
          ListAvailableModels: (roomId: string) => Promise<unknown[]>;
          GetModelInfo: (modelId: string) => Promise<unknown>;
          DeleteModel: (modelId: string) => Promise<void>;
          SyncModelsInLAN: () => Promise<unknown[]>;
          DownloadModelFromLAN: (
            modelId: string,
            sourceIP: string,
          ) => Promise<void>;
          GetPlayerProfile: (playerId: string) => Promise<unknown>;
          UpdatePlayerProfile: (
            nickname: string,
            skinColor: string,
          ) => Promise<void>;
          GetPlayerStatistics: (playerId: string) => Promise<unknown>;
        };
      };
    };
  }
}

/**
 * 获取 Wails Go 后端绑定。
 * 在非 Wails 环境下返回模拟实现。
 */
function getBackend() {
  if (window.go?.main?.App) {
    return window.go.main.App;
  }
  logger.warn("Wails", "Wails 运行时不可用，使用模拟模式");
  return null;
}

/**
 * 创建游戏房间。
 * 返回值为标准化后的房间配置对象。
 */
export async function CreateRoom(): Promise<RoomConfig> {
  const backend = getBackend();
  let roomId = backend
    ? await backend.CreateRoom()
    : String(100000 + Math.floor(Math.random() * 900000));

  // 校验 roomId 格式，确保是6位数字
  if (!/^\d{6}$/.test(roomId)) {
    logger.error("Wails", `roomId 格式错误: ${roomId}，将重新生成`);
    roomId = String(100000 + Math.floor(Math.random() * 900000));
  }

  const now = Date.now();
  const port = Number.parseInt(roomId, 10);

  return {
    roomId,
    port,
    maxPlayers: 10,
    currentPlayers: 1,
    worldSeed: "",
    createdAt: now,
    lastActivityAt: now,
    isPublic: true,
    gameMode: "sandbox",
    ip: "127.0.0.1",
  };
}

/**
 * 加入游戏房间。
 * 参数 ip 为空时默认走本机回环地址。
 */
export async function JoinRoom(roomId: string, ip: string): Promise<boolean> {
  const backend = getBackend();
  if (backend) {
    return backend.JoinRoom(roomId, ip || "127.0.0.1");
  }
  return true;
}

/** 离开当前房间 */
export async function LeaveRoom(): Promise<void> {
  const backend = getBackend();
  if (backend) {
    return backend.LeaveRoom();
  }
}

/** 局域网服务器原始数据 */
interface LANServerRaw {
  roomId: string;
  ip: string;
  playerCount: number;
  maxPlayers: number;
  gameMode: string;
}

/** 验证 LANServerRaw 对象 */
function isLANServerRaw(obj: unknown): obj is LANServerRaw {
  if (typeof obj !== "object" || obj === null) return false;
  const s = obj as LANServerRaw;
  return (
    typeof s.roomId === "string" &&
    typeof s.ip === "string" &&
    typeof s.playerCount === "number" &&
    typeof s.maxPlayers === "number" &&
    typeof s.gameMode === "string"
  );
}

/** 发现局域网服务器 */
export async function FindLANServers(): Promise<LANServerInfo[]> {
  const backend = getBackend();
  if (backend) {
    const servers = await backend.FindLANServers();
    if (!Array.isArray(servers)) {
      logger.error("Wails", "FindLANServers 返回值不是数组:", servers);
      return [];
    }
    return servers.filter(isLANServerRaw).map((item) => ({
      roomId: item.roomId,
      ip: item.ip,
      playerCount: item.playerCount,
      maxPlayers: item.maxPlayers,
      gameMode: item.gameMode,
    }));
  }
  return [];
}

/** 以下为兼容旧命名导出，避免历史调用点报错。 */
export const createRoom = CreateRoom;
export const joinRoom = JoinRoom;
export const leaveRoom = LeaveRoom;
export const findLANServers = FindLANServers;

/** 放置方块 */
export async function placeBlock(
  x: number,
  y: number,
  z: number,
  blockType: string,
) {
  const backend = getBackend();
  if (backend) {
    return backend.PlaceBlock(x, y, z, blockType);
  }
}

/** 移除方块 */
export async function removeBlock(x: number, y: number, z: number) {
  const backend = getBackend();
  if (backend) {
    return backend.RemoveBlock(x, y, z);
  }
}

/** 管理员密码验证 */
export async function verifyAdminPassword(roomId: string, password: string) {
  const backend = getBackend();
  if (backend) {
    return backend.VerifyAdminPassword(roomId, password);
  }
  if (password === "admin") {
    return {
      token: "mock-admin-token",
      expiresAt: Date.now() + 2 * 60 * 60 * 1000,
    };
  }
  throw new Error("invalid admin password");
}

/** 设置房间锁定状态 */
export async function setRoomLocked(roomId: string, locked: boolean) {
  const backend = getBackend();
  if (backend) {
    return backend.SetRoomLocked(roomId, locked);
  }
}

/** 获取房间锁定状态 */
export async function isRoomLocked(roomId: string) {
  const backend = getBackend();
  if (backend) {
    return backend.IsRoomLocked(roomId);
  }
  return false;
}

/** 向房间广播公告 */
export async function broadcastToRoom(roomId: string, message: string) {
  const backend = getBackend();
  if (backend) {
    return backend.BroadcastToRoom(roomId, message);
  }
}

/** 切换游戏模式 */
export async function changeGameMode(roomId: string, mode: string) {
  const backend = getBackend();
  if (backend) {
    return backend.ChangeGameMode(roomId, mode);
  }
}

/** 为玩家恢复生命值 */
export async function healPlayer(roomId: string, playerId: string) {
  const backend = getBackend();
  if (backend) {
    return backend.HealPlayer(roomId, playerId);
  }
}

/** 传送玩家到指定位置 */
export async function teleportPlayer(
  roomId: string,
  playerId: string,
  x: number,
  y: number,
  z: number,
) {
  const backend = getBackend();
  if (backend) {
    return backend.TeleportPlayer(roomId, playerId, x, y, z);
  }
}

/** 获取房间配置 */
export async function getRoomConfig(roomId: string) {
  const backend = getBackend();
  if (backend) {
    return backend.GetRoomConfig(roomId);
  }
  return null;
}

/** 兼容旧调用的大写导出。 */
export const PlaceBlock = placeBlock;
export const RemoveBlock = removeBlock;

/** 获取在线玩家列表 */
export async function getOnlinePlayers(roomId: string): Promise<unknown[]> {
  const backend = getBackend();
  if (backend) {
    return backend.GetOnlinePlayers(roomId);
  }
  return [];
}

/** 踢出玩家 */
export async function kickPlayer(
  roomId: string,
  playerId: string,
  reason: string,
): Promise<void> {
  const backend = getBackend();
  if (backend) {
    return backend.KickPlayer(roomId, playerId, reason);
  }
}

/** 静音玩家 */
export async function mutePlayer(
  roomId: string,
  playerId: string,
  durationSeconds: number,
): Promise<void> {
  const backend = getBackend();
  if (backend) {
    return backend.MutePlayer(roomId, playerId, durationSeconds);
  }
}

/** 获取房间统计 */
export async function getRoomStats(roomId: string): Promise<unknown> {
  const backend = getBackend();
  if (backend) {
    return backend.GetRoomStats(roomId);
  }
  return null;
}
