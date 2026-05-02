import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

export default defineConfig({
  envDir: "..",
  plugins: [react()],
  server: {
    host: "127.0.0.1",
    allowedHosts: ["deal.adsrivatsa.com", "127.0.0.1", "deal-test.adsrivatsa.com"],
  },
});
