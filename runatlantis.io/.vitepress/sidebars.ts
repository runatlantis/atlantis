const en = [
  {
    text: "Guide",
    collapsed: false,
    items: [
      { text: "Introduction", link: "/guide/introduction" },
      { text: "Test Drive", link: "/guide/test-drive" },
      { text: "Testing locally", link: "/guide/testing-locally" },
    ],
  },
  {
    text: "Docs",
    collapsed: true,
    items: [
      {
        text: "Installing Atlantis",
        collapsed: true,
        items: [
          { text: "Installing Guide", link: "/docs/installation-guide" },
          { text: "Requirements", link: "/docs/requirements" },
          { text: "Git Host Access Credentials", link: "/docs/access-credentials" },
          { text: "Webhook Secrets", link: "/docs/webhook-secrets" },
          { text: "Deployment", link: "/docs/deployment" },
          { text: "Configuring Webhooks", link: "/docs/configuring-webhooks" },
          { text: "Provider Credentials", link: "/docs/provider-credentials" },
        ]
      },
      {
        text: "Configuring Atlantis",
        collapsed: true,
        items: [
          { text: "Overview", link: "/docs/configuring-atlantis" },
          "/docs/configuring-atlantis",
          "/docs/server-configuration",
          "/docs/server-side-repo-config",
          "/docs/pre-workflow-hooks",
          "/docs/post-workflow-hooks",
          "/docs/policy-checking",
          "/docs/custom-workflows",
          "/docs/repo-level-atlantis-yaml",
          "/docs/upgrading-atlantis-yaml",
          "/docs/command-requirements",
          "/docs/checkout-strategy",
          "/docs/terraform-versions",
          "/docs/terraform-cloud",
          "/docs/using-slack-hooks",
          "/docs/stats",
          "/docs/faq",
        ]
      },
      {
        text: "Using Atlantis",
        collapsed: true,
        items: [
          { text: "Overview", link: "/docs/using-atlantis" },
          "/docs/api-endpoints",
        ]
      },
      {
        text: 'How Atlantis Works',
        collapsed: true,
        items: [
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
        collapsed: true,
        items: [
          '/docs/streaming-logs',
        ]
      },
      {
        text: 'Troubleshooting',
        collapsed: true,
        items: [
          '/docs/troubleshooting',
        ]
      },
    ],
  },
  {
    text: "Contributing",
    link: "/contributing",
    collapsed: false,
    items: [
      {
        text: 'Implementation Details',
        items: [
            {text: "Events Controller", link: "/contributing/events-controller"},
        ]
      },
      {text: "Glossry", link: "/contributing/glossary"},
    ]

  }
];

export { en };
