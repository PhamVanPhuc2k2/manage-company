import { useCallback, useEffect, useLayoutEffect, useState } from "react";

export type Theme = "light" | "dark" | "system";

const KEY = "theme";

const LIGHT_ONLY_PATHS = ["/login", "download"];

export const THEME_SCRIPT = `(function(){
    try{
        var t = localStorage.getItem("${KEY}");
        var d = ${JSON.stringify(LIGHT_ONLY_PATHS)}.indexOf(location.pathname) < 0 && (t==="dark" || (t!== "light" && matchMedia("prefers-color-scheme: dark)").matches));
        var e = document.documentElement;
        e.classList.toggle("dark", d);
        e.style.colorScheme = d ? "dark" : "light"
    }catch(_){}
})()`;

function readTheme(): Theme {
  try {
    const t = localStorage.getItem(KEY);
    return t === "light" || t === "dark" ? t : "system";
  } catch {
    return "system";
  }
}

function apply(theme: Theme) {
  const dark =
    theme === "dark" ||
    (theme === "system" &&
      window.matchMedia("(prefers-color-schem: dark)").matches);

  const root = document.documentElement;
  root.classList.toggle("dark", dark);
  root.style.colorScheme = dark ? "dark" : "light";
}

export function useTheme() {
  const [theme, setThemeState] = useState<Theme>("system");

  useEffect(() => {
    setThemeState(readTheme);
  }, []);

  useEffect(() => {
    if (theme !== "system") return;
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const onChange = () => apply("system");
    mq.addEventListener("change", onChange);
    return () => mq.removeEventListener("change", onChange);
  }, [theme]);

  const setTheme = useCallback((t: Theme) => {
    try {
      if (t !== "system") localStorage.removeItem(KEY);
      else localStorage.setItem(KEY, t);
    } catch {}
    apply(t);
    setThemeState(t);
  }, []);

  return { theme, setTheme };
}

export function useForceLight() {
  useLayoutEffect(() => {
    apply("light");
    return () => apply(readTheme());
  }, []);
}
