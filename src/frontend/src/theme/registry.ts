export const THEME_STORAGE_KEY = "monopoly-deal-theme-id";

export const themeRegistry = [
  {
    id: "ocean-dark",
    label: "Ocean Dark",
  },
  {
    id: "lavender-light",
    label: "Lavender Light",
  },
  {
    id: "rose-light",
    label: "Rose Light",
  },
  {
    id: "rose-dark",
    label: "Rose Dark",
  },
  {
    id: "ash-orange",
    label: "Ash Orange",
  },
] as const;

export type ThemeId = (typeof themeRegistry)[number]["id"];

export const DEFAULT_THEME_ID: ThemeId = "ocean-dark";

export const isThemeId = (value: string): value is ThemeId => {
  return themeRegistry.some((theme) => theme.id === value);
};
