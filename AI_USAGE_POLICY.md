# Atlantis AI Usage Policy

**AI is welcome. Humans are responsible.**

Current AI tools are useful as coding assistants — but not as autonomous contributors. Anyone submitting content to Atlantis is fully responsible for the correctness, intent, testing, and licensing compliance of their contribution.

## General Guidelines

1. Contributors are fully responsible and accountable for all their submissions. This includes pull requests (PRs), issues, comments, or any other form of engagement with the project and its maintainers.

2. To avoid maintainer overload, contributors are limited to 8 open PRs at any given time. Contributors are expected to address all open PR review comments and work through open PRs before creating additional ones.

3. Contributors using AI to generate content should:

   * Thoroughly review all AI-generated content before submission
   * Understand the reason and impact of the changes
   * Refine AI output to meet project quality standards (see [CONTRIBUTING.md](CONTRIBUTING.md))
   * Take full ownership of all submitted content regardless of origin
   * Ensure content does not violate legal copyrights or other laws

4. Pull requests that include AI-generated code should only target issues that have been accepted (i.e., not labeled as "triage" or "needs-discussion").

5. Contributors MUST disclose any substantial use of AI. Disclosure MUST take the form of a trailer line within the commit attributing the AI tool used. Acceptable formats include:

   * `Assisted-by: Claude <noreply@anthropic.com>`
   * `Co-authored-by: Claude <noreply@anthropic.com>`
   * `Assisted-by: GitHub Copilot <noreply@github.com>`
   * `Co-authored-by: GitHub Copilot <noreply@github.com>`

   Many AI coding tools automatically add `Co-authored-by` trailers — this is acceptable and need not be changed to `Assisted-by`.

## AI Tooling in Atlantis Workflows

Atlantis maintainers use several AI tools as part of the development workflow. These tools may change over time as the ecosystem evolves.

### Dosu

[Dosu](https://dosu.dev/) is used in Issues to help provide users with context to the codebase. Feel free to ask it follow-up questions via `@dosu`, but note that it is only basing its answers on the code and history, so may lack high-level goals of the project.

### Copilot

[GitHub Copilot](https://github.com/features/copilot) is used in Pull Requests to summarize, as well as add comments and suggestions.

When responding to Copilot review comments, contributors should:

* Evaluate each suggestion on its technical merits
* Not blindly apply AI-suggested changes without understanding them
* Feel free to dismiss suggestions that are incorrect or not applicable, with a brief explanation

Reviewers will typically require that all comments be resolved before approving a pull request, including comments from Copilot. Contributors are expected to apply human judgment and may dismiss suggestions that are incorrect, irrelevant, or not useful, but comments should still be explicitly resolved. If these automated reviews stop providing enough value to justify the overhead, we may revisit this policy.

See [AGENTS.md](AGENTS.md) for guidance on how AI coding agents should interact with this repository.

## Legal and Licensing Considerations

Contributors must ensure:

* AI tool terms of service do not conflict with the project's [Apache 2.0 license](LICENSE)
* No copyrighted or improperly licensed material is included in contributions
* All third-party content is properly attributed
* The [Developer Certificate of Origin (DCO)](https://developercertificate.org/) can be truthfully signed (`git commit -s`)

Contributors must also comply with their employer's policies regarding AI-assisted open source work.

Atlantis follows CNCF / Linux Foundation guidance on AI-assisted development.

## Policy Evolution

This policy will be reviewed and updated as needed to reflect:

* Changes in AI tooling and its use across open source projects
* Legal or regulatory developments
* Maintainer and reviewer experience together with community feedback
* Evolution of CNCF and industry best practices

## Questions and Feedback

Please share feedback and any questions or concerns about this policy — including areas that feel too strict or too permissive, enforcement concerns, or gaps related to Atlantis-specific workflows:

* Open an issue in the [Atlantis repository](https://github.com/runatlantis/atlantis/issues)
* Discuss in the [CNCF Slack](https://slack.cncf.io/) channels **#atlantis** and **#atlantis-contributors**

## References

* [CNCF / Linux Foundation: Guidance Regarding Use of Generative AI Tools for Open Source Software Development](https://www.linuxfoundation.org/blog/generative-ai-tools)
* [Kyverno AI Usage Policy](https://github.com/kyverno/community/blob/main/AI_USAGE_POLICY.md)
* [Developer Certificate of Origin](https://developercertificate.org/)
* [AGENTS.md](AGENTS.md)
