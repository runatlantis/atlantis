<template>
  <div class="container">
    <div class="home">
      <div class="hero">
        <img v-if="data.heroImage" :src="$withBase(data.heroImage)" alt="hero">
        <h1>{{ data.heroText || $title || 'Hello' }}</h1>
        <p class="description">
          Terraform Pull Request Automation
        </p>
        <p class="action" v-if="data.actionText && data.actionLink">
          <a href="/guide/" class="nav-link action-button">Get Started →</a>
        </p>
      </div>
    </div>
    <div class="workflow-container">
      <div class="workflow">
        <h1>The Atlantis Workflow</h1>
        <img src="/mobile-workflow-min.png" class="mobile" alt="Atlantis Workflow">
        <img src="/workflow-min.png" class="desktop" alt="Atlantis Workflow">
      </div>
    </div>
    <div class="benefits-container">
      <div class="benefit-container -dark">
        <div class="title">
          <h1>Benefits</h1>
        </div>
        <div class="benefit">
          <div class="item">
            <div class="description">
              <h2>Fewer Mistakes</h2>
              <p>Bring the benefits of <strong>code review</strong> to your operations
                workflow.</p>
              <ul>
                <li><img class="checkmark" src="/checkmark.svg">Catch errors in
                  the Terraform plan output before it's applied.
                </li>
                <li><img class="checkmark" src="/checkmark.svg">Ensure that you
                  apply changes before merging to main.
                </li>
              </ul>
            </div>
          </div>
          <div class="item image">
            <div class="image">
              <img src="/list.svg">
            </div>
          </div>
        </div>
      </div>
      <div class="benefit-container">
        <div class="benefit">
          <div class="item image">
            <div class="image">
              <img src="/coding.svg">
            </div>
          </div>
          <div class="item">
            <div class="description">
              <h2>Put the <span style="text-decoration: underline">Dev</span> back into DevOps</h2>
              <p>Empower your developers to write Terraform. <strong>Safely.</strong></p>
              <ul>
                <li><img class="checkmark" src="/checkmark.svg">Developers can
                  submit Terraform pull requests without needing credentials.
                </li>
                <li><img class="checkmark" src="/checkmark.svg">Operators can
                  require approvals prior to allowing an apply.
                </li>
              </ul>
            </div>
          </div>
        </div>
      </div>
      <div class="benefit-container -dark">
        <div class="benefit">
          <div class="item">
            <div class="description">
              <h2>Instant Audit Logs And Compliance</h2>
              <p>Pass audits without compromising your workflow.</p>
              <ul>
                <li><img class="checkmark" src="/checkmark.svg">Each pull request now holds a detailed log of what infrastructure changes were made and when; along with who made the change and who approved it.
                </li>
                <li><img class="checkmark" src="/checkmark.svg">Atlantis can be configured to require approvals on every production change.
                </li>
              </ul>
            </div>
          </div>
          <div class="item image">
            <div class="image">
              <img src="/certificate.svg">
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="benefits-container">
      <div class="benefit-container">
        <h1 class="title">Proven at Scale</h1>
        <div class="benefit">
          <div class="item image">
            <div class="image">
              <img src="/powerful.svg">
            </div>
          </div>
          <div class="item">
            <div class="description">
              <ul>
                <li><img class="checkmark" src="/checkmark.svg">Used by one of the world's top companies to manage over 600 Terraform repos with 300 developers.
                </li>
                <li><img class="checkmark" src="/checkmark.svg">In production use since 2017.
                </li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="benefits-container">
      <div class="benefit-container">
        <h1 class="title">How It Works</h1>
        <div class="benefit">
          <div class="item">
            <div class="description">
              <ul>
                <li><img class="checkmark" src="/checkmark.svg">Atlantis is
                  self-hosted. Your credentials don't leave your infrastructure.
                </li>
                <li><img class="checkmark" src="/checkmark.svg">Runs as a Golang
                  binary or Docker image and can be deployed on VMs, Kubernetes,
                  Fargate, etc.
                </li>
                <li><img class="checkmark" src="/checkmark.svg">Listens for
                  webhooks from GitHub/GitLab/Bitbucket/Azure DevOps.
                </li>
                <li><img class="checkmark" src="/checkmark.svg">Runs terraform
                  commands remotely and comments back with their output.
                </li>
              </ul>
            </div>
          </div>
          <div class="item image">
            <div class="image">
              <img src="/hero.png">
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="home getting-started-footer">
      <div class="hero">
        <p class="action" v-if="data.actionText && data.actionLink">
          <a href="/guide/" class="nav-link action-button">Get Started →</a>
        </p>
      </div>
    </div>
    <div class="footer" v-if="data.footer">
      {{ data.footer }}
    </div>
  </div>
</template>

<script>
    export default {
        computed: {
            data() {
                return this.$page.frontmatter
            },
            actionLink() {
                return {
                    link: this.data.actionLink,
                    text: this.data.actionText
                }
            }
        }
    }
</script>
