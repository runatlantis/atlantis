import { defineConfig } from 'vitepress';
import * as navbars from "./navbars";
import * as sidebars from "./sidebars";

// https://vitepress.dev/reference/site-config
export default defineConfig({
    title: 'Atlantis',
    description: 'Atlantis: Terraform Pull Request Automation',
    lang: 'en-US',
    lastUpdated: true,
    locales: {
        root: {
            label: 'English',
            lang: 'en-US',
            themeConfig: {
                nav: navbars.en,
                sidebar: sidebars.en,
            },
        },
    },
    themeConfig: {
        // https://vitepress.dev/reference/default-theme-config
        editLink: {
            pattern: 'https://github.com/runatlantis/atlantis/edit/main/runatlantis.io/:path'
        },
        // headline "depth" the right nav will show for its TOC
        //
        // https://vitepress.dev/reference/frontmatter-config#outline
        outline: [2, 3],
        search: {
            provider: 'algolia',
            options: {
                // We internally discussed how this API key is exposed in the code and decided
                // that it is a non-issue because this API key can easily be extracted by
                // looking at the browser dev tools since the key is used in the API requests.
                apiKey: '3b733dff1539ca3a210775860301fa86',
                indexName: 'runatlantis',
                appId: 'BH4D9OD16A',
                locales: {
                    '/': {
                        placeholder: 'Search Documentation',
                        translations: {
                            button: {
                                buttonText: 'Search Documentation',
                            },
                        },
                    },
                },
            }
        },
        socialLinks: [
          { icon: "slack", link: "https://join.slack.com/t/atlantis-community/shared_invite/zt-9xlxtxtc-CUSKB1ATt_sQy6um~LDPNw" },
          { icon: "twitter", link: "https://twitter.com/runatlantis" },
          { icon: "github", link: "https://github.com/runatlantis/atlantis" },
        ],
    },
    head: [
        ['link', { rel: 'icon', type: 'image/png', href: '/favicon-196x196.png', sizes: '196x196' }],
        ['link', { rel: 'icon', type: 'image/png', href: '/favicon-96x96.png', sizes: '96x96' }],
        ['link', { rel: 'icon', type: 'image/png', href: '/favicon-32x32.png', sizes: '32x32' }],
        ['link', { rel: 'icon', type: 'image/png', href: '/favicon-16x16.png', sizes: '16x16' }],
        ['link', { rel: 'icon', type: 'image/png', href: '/favicon-128.png', sizes: '128x128' }],
        ['link', { rel: 'apple-touch-icon-precomposed', sizes: '57x57', href: '/apple-touch-icon-57x57.png' }],
        ['link', { rel: 'apple-touch-icon-precomposed', sizes: '114x114', href: '/apple-touch-icon-114x114.png' }],
        ['link', { rel: 'apple-touch-icon-precomposed', sizes: '72x72', href: '/apple-touch-icon-72x72.png' }],
        ['link', { rel: 'apple-touch-icon-precomposed', sizes: '144x144', href: '/apple-touch-icon-144x144.png' }],
        ['link', { rel: 'apple-touch-icon-precomposed', sizes: '60x60', href: '/apple-touch-icon-60x60.png' }],
        ['link', { rel: 'apple-touch-icon-precomposed', sizes: '120x120', href: '/apple-touch-icon-120x120.png' }],
        ['link', { rel: 'apple-touch-icon-precomposed', sizes: '76x76', href: '/apple-touch-icon-76x76.png' }],
        ['link', { rel: 'apple-touch-icon-precomposed', sizes: '152x152', href: '/apple-touch-icon-152x152.png' }],
        ['meta', { name: 'msapplication-TileColor', content: '#FFFFFF' }],
        ['meta', { name: 'msapplication-TileImage', content: '/mstile-144x144.png' }],
        ['meta', { name: 'msapplication-square70x70logo', content: '/mstile-70x70.png' }],
        ['meta', { name: 'msapplication-square150x150logo', content: '/mstile-150x150.png' }],
        ['meta', { name: 'msapplication-wide310x150logo', content: '/mstile-310x150.png' }],
        ['meta', { name: 'msapplication-square310x310logo', content: '/mstile-310x310.png' }],
        ['link', { rel: 'stylesheet', sizes: '152x152', href: 'https://fonts.googleapis.com/css?family=Lato:400,900' }],
        ['meta', { name: 'google-site-verification', content: 'kTnsDBpHqtTNY8oscYxrQeeiNml2d2z-03Ct9wqeCeE' }],
        // google analytics
        [
            'script',
            { async: '', src: 'https://www.googletagmanager.com/gtag/js?id=UA-6850151-3' }
        ],
        [
            'script',
            {},
            `window.dataLayer = window.dataLayer || [];
            function gtag(){dataLayer.push(arguments);}
            gtag('js', new Date());

            gtag('config', 'UA-6850151-3');`
        ],
        [
            'script',
            { id: 'restore-banner-preference' },
            `
        (() => {
          const restore = (key, cls, def = false) => {
            const saved = localStorage.getItem(key);
            if (saved ? saved !== 'false' && new Date() < saved : def) {
              document.documentElement.classList.add(cls);
            }
          };
          restore('survey-banner', 'banner-dismissed');
        })();`,
        ]
    ],
    sitemap: {
        hostname: 'https://runatlantis.io'
    },
    vite: {
        server: {
            fs: {
                cachedChecks: false,
            },
        }
    }
})
