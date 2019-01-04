module.exports = {
    title: 'Atlantis',
icon', type: 'image/png', href: '/favicon-32x32.png', sizes: '32x32' }],
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
        ['meta', {name: 'msapplication-TileColor', content: '#FFFFFF' }],
        ['meta', {name: 'msapplication-TileImage', content: '/mstile-144x144.png' }],
        ['meta', {name: 'msapplication-square70x70logo', content: '/mstile-70x70.png' }],
        ['meta', {name: 'msapplication-square150x150logo', content: '/mstile-150x150.png' }],
        ['meta', {name: 'msapplication-wide310x150logo', content: '/mstile-310x150.png' }],
        ['meta', {name: 'msapplication-square310x310logo', content: '/mstile-310x310.png' }],
        ['link', { rel: 'stylesheet', sizes: '152x152', href: 'https://fonts.googleapis.com/css?family=Lato:400,900' }],
        ['meta', {name: 'google-site-verification', content: 'kTnsDBpHqtTNY8oscYxrQeeiNml2d2z-03Ct9wqeCeE' }]
    ],
    themeConfig: {
        logo: '/hero.png',
        nav: [
            {text: 'Home', link: '/'},
            {text: 'Guide', link: '/guide/'},
            {text: 'Docs', link: '/docs/'},
            {text: 'Blog', link: 'https://medium.com/runatlantis'}
        ],
        sidebar: {
            '/docs/': [
                '',
                'pull-request-commands',
                'deployment',
                'server-configuration',
                'apply-requirements',
                'locking',
                'autoplanning',
                ['atlantis-yaml-reference', 'atlantis.yaml Reference'],
                'upgrading-atlantis-yaml-to-version-2',
                'security',
                'faq',
            ],
            '/guide/': [
                '',
                'test-drive',
                'getting-started',
                'requirements',
                'atlantis-yaml-use-cases'
            ]
        },
        repo: 'runatlantis/atlantis',
        docsDir: 'runatlantis.io',
        editLinks: true,
    }
}
