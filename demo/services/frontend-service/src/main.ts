import { loadConfig } from "./config.js";
import { startServer } from "./service.js";

async function main(): Promise<void> {
  const cfg = await loadConfig();
  await startServer(cfg);
}

main().catch((err: unknown) => {
  console.error("startup failed", err);
  process.exit(1);
});
