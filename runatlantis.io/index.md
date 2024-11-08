---
# https://vitepress.dev/reference/default-theme-home-page
layout: home

pageClass: home-custom

hero:
  name: Atlantis
  text: Terraform Pull Request Automation
  tagline: Running Terraform Workflows with Ease
  image: /hero.png
  actions:
    - theme: brand
      text: Get Started
      link: /guide
    - theme: alt
      text: What is Atlantis?
      link: /blog/2017/introducing-atlantis
    - theme: alt
      text: Join us on Slack
      link: https://communityinviter.com/apps/cloud-native/cncf

features:
  - title: Fewer Mistakes
    details: "Catch errors in Terraform plan output before applying changes. Ensure changes are applied before merging."
    icon: ✅
  - title: Empower Developers
    details: "Developers can safely submit Terraform pull requests without credentials. Require approvals for applies."
    icon: 💻
  - title: Instant Audit Logs
    details: "Detailed logs for infrastructure changes, approvals, and user actions. Configure approvals for production changes."
    icon: 📋
  - title: Proven at Scale
    details: "Used by top companies to manage over 600 repos with 300 developers. In production since 2017."
    icon: 🌍
  - title: Self-Hosted
    details: "Your credentials remain secure. Deployable on VMs, Kubernetes, Fargate, etc. Supports GitHub, GitLab, Bitbucket, Azure DevOps."
    icon: ⚙️
  - title: Open Source
    details: "Atlantis is an open source project with strong community support, powered by volunteer contributions."
    icon: 🌐

---
