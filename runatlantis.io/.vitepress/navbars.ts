import { existsSync } from "node:fs";
import path from "node:path";

interface NavItem {
  text: string;
  link: string;
}

const en: NavItem[] = [
  { text: "Home", link: "/" },
  { text: "Guide", link: "/guide" },
  { text: "Docs", link: "/docs" },
  { text: "Contributing", link: "/contributing" },
  { text: "Blog", link: "/blog" },
];

const SITE_ROOT = path.resolve(import.meta.dirname, "..");

// Same fallback rule as the sidebar: link to the localized page when it exists,
// otherwise send the reader to the English one rather than a dead link.
const localizedLink = (link: string, locale: string): string => {
  const candidate = link === "/" ? "index" : link;
  return existsSync(path.join(SITE_ROOT, locale, `${candidate}.md`))
    ? `/${locale}${link === "/" ? "/" : link}`
    : link;
};

const esLabels: Record<string, string> = {
  Home: "Inicio",
  Guide: "Guía",
  Docs: "Documentación",
  Contributing: "Contribuir",
  Blog: "Blog",
};

const es: NavItem[] = en.map((item) => ({
  text: esLabels[item.text] ?? item.text,
  link: localizedLink(item.link, "es"),
}));

export { en, es };
