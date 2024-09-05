const en = [
  {
    text: "Guide",
    link: "/guide",
    collapsed: false,
    items: [
      { text: "Test Drive", link: "/guide/test-drive" },
      { text: "Testing locally", link: "/guide/testing-locally" },
    ],
  },
  {
    text: "Docs",
    link: "/docs",
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
          { text: "Server Configuration", link: "/docs/server-configuration" },
          { text: "Server Side Repo Config", link: "/docs/server-side-repo-config" },
          { text: "Pre Workflow Hooks", link: "/docs/pre-workflow-hooks" },
          { text: "Post Workflow Hooks", link: "/docs/post-workflow-hooks" },
          { text: "Conftest Policy Checking", link: "/docs/policy-checking" },
          { text: "Custom Workflows", link: "/docs/custom-workflows" },
          { text: "Repo and Project Permissions", link: "/docs/repo-and-project-permissions" },
          { text: "Repo Level atlantis.yaml", link: "/docs/repo-level-atlantis-yaml" },
          { text: "Upgrading atlantis.yaml", link: "/docs/upgrading-atlantis-yaml" },
          { text: "Command Requirements", link: "/docs/command-requirements" },
          { text: "Checkout Strategy", link: "/docs/checkout-strategy" },
          { text: "Terraform Versions", link: "/docs/terraform-versions" },
          { text: "Terraform Cloud", link: "/docs/terraform-cloud" },
          { text: "Using Slack Hooks", link: "/docs/using-slack-hooks" },
          { text: "Stats", link: "/docs/stats" },
          { text: "FAQ", link: "/docs/faq" },
        ]
      },
      {
        text: "Using Atlantis",
        collapsed: true,
        items: [
          { text: "Overview", link: "/docs/using-atlantis" },
          { text: "API endpoints", link: "/docs/api-endpoints" },
        ]
      },
      {
        text: 'How Atlantis Works',
        collapsed: true,
        items: [
          { text: 'Overview', link: '/docs/how-atlantis-works', },
          { text: 'Locking', link: '/docs/locking', },
          { text: 'Autoplanning', link: '/docs/autoplanning', },
          { text: 'Automerging', link: '/docs/automerging', },
          { text: 'Security', link: '/docs/security', },
        ]
      },
      {
        text: 'Real-time Terraform Logs',
        link: '/docs/streaming-logs',
      },
      {
        text: 'Troubleshooting',
        collapsed: true,
        items: [
          { text: 'HTTPS, SSL, TLS', 'link': '/docs/troubleshooting-https', },
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
          { text: "Events Controller", link: "/contributing/events-controller" },
        ]
      },
      { text: "Glossary", link: "/contributing/glossary" },
    ]

  },
  {
    text: "Blog",
    link: "/blog",
    collapsed: false,
    items: [
      {
        text: "2024",
        collapsed: true,
        items: [
          {
            text: "Integrating Atlantis with OpenTofu",
            link: "/blog/2024/integrating-atlantis-with-opentofu"
          },
          {
            text: "Atlantis User Survey Results",
            link: "/blog/2024/april-2024-survey-results"
          },
        ]
      },
      {
        text: "2019",
        collapsed: true,
        items: [
          {
            text: "4 Reasons To Try HashiCorp's (New) Free Terraform Remote State Storage",
            link: "/blog/2019/4-reasons-to-try-hashicorps-new-free-terraform-remote-state-storage"
          },
        ]
      },
      {
        text: "2018",
        collapsed: true,
        items: [
          {
            text: "I'm Joining HashiCorp!",
            link: "/blog/2018/joining-hashicorp"
          },
          {
            text: "Putting The Dev Into DevOps: Why Your Developers Should Write Terraform Too",
            link: "/blog/2018/putting-the-dev-into-devops-why-your-developers-should-write-terraform-too"
          },
          {
            text: "Atlantis 0.4.4 Now Supports Bitbucket",
            link: "/blog/2018/atlantis-0-4-4-now-supports-bitbucket"
          },
          {
            text: "Terraform And The Dangers Of Applying Locally",
            link: "/blog/2018/terraform-and-the-dangers-of-applying-locally"
          },
          {
            text: "Hosting Our Static Site over SSL with S3, ACM, CloudFront and Terraform",
            link: "/blog/2018/hosting-our-static-site-over-ssl-with-s3-acm-cloudfront-and-terraform"
          },
        ]
      },
      {
        text: "2017",
        collapsed: true,
        items: [
          { text: "Introducing Atlantis", link: "/blog/2017/introducing-atlantis" },
        ]
      },
    ]
  }
]

export { en }
