import { googleAnalyticsPlugin } from '@vuepress/plugin-google-analytics'
import { docsearchPlugin } from '@vuepress/plugin-docsearch'
import { getDirname, path } from '@vuepress/utils'
import { defaultTheme, defineUserConfig } from 'vuepress'
import { sitemapPlugin } from 'vuepress-plugin-sitemap2';

const __dirname = getDirname(import.meta.url)

export default defineUserConfig({
    alias: {
        '@theme/Home.vue': path.resolve(__dirname, './theme/components/Home.vue'),
    },
    locales: {
        '/': {
            lang: 'en-US',
            title: 'Atlantis',
            description: 'Atlantis: Terraform Pull Request Automation',
        },
/*
        '/es/': {
            lang: 'es-ES',
            title: 'Atlantis',
            description: 'Atlantis: Automatización de Pull Requests para Terraform',
        },
*/
    },
    plugins: [
        googleAnalyticsPlugin({
            id: 'UA-6850151-3',
        }),
        sitemapPlugin({
            hostname: 'https://runatlantis.io',
        }),
        docsearchPlugin({
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
        }),
    ],
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
    ],
    themePlugins: {
        activeHeaderLinks: false,
    },
    theme: defaultTheme({
        docsBranch: "main",
        logo: '/hero.png',
        locales: {
            '/': {
                selectLanguageName: 'English',
                navbar: [
                    { text: 'Home', link: '/' },
                    { text: 'Guide', link: '/guide/' },
                    { text: 'Docs', link: '/docs/' },
                    { text: 'Blog', link: 'https://medium.com/runatlantis' },
                ],
            },
/*
            '/es/': {
                selectLanguageName: 'Spanish',
                navbar: [
                    { text: 'Home', link: '/es/' },
                    { text: 'Guide', link: '/es/guide/' },
                    { text: 'Docs', link: '/es/docs/' },
                    { text: 'Blog', link: 'https://medium.com/runatlantis' },
                ],
            },
*/
        },
        sidebar: {
            '/guide/': [
                '',
                'test-drive',
                'testing-locally',
            ],
            '/docs/': [
                {
                    text: 'Installing Atlantis',
                    collapsible: true,
                    children: [
                        'installation-guide',
                        'requirements',
                        'access-credentials',
                        'webhook-secrets',
                        'deployment',
                        'configuring-webhooks',
                        'provider-credentials',
                    ]
                },
                {
                    text: 'Configuring Atlantis',
                    collapsible: true,
                    children: [
                        {
                            text: 'Overview',
                            link: 'configuring-atlantis',
                        },
                        'server-configuration',
                        'server-side-repo-config',
                        'pre-workflow-hooks',
                        'post-workflow-hooks',
                        'policy-checking',
                        'custom-workflows',
                        'repo-level-atlantis-yaml',
                        'upgrading-atlantis-yaml',
                        'command-requirements',
                        'checkout-strategy',
                        'terraform-versions',
                        'terraform-cloud',
                        'using-slack-hooks',
                        'stats',
                        'faq',
                    ]
                },
                {
                    text: 'Using Atlantis',
                    collapsible: true,
                    children: [
                        {
                            text: 'Overview',
                            link: 'using-atlantis',
                        },
                        'api-endpoints',
                    ]
                },
                {
                    text: 'How Atlantis Works',
                    collapsible: true,
                    children: [
                        {
                            text: 'Overview',
                            link: 'how-atlantis-works',
                        },
                        'locking',
                        'autoplanning',
                        'automerging',
                        'security',
                    ]
                },
                {
                    text: 'Real-time Terraform Logs',
                    collapsible: true,
                    children: [
                        'streaming-logs',
                    ]
                },
                {
                    text: 'Troubleshooting',
                    collapsible: true,
                    children: [
                        'troubleshooting-https',
                    ]
                }
            ]
        },
        repo: 'runatlantis/atlantis',
        docsDir: 'runatlantis.io',
        editLink: true,
    })
})
