<template>
  <div class="container">
    <div class="home">
      <div class="hero">
        <img v-if="data.heroImage" :src="$withBase(data.heroImage)" alt="hero">
        <h1>{{ data.heroText || $title || 'Hello' }}</h1>
        <p class="description">
          Start working on Terraform <strong>as a team</strong>.
        </p>
        <p class="action" v-if="data.actionText && data.actionLink">
          <a href="/guide/" class="nav-link action-button">Get Started →</a>
        </p>
      </div>
      <!--<div class="companies">-->
      <!--<h2>Used By</h2>-->
      <!--<div class="company-list">-->
      <!--<div class="company">Nike</div>-->
      <!--<div class="company">Samsung</div>-->
      <!--<div class="company">Lyft</div>-->
      <!--<div class="company -last">Hootsuite</div>-->
      <!--</div>-->
      <!--</div>-->
    </div>
    <div class="workflow-container">
      <div class="workflow">
        <h1>The Atlantis Workflow</h1>
        <img src="/mobile-workflow.png" class="mobile" alt="Atlantis Workflow">
        <img src="/workflow.png" class="desktop" alt="Atlantis Workflow">
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
              <p>Bring the benefits of code review to your operations workflow.</p>
              <ul>
                <li><img class="checkmark" src="/checkmark.svg">Catch errors in the Terraform plan output before it's applied.
                </li>
                <li><img class="checkmark" src="/checkmark.svg">Ensure that you
                  apply changes before merging to master.
                </li>
              </ul>
            </div>
          </div>
          <div class="item image">
            <div class="image">
              <img src="/mistakes.png">
            </div>
          </div>
        </div>
      </div>
      <div class="benefit-container">
        <div class="benefit">
          <div class="item image">
            <div class="image">
              <img src="/devops.png">
            </div>
          </div>
          <div class="item">
            <div class="description">
              <h2>Put the "Dev" back into DevOps</h2>
              <p>Empower your developers to write Terraform. Safely.</p>
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
              <h2>Free Audit Logs</h2>
              <p>All production changes are recorded on the pull request where
                they were made.</p>
              <ul>
                <li><img class="checkmark" src="/checkmark.svg">See exactly
                  which resources were changed.
                </li>
                <li><img class="checkmark" src="/checkmark.svg">Know who apply'd
                  the change and who approved it.
                </li>
              </ul>
            </div>
          </div>
          <div class="item image">
            <div class="image">
              <img src="/logs.png">
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="benefits-container">
      <div class="benefit-container">
        <h1 class="title">How It Works</h1>
        <div class="benefit">
          <div class="item image">
            <div class="image">
              <img src="/hero.png">
            </div>
          </div>
          <div class="item">
            <div class="description">
              <ul>
                <li><img class="checkmark" src="/checkmark.svg">Atlantis is self-hosted. Your credentials don't leave your infrastructure.
                </li>
                <li><img class="checkmark" src="/checkmark.svg">Runs as a Golang binary or Docker image and can be deployed on VMs, Kubernetes, Fargate, etc.
                </li>
                <li><img class="checkmark" src="/checkmark.svg">Listens for webhooks from GitHub/GitLab/Bitbucket
                </li>
                <li><img class="checkmark" src="/checkmark.svg">Runs terraform commands remotely and comments back with their output
                </li>
              </ul>
            </div>
          </div>
        </div>
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

<style lang="stylus">
  // colors
  $accentColor = #3eaf7c
  $textColor = #2c3e50
  $borderColor = #eaecef
  $codeBgColor = #282c34
  $arrowBgColor = #ccc

  // layout
  $navbarHeight = 3.6rem
  $sidebarWidth = 20rem
  $contentWidth = 740px

  // responsive breakpoints
  $MQNarrow = 959px
  $MQMobile = 719px
  $MQMobileNarrow = 419px

  @import '~@temp/override.styl' // generated from user config

  $homeWidth = 960px
  $buttonColor = #34be68
  $darkBackground = rgb(250, 250, 250)
  .home
    padding 0 2rem
    max-width $homeWidth
    margin 0px auto 80px
    .hero
      text-align center
      img
        max-height 280px
        display block
        margin 3rem auto 1.5rem
      h1
        font-size 3rem
      h1, .description, .action
        margin 1.8rem auto
      .description
        max-width 35rem
        font-family Lato, sans-serif
        font-size 1.9rem
        line-height 1.3
        color lighten($textColor, 20%)
      .action
        display inline
      .action-button
        display inline-block
        font-size 1.2rem
        color #fff
        background-color $buttonColor
        padding 0.8rem 1.6rem
        border-radius 4px
        transition background-color .1s ease
        box-sizing border-box
        border-bottom 1px solid darken($buttonColor, 10%)
        margin: 0 10px
        &:hover
          background-color lighten($buttonColor, 10%)
        &.-nofill
          color darken($buttonColor, 10%)
          background-color white
          border 2px solid darken($buttonColor, 10%)
          padding 0.7rem 1.5rem
          &:hover
            color #fff
            background-color $buttonColor
    h2
      border-bottom none
    .companies
      text-align center
      margin-top: 40px
      h2
        color lighten($textColor, 20%)
        font-size 1.3rem
    .company-list
      display flex
      align-items center
      align-content stretch
      justify-content space-between
    .company
      flex-grow 1
      flex-basis 25%
      max-width 25%
      text-align center
      font-family Lato, sans-serif
      font-size 1.6rem
      line-height 1.3
      color lighten($textColor, 20%)
      border-right: 1px solid $borderColor
      &.-last
        border-right: none
    .features
      border-top 1px solid $borderColor
      padding 1.2rem 0
      margin-top 0
    .footer
      padding 2.5rem
      border-top 1px solid $borderColor
      text-align center
      color lighten($textColor, 25%)

  .workflow-container
    border-top 2px solid $borderColor

  .workflow
    text-align center
    margin: 80px auto
    max-width $homeWidth
    img {
      width 100%
    }
    .mobile {
      display none
    }
    .steps-container
      max-width $homeWidth
      margin 0 auto
      padding 0 10px
      .steps
        display flex
        flex-flow row nowrap
        justify-content flex-start
        .step
          text-align: center
          flex-grow 0
          flex-shrink 0
          flex-basis 192px
          &.-first
            padding-right 10px

  .benefits-container
    border-top 1px solid darken($darkBackground, 10%)
    .benefit-container
      border-bottom 1px solid darken($darkBackground, 10%)
      .title
        padding-top 40px
        text-align center
      &.-dark
        background-color rgb(250, 250, 250)
      &.-green
        background-color #61c589
        color white
      .benefit
        max-width $homeWidth
        margin: 0 auto
        display flex
        flex-flow row wrap
        align-items center
        .item
          flex-basis 50%
          flex-grow 1
          min-width 250px
          .image
            padding 40px
            text-align center
            img
              max-height 200px
        .description
          padding 40px
          h2
            border none
          ul
            list-style-type: none
            padding-left 0
          li
            display flex
            align-items center
            line-height 25px
            margin-bottom 20px
          .checkmark
            width 20px
            margin-right 10px
            vertical-align middle
            align-self baseline
            padding-top 5px

  @media (max-width: $MQMobile)
    .home
      .features
        flex-direction column
      .feature
        max-width 100%
        padding 0 2.5rem
    .workflow
      .mobile
        display block
      .desktop
        display none
    .benefits-container
      .benefit-container
        .benefit
          flex-direction column
          .item
            &.image
              order -1


  @media (max-width: $MQMobileNarrow)
    .home
      padding-left 1.5rem
      padding-right 1.5rem
      .hero
        img
          max-height 210px
          margin 2rem auto 1.2rem
        h1
          font-size 2rem
        h1, .description, .action
          margin 1.2rem auto
        .description
          font-size 1.2rem
        .action-button
          font-size 1rem
          padding 0.6rem 1.2rem
      .feature
        h2
          font-size 1.25rem
</style>