const siteHosts = new Set(['runatlantis.io', 'www.runatlantis.io']);
const sharedSection = /^\/(?:es\/)?(contributing|blog)(\/|\.|$)/;

export const versionBase = (version) => version ? `/versions/${version}/` : '/';

export function flagVersionBadges(md) {
  md.core.ruler.before('inline', 'flag-version-badges', (state) => {
    for (let i = 1; i < state.tokens.length; i++) {
      const token = state.tokens[i];
      if (state.tokens[i - 1].type === 'heading_open' && token.type === 'inline' && /^`--[^`]+`/.test(token.content)) {
        token.content = token.content.replace(/\s*<Badge\b(?=[^>]*\btext=(["'])v\d+\.\d+(?:\.\d+)?\+?\1)[^>]*\/>/gi, '');
      }
    }
  });
}

// Markdown links still pass through VitePress's base handling; raw HTML does not.
export function historicalLink(href, version, rawHTML = false) {
  if (!href || href.startsWith('#')) return href;
  let result = href;
  if (/^(https?:)?\/\//i.test(result)) {
    const url = new URL(result, 'https://www.runatlantis.io');
    if (!siteHosts.has(url.hostname) || !/^\/(?:es\/)?(docs|guide)(\/|\.|$)/.test(url.pathname)) return href;
    result = url.pathname + url.search + url.hash;
  }
  if (sharedSection.test(result)) return result.replace(/\.md(?=[?#]|$)/, '.html');
  result = result.replace(/(api-endpoints\.(?:md|html))#post-apidriftdetect$/, '$1#post-api-drift-detect');
  if (rawHTML && result.startsWith('/') && !result.startsWith('//') && !result.startsWith(versionBase(version))) {
    result = versionBase(version) + result.slice(1);
  }
  return result;
}

export function versionedLinks(md, version, versions = [], sharedPages, rootVersion) {
  md.core.ruler.after('inline', 'historical-links', (state) => {
    const relativePath = state.env.relativePath ?? '';
    const historical = Boolean(version) || /^(?:es\/)?(?:docs|guide)(\/|\.md$)/.test(relativePath);
    const rootRelease = !version && versions.find((entry) => rootVersion ? entry.version === rootVersion : entry.version !== 'main');
    const latest = rootRelease && relativePath.startsWith('es/') && rootRelease;
    if (!historical && !latest && !sharedPages) return;
    function rewriteLink(href, rawHTML = false) {
      if (latest && href) {
        const url = new URL(href, `https://docs.invalid/${relativePath}`);
        if (url.hostname === 'docs.invalid' && /^\/es\/(?:docs|guide)(\/|\.|$)/.test(url.pathname)
          && !pageForLink(url.pathname, latest.pages) && pageForLink(url.pathname.slice(3), latest.pages)) {
          href = url.pathname.slice(3) + url.search + url.hash;
        }
      }
      href = historical ? historicalLink(href, version, rawHTML) : href;
      if (sharedPages) validateSharedLink(href, sharedPages, relativePath);
      return href;
    }
    function rewrite(tokens) {
      for (const token of tokens) {
        if (token.type === 'link_open') {
          const href = rewriteLink(token.attrGet('href'));
          token.attrSet('href', href);
          if (version && sharedSection.test(href)) token.attrSet('target', '_self');
        }
        if (token.type === 'html_inline' || token.type === 'html_block') {
          token.content = token.content.replace(/\bhref=(['"])(.*?)\1/g, (_, quote, href) => `href=${quote}${rewriteLink(href, true)}${quote}`);
        }
        if (token.children) rewrite(token.children);
      }
    }
    rewrite(state.tokens);
  });
}

function pageForLink(link, pages) {
  const path = link.split(/[?#]/)[0].replace(/^\//, '').replace(/\.(html|md)$/, '');
  return pages.includes(`${path}.md`) || pages.includes(`${path.replace(/\/$/, '')}/index.md`);
}

export function validateSharedLink(href, pages, source) {
  if (!href || !sharedSection.test(href)) return;
  let pathname;
  try {
    pathname = decodeURIComponent(new URL(href, 'https://docs.invalid').pathname);
  } catch (cause) {
    throw new Error(`Invalid current-site link ${href} in ${source}`, { cause });
  }
  if (!pages.includes(pathname.slice(1)) && !pageForLink(pathname, pages)) {
    throw new Error(`Broken current-site link ${href} in ${source}`);
  }
}

export function historicalSidebar(items, pages) {
  return items.flatMap((item) => {
    const children = item.items && historicalSidebar(item.items, pages);
    const candidate = item.link?.startsWith('/es/') && !pageForLink(item.link, pages) ? item.link.slice(3) : item.link;
    let link = candidate && pageForLink(candidate, pages) ? candidate : undefined;
    if (link && !/[.#?]|\/$/.test(link) && !pages.includes(`${link.slice(1)}.md`)) link += '/';
    if (!link && !children?.length) return [];
    return [{ ...item, link, ...(children ? { items: children } : {}) }];
  });
}

export function configureVersions(config, version, versions, sidebar, metadata = {}) {
  if (metadata.sharedPages) {
    const check = (items) => items?.forEach((item) => {
      validateSharedLink(item.link, metadata.sharedPages, 'sidebar');
      check(item.items);
    });
    check(sidebar);
    for (const locale of Object.values(config.locales)) check(locale.themeConfig?.sidebar);
  }
  config.themeConfig.docVersions = versions.map(({ version }) => ({ version }));
  config.themeConfig.docVersion = version;
  if (!versions.length) return config;
  const latest = versions.find((entry) => entry.version !== 'main');
  const rootRelease = versions.find((entry) => entry.version === metadata.rootVersion) ?? latest;
  if (!version && rootRelease?.version === 'main') config.themeConfig.docVersion = 'main';
  const transformPageData = config.transformPageData;
  config.transformPageData = async (page, context) => {
    Object.assign(page, await transformPageData?.(page, context));
    const documentation = /^(?:es\/)?(?:docs|guide)(\/|\.md$)/.test(page.relativePath);
    const ref = version || (documentation ? rootRelease.version : 'main');
    const source = version || documentation ? metadata.sources?.[ref]?.[page.relativePath] ?? metadata.sources?.[ref]?.[page.filePath] : page.filePath;
    if (version) {
      if (source) page.sourceURL = `https://github.com/runatlantis/atlantis/blob/${ref}/runatlantis.io/${source}`;
      else page.frontmatter.editLink = false;
    }
    if (!version && page.frontmatter.lastUpdated !== false && !(page.frontmatter.lastUpdated instanceof Date)) {
      const timestamp = metadata.lastUpdated?.[ref]?.[source];
      if (timestamp) page.lastUpdated = timestamp;
    }
    if (!version && documentation) page.filePath = metadata.sources?.main?.[page.relativePath] ?? page.filePath;
  };
  if (!version) {
    if (rootRelease && !rootRelease.pages.some((page) => page.startsWith('es/docs/'))) config.themeConfig.i18nRouting = false;
    if (rootRelease) config.locales.root.themeConfig.sidebar = sidebar.flatMap((item) =>
      sharedSection.test(item.link ?? '') ? [item] : historicalSidebar([item], rootRelease.pages));
    if (rootRelease && config.locales.es) config.locales.es.themeConfig.sidebar = config.locales.es.themeConfig.sidebar.flatMap((item) =>
      sharedSection.test(item.link ?? '') ? [item] : historicalSidebar([item], rootRelease.pages));
    return config;
  }
  const release = versions.find((entry) => entry.version === version);
  if (!release) throw new Error(`missing documentation manifest for ${version}`);
  config.base = versionBase(version);
  config.title = `Atlantis ${version}`;
  config.themeConfig.siteTitle = 'Atlantis';
  config.lastUpdated = false;
  config.themeConfig.sharedSidebar = sidebar.filter((item) => sharedSection.test(item.link ?? ''));
  config.themeConfig.editLink = {
    pattern: ({ sourceURL }) => sourceURL || '',
    text: version === 'main' ? 'Browse main source' : 'Browse release source',
  };
  const spanish = config.locales.es;
  config.locales = { root: { label: 'English', lang: 'en-US', themeConfig: {
    nav: [{ text: 'Documentation', link: '/docs' }],
    sidebar: historicalSidebar(sidebar.filter((item) => !sharedSection.test(item.link ?? '')), release.pages),
  } } };
  if (spanish && release.pages.some((page) => page.startsWith('es/docs/'))) {
    config.locales.es = { ...spanish, themeConfig: {
      ...spanish.themeConfig,
      editLink: config.themeConfig.editLink,
      nav: [{ text: 'Documentación', link: '/es/docs' }],
      sidebar: historicalSidebar(spanish.themeConfig.sidebar.filter((item) => !sharedSection.test(item.link ?? '')), release.pages),
      sharedSidebar: spanish.themeConfig.sidebar.filter((item) => sharedSection.test(item.link ?? '')),
    } };
  }
  config.themeConfig.search = { provider: 'local' };
  config.head = config.head.map(([tag, attributes, ...rest]) => [tag, Object.fromEntries(
    Object.entries(attributes).map(([key, value]) => [key,
      ['href', 'src', 'content'].includes(key) && typeof value === 'string' && value.startsWith('/') && !value.startsWith('//')
        ? versionBase(version) + value.slice(1) : value]),
  ), ...rest]);
  config.head.push(['meta', { name: 'robots', content: 'noindex, follow' }]);
  config.buildEnd = undefined;
  return config;
}

export function versionTarget(relativePath, version, versions) {
  const pages = versions.find((entry) => entry.version === version)?.pages;
  if (!pages?.includes(relativePath) && relativePath.startsWith('es/')) {
    const english = relativePath.slice(3);
    if (pages?.includes(english)) relativePath = english;
  }
  const missingPage = Boolean(version && !pages?.includes(relativePath));
  const indexes = relativePath.startsWith('es/') ? ['es/docs.md', 'es/index.md', 'docs.md', 'index.md'] : ['docs.md', 'index.md'];
  const page = missingPage ? indexes.find((index) => pages?.includes(index)) ?? 'index.md' : relativePath;
  return {
    path: versionBase(version) + page.replace(/(^|\/)index\.md$/, '$1').replace(/\.md$/, '.html'),
    missingPage,
  };
}
