const en = {
  '/guide/': [
      '/guide/',
      '/guide/test-drive',
      '/guide/testing-locally',
  ],
  '/docs/': [
      {
          text: 'Installing Atlantis',
          collapsible: true,
          children: [
              '/docs/installation-guide',
              '/docs/requirements',
              '/docs/access-credentials',
              '/docs/webhook-secrets',
              '/docs/deployment',
              '/docs/configuring-webhooks',
              '/docs/provider-credentials',
          ]
      },
      {
          text: 'Configuring Atlantis',
          collapsible: true,
          children: [
              {
                  text: 'Overview',
                  link: '/docs/configuring-atlantis',
              },
              '/docs/server-configuration',
              '/docs/server-side-repo-config',
              '/docs/pre-workflow-hooks',
              '/docs/post-workflow-hooks',
              '/docs/policy-checking',
              '/docs/custom-workflows',
              '/docs/repo-level-atlantis-yaml',
              '/docs/upgrading-atlantis-yaml',
              '/docs/command-requirements',
              '/docs/checkout-strategy',
              '/docs/terraform-versions',
              '/docs/terraform-cloud',
              '/docs/using-slack-hooks',
              '/docs/stats',
              '/docs/faq',
          ]
      },
      {
          text: 'Using Atlantis',
          collapsible: true,
          children: [
              {
                  text: 'Overview',
                  link: '/docs/using-atlantis',
              },
              '/docs/api-endpoints',
          ]
      },
      {
          text: 'How Atlantis Works',
          collapsible: true,
          children: [
              {
                  text: 'Overview',
                  link: '/docs/how-atlantis-works',
              },
              '/docs/locking',
              '/docs/autoplanning',
              '/docs/automerging',
              '/docs/security',
          ]
      },
      {
          text: 'Real-time Terraform Logs',
          collapsible: true,
          children: [
              '/docs/streaming-logs',
          ]
      },
      {
          text: 'Troubleshooting',
          collapsible: true,
          children: [
              '/docs/troubleshooting-https',
          ]
      }
  ],
  '/contributing/': [
      {
          text: 'Implementation Details',
          children: [
              '/contributing/events-controller',
          ]
      },
      '/contributing/glossary',
  ],
};

export default { en };
