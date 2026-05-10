import { readFile } from "node:fs/promises";
import path from "node:path";
import { parse as parseYaml } from "yaml";

export interface Config {
  app: { "shutdown-timeout": string };
  http: { port: number };
  logs: Record<string, unknown>;
  shortcut: { "base-url": string };
}

const defaultConfigsDir = "./configs";
const baseConfigName = "base.yaml";

export async function loadConfig(): Promise<Config> {
  const env = process.env.ENV || "dev";
  const configsDir = process.env.CONFIGS_DIR || defaultConfigsDir;

  const base = await loadFile(path.join(configsDir, baseConfigName));
  const override = await loadFile(path.join(configsDir, `${env}.yaml`));

  return deepMerge(base, override) as unknown as Config;
}

async function loadFile(filePath: string): Promise<Record<string, unknown>> {
  try {
    const raw = await readFile(filePath, "utf-8");
    const parsed = parseYaml(raw);
    return (parsed ?? {}) as Record<string, unknown>;
  } catch (err) {
    if ((err as NodeJS.ErrnoException).code === "ENOENT") {
      return {};
    }
    throw err;
  }
}

function deepMerge(
  a: Record<string, unknown>,
  b: Record<string, unknown>,
): Record<string, unknown> {
  const out: Record<string, unknown> = { ...a };
  for (const [key, value] of Object.entries(b)) {
    const prev = out[key];
    if (isObject(prev) && isObject(value)) {
      out[key] = deepMerge(prev, value);
    } else {
      out[key] = value;
    }
  }
  return out;
}

function isObject(v: unknown): v is Record<string, unknown> {
  return typeof v === "object" && v !== null && !Array.isArray(v);
}
