// Minimal hash router. See docs/ARCHITECTURE.md ADR-006.
// Routes: "#/", "#/scenes/<id>", "#/performers", "#/tags", "#/studios".

export type RouteName = "scenes" | "scene" | "performers" | "tags" | "studios";

export type Route = {
  name: RouteName;
  id: number;
};

export type NavItem = {
  name: RouteName;
  path: string;
  label: string;
};

// 导航是路由的单一事实来源，由本模块导出。
export const NAV_ITEMS: NavItem[] = [
  { name: "scenes", path: "/", label: "场景" },
  { name: "performers", path: "/performers", label: "演员" },
  { name: "tags", path: "/tags", label: "标签" },
  { name: "studios", path: "/studios", label: "工作室" },
];

const ROUTES: Array<{ re: RegExp; name: RouteName }> = [
  { re: /^\/scenes\/(\d+)\/?$/, name: "scene" },
  { re: /^\/scenes\/?$/, name: "scenes" },
  { re: /^\/?$/, name: "scenes" },
  { re: /^\/performers\/?$/, name: "performers" },
  { re: /^\/tags\/?$/, name: "tags" },
  { re: /^\/studios\/?$/, name: "studios" },
];

function parseHash(hash: string): Route {
  const path = hash.startsWith("#") ? hash.slice(1) : hash;
  for (const r of ROUTES) {
    const match = r.re.exec(path);
    if (match) {
      return { name: r.name, id: match[1] ? Number(match[1]) : 0 };
    }
  }
  return { name: "scenes", id: 0 };
}

export const route = $state<Route>({ name: "scenes", id: 0 });

function sync() {
  const next = parseHash(window.location.hash);
  route.name = next.name;
  route.id = next.id;
}

export function navigate(path: string): void {
  const target = path.startsWith("#") ? path : "#" + path;
  if (window.location.hash === target) {
    sync();
    return;
  }
  window.location.hash = target;
}

export function initRouter(): () => void {
  sync();
  window.addEventListener("hashchange", sync);
  return () => window.removeEventListener("hashchange", sync);
}

// 导航高亮：scene 详情视为 scenes 区。
export function isActive(name: RouteName): boolean {
  if (name === "scenes") {
    return route.name === "scenes" || route.name === "scene";
  }
  return route.name === name;
}
