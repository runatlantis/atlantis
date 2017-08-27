# kube Theme for Hugo

`kube` Kube is a professional  and a responsive Hugo theme for developers and designers that offers a documentation section mixed with a landing page and a blog.

I create this theme  based on the `Version 6.5.2` [Kube Framework](https://imperavi.com/kube/). 

![kube hugo landingPage](https://cldup.com/RjWtdJZNae.png)

# Demo

To see this theme in action, check out [kube project](http://kube.elemnts.org) which is rendered with this theme and some conetnt for documentation and blog posts.

## Features

- Mobile-first Design : Every element in kube is mobile-first and fully embraces latest and greatest tech.
- Responsive Design : Optimized for mobile, tablet, desktop
- Horizontal Rhythm : Like Kube framework this theme is based on a 4px vertical grid.
- Typography : beautiful typographie choice
- Google Analytics : Google Analytics using the internal async template
- Disqus Commenting : Post comments with Disqus using the internal template
- OpenGraph support : SEO-optimized using OpenGraph
- Schema Structured Data : Schema Structured Data and Meta tags
- Paginated Lists : Simple list pagination with page indicators
- Reading Time : Post reading time and update notice set user expectations
- Meta data for all blog article : Rich post data including links to category and tag taxonomy listings, author and word count
- Related Posts : Related Content for increased page views and reader loyalty
- Block Templates : Block Templates for foolproof layout extensions
- Table of Contents : Accessible Table of Contents for documentation
- SEO Site Verification : Site verification with Google, Bing Alexa and Yandex
- 404 page : 404 page with animated background

## Installation

Inside the folder of your Hugo site run:

    $ mkdir themes
    $ cd themes
    $ git clone https://github.com/jeblister/kube.git

For more information read the official [setup guide](//gohugo.io/overview/installing/) for Hugo.


Copy custom archetypes to your site:

```shell
cp themes/kube/archetypes/* archetypes
```


Next, take a look in the `exampleSite` folder at. This directory contains an example config file and the content for the demo. It serves as an example setup for your blog. 

Copy at least the `config.toml` in the root directory of your website. Overwrite the existing config file if necessary. 

Hugo includes a development server, so you can view your changes as you go :

``` sh
hugo server -w
```

Now you can go to [localhost:1313](http://localhost:1313) and the `kube`
theme should be visible.


## Getting Started

There are a few concepts this theme employs to make a personal documentation site. It's important to read this as you may not see what you expect upon launching. It assumes you want to call your documentation posts `docs` and organizes them as such. For example, creating a new docs with Hugo would require you typing:

```
  $ hugo new --kind docs docs/my-new-doc.md

```

It also assumes you want to display three types of content `docs` and `blog` and some pages : the `faq`, `company` and `sign-in` pages and and display links to this pages in the menu. This guide will take you through the steps to configure your documentation site to use the theme.

### Configuring you website

#### Where should blog post markdown files be stored?

The theme works with other content types, but docs pages work best when grouped under `docs`. When using the `docs` content type you'll have a customized list page sorted by `weight` and the default list page for all documentation. Here's an example:

![Custom List docs Page](https://cldup.com/8k1nU8TLuU.png)



#### Defining yourself as the Author

In this case you would want to add `author = "your name"` variable like your name to your post's Front Matter.


#### Webmaster Verification

Verify your site with several webmaster tools including Google, Bing, Alexa and Yandex. To allow verification of your site with any or all of these providers simply add the following to your `config.toml` and fill in their respective values:

```toml
[params.seo.webmaster_verifications]
  google = "" # Optional, Google verification code
  bing = "" # Optional, Bing verification code
  alexa = "" # Optional, Alexa verification code
  yandex = "" # Optional, Yandex verification code
```

### Index Blocking

Just because a page appears in your `sitemap.xml` does not mean you want it to appear in a SERP. Examples of pages which will appear in your `sitemap.xml` that you typically do not want indexed by crawlers include error pages, search pages, legal pages, and pages that simply list summaries of other pages.

Though it's possible to block search indexing from a `robots.txt` file, kube makes it possible to block page indexing using Hugo configuration as well. By default the following page types will be blocked:

- Section Pages (e.g. Post listings)
- Taxonomy Pages (e.g. Category and Tag listings)
- Taxonomy Terms Pages (e.g. Pages listing taxonomies)

To customize default blocking configure the `noindex_kinds` setting in the `[params]` section of your `config.toml`. For example, if you want to enable crawling for sections appearing in [Section Menu](#adding-a-section-menu) add the following to your configuration file:

```
[params]
  noindex_kinds = [
    "taxonomy",
    "taxonomyTerm"
  ]
```

To block individual pages from being indexed add `nofollow` to your page's front matter and set the value to `true`, like:

```toml
noindex = true
```

And, finally, if you're using Hugo `v0.18` or better, you can also add an `_index.md` file with the `noindex` front matter to control indexing for specific section list layouts:

```shell
├── content
│   ├── modules
│   │   ├── starry-night.md
│   │   └── flying-toilets.md
│   └── news
│       ├── _index.md
│       └── return-flying-toasters.md
```

To learn more about how crawlers use this feature read [block search indexing with meta tags](https://support.google.com/webmasters/answer/93710).

### Custom CSS

To add your own theme css or override existing CSS without having to change theme files do the following:

1. Create a `style.css` in your site's `layouts/static/css directory` or use `custom.css` file in 'themes/kube/static/css/custom.css`
1. Add link to this file in 'themes/kube/layouts/_default/baseof.html'.

Default `style block` :

```html
<!-- Your own theme here -->
 <link href="/css/custom.css" rel="stylesheet" type="text/css">

```


## Contributing

Did you find a bug or have an ideas for new features? Feel free to use the issue tracker to let me know or make a pull request.

There's only one rule...there are no rules.

## License

MIT

## Credit 

- [kube framework] (https://imperavi.com/kube/)
- [after dark theme] (https://github.com/comfusion/after-dark)

## Contact

This is the second theme I've made for Hugo, so I'm sure I've done some things wrong or assumed too much. If you have ideas or things that should be fixed, please let me know.

- [Mohamed JEBLI](http://findme.surge.sh) [@jebli_7](http://twitter.com/jebli_7)
