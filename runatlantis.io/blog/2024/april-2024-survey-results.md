---
title: Atlantis User Survey Results
lang: en-US
---

# Atlantis User Survey Results

In April 2024, the Core Atlantis Team launched an anonymous survey of our users. Over the two months the survey was open we received 354 responses, which we will use to better understand our community's needs and help prioritize our roadmap.

Overall, the results below show that we have a diverse set of enthusiastic users, and that though many are still the classic Atlantis setup (a handful of repos running terraform against AWS in GitHub), there are many different use cases and directions the community are going and would like to see Atlantis support.

We are grateful for everyone who took the time to share their experiences with Atlantis. We plan to run this kind of survey on a semi-regular basis, stay tuned!

## Anonymized Results

### How do you interact with Atlantis?

![](/blog/april-2024-survey-results/interact.webp)

Unsurprisingly, most users of Atlantis wear multiple hats, involved throughout the development process.

### How do you/your organization deploy Atlantis

![](/blog/april-2024-survey-results/deploy.webp)

Most users of terraform deploy using Kubernetes and/or AWS. "Other Docker" use docker but do not use EKS or Helm directly, while a minority use some other combination of technologies.

### What Infrastructure as Code (IaC) tool(s) do you use with Atlantis?

![](/blog/april-2024-survey-results/iac.webp)

The vast majority of Atlantis users are still using terraform as some part of their deployment. About half of them are in addition using Terragrunt, and OpenTofu seems to be gaining some ground.

### How many repositories does your Atlantis manage?

![](/blog/april-2024-survey-results/repos.webp)

Most users have relatively modest footprints to managed with Atlantis (though a few large monorepos could be obscured in the numbers).

### Which Version Control Systems (VCSs) do you use?

![](/blog/april-2024-survey-results/vcs.webp)

Most users of Atlantis are using GitHub, with a sizeable chunk on GitLab, followed by Bitbucket and others. This is analogous to the support and feature requests that the maintainers see for the various VCSs in the codebase.

### What is the most important feature you find missing from Atlantis?

![](/blog/april-2024-survey-results/features.webp)

This being a free form question, there was a long tail of responses, so the above only shows answers after normalizing that had three or more instances.

Drift Detection as well as infrastructure improvements were the obvious winners here. After that, users focused on various integrations and improvements to the UI.

## Conclusion

It is always interesting and exciting for the core team to see the breadth of the use of Atlantis, and we look forward to using this information to understand the needs of the community. Atlantis has always been a community led effort, and we hope to continue to carry that spirit forward!
