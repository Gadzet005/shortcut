import { readFile } from "node:fs/promises";
import path from "node:path";
import type { Request, Response } from "express";
import { parse as parseYaml } from "yaml";
import { readCookie } from "./http.js";

const USER_COOKIE = "user_id";
const COOKIE_MAX_AGE_MS = 365 * 24 * 60 * 60 * 1000;

export class UserSession {
  constructor(private readonly userIds: string[]) {
    if (userIds.length === 0) {
      throw new Error("UserSession requires at least one user id");
    }
  }

  get availableIds(): string[] {
    return [...this.userIds];
  }

  pickUserId(req: Request, res: Response): string {
    const fromCookie = readCookie(req, USER_COOKIE);
    if (fromCookie && this.userIds.includes(fromCookie)) return fromCookie;
    const picked = this.userIds[Math.floor(Math.random() * this.userIds.length)] as string;
    res.cookie(USER_COOKIE, picked, {
      httpOnly: true,
      sameSite: "lax",
      maxAge: COOKIE_MAX_AGE_MS,
      path: "/",
    });
    return picked;
  }
}

export async function createUserSession(): Promise<UserSession> {
  const ids = await loadUserIds();
  return new UserSession(ids);
}

async function loadUserIds(): Promise<string[]> {
  const dir = process.env.MOCK_DATA_DIR ?? "/app/mock-data";
  const filePath = path.join(dir, "users.yaml");
  try {
    const raw = await readFile(filePath, "utf-8");
    const parsed = parseYaml(raw) as { users?: { id: unknown }[] } | null;
    const ids = (parsed?.users ?? [])
      .map((u) => (u && u.id != null ? String(u.id) : ""))
      .filter((id) => id !== "");
    if (ids.length > 0) return ids;
  } catch (err) {
    console.warn(JSON.stringify({ msg: "failed to load users.yaml", err: String(err) }));
  }
  return ["1"];
}
