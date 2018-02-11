+++
title = "Moving Atlantis To atlantisnorth"
date = "2018-02-07T14:01:11-08:00"
draft = false
weight = 20
description = "Why Atlantis is moving to github.com/atlantisnorth/atlantis"
+++

<p style="text-align: center">
<img src="/img/luke.png" style="max-height: 250px">
</p>
Hi, my name is Luke and I'm the maintainer of Atlantis.

I wanted to explain why I'm going to be continuing work on Atlantis
under [github.com/atlantisnorth/atlantis](https://github.com/atlantisnorth/atlantis) instead of under [github.com/hootsuite/atlantis](https://github.com/hootsuite/atlantis).

## Short Story Of Atlantis
Atlantis was originally created by Anubhav Mishra while he worked at [Hootsuite](https://hootsuite.com) (you can read more about why he created Atlantis [here](/blog/atlantis-release)).
Mishra and I worked together at Hootsuite and after realizing that Atlantis would be a useful tool for others, we worked together to rewrite it in Golang and open source it.
This resulted in [github.com/hootsuite/atlantis](https://github.com/hootsuite/atlantis).

## Where We Are Now
Since then, Mishra has left Hootsuite to work at HashiCorp and I've left Hootsuite as well.

I (Luke) want to continue working on Atlantis but doing so under the `hootsuite` GitHub organization as the sole maintainer is problematic because:

* I don't have administrative control over the repo because I'm no longer in the `hootsuite` organization. This means I need permission to add apps like [Codecov.io](https://codecov.io) and [CircleCI](https://circleci.com/).
* It also means I **could** lose control over the repo if Hootsuite's open source policies changed. I don't think this would happen but it is a risk.

Because Hootsuite is heavily invested in Atlantis–it's used every day by 20 operations developers and over 50 developers to manage 100+ Terraform repos–they want the project
to continue developing. As such, Hootsuite and I came to the mutual conclusion that it would be best for me to work on Atlantis under my own organization. Thus: [github.com/atlantisnorth/atlantis](https://github.com/atlantisnorth/atlantis).

## Future of Atlantis
I'm really excited to continue working on Atlantis so that it can be used by more teams and support more use-cases. Please create any issues and pull requests under the new repository: [github.com/atlantisnorth/atlantis](https://github.com/atlantisnorth/atlantis).
I'll be copying issues over from `hootsuite/atlantis` to `atlantisnorth/atlantis` but that repo will not be maintained further.

If you have any questions, please DM me on Twitter at [@lkysow](https://twitter.com/lkysow).
