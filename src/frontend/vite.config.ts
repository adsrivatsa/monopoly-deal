import react from "@vitejs/plugin-react";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { defineConfig } from "vite";

const parseEnvFile = (filePath: string): Record<string, string> => {
  const parsed: Record<string, string> = {};

  try {
    const fileContents = readFileSync(filePath, "utf-8");
    for (const rawLine of fileContents.split(/\r?\n/)) {
      const line = rawLine.trim();
      if (!line || line.startsWith("#")) {
        continue;
      }

      const separatorIndex = line.indexOf("=");
      if (separatorIndex <= 0) {
        continue;
      }

      const key = line.slice(0, separatorIndex).trim();
      const rawValue = line.slice(separatorIndex + 1).trim();
      const value = rawValue.replace(/^['"]|['"]$/g, "");
      parsed[key] = value;
    }
  } catch {
    return parsed;
  }

  return parsed;
};

export default defineConfig(() => {
  const devEnvPath = resolve(__dirname, "../prod.env");
  const devEnv = parseEnvFile(devEnvPath);
  const viteEnv = Object.fromEntries(
    Object.entries(devEnv).filter(([key]) => key.startsWith("VITE_")),
  );

  return {
    envDir: "..",
    define: Object.fromEntries(
      Object.entries(viteEnv).map(([key, value]) => [
        `import.meta.env.${key}`,
        JSON.stringify(value),
      ]),
    ),
    plugins: [react()],
    server: {
      host: "127.0.0.1",
      allowedHosts: ["deal.adsrivatsa.com", "127.0.0.1"],
    },
  };
});
