import { Browser } from "@wailsio/runtime";

/**
 * Opens a URL in the system browser. Falls back to logging on failure so that
 * a bad URL never blocks the page. See docs/ARCHITECTURE.md ADR-006.
 */
export async function openExternal(url: string): Promise<void> {
  try {
    await Browser.OpenURL(url);
  } catch (err) {
    console.error("打开外部链接失败:", url, err);
  }
}
