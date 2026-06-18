import { defineConfig } from 'vitepress'

const siteUrl = 'https://rgcsekaraa.github.io/controli/'
const basePath = '/controli/'
const siteName = 'Controli'
const siteTitle = 'Controli Documentation'
const siteDescription = 'CLI sharing for support sessions through an outbound Cloudflare relay.'
const ogImage = `${siteUrl}og-image.png`

function pageUrl(page: string): string {
  const clean = page.replace(/(^|\/)index\.md$/, '').replace(/\.md$/, '')
  return new URL(clean, siteUrl).href
}

export default defineConfig({
  title: siteName,
  titleTemplate: ':title | Controli Docs',
  description: siteDescription,
  base: basePath,
  cleanUrls: true,
  appearance: 'force-auto',
  lastUpdated: true,
  sitemap: {
    hostname: siteUrl
  },
  head: [
    ['meta', { name: 'application-name', content: siteName }],
    ['meta', { name: 'apple-mobile-web-app-title', content: siteName }],
    ['meta', { name: 'theme-color', media: '(prefers-color-scheme: light)', content: '#ffffff' }],
    ['meta', { name: 'theme-color', media: '(prefers-color-scheme: dark)', content: '#17181d' }],
    ['meta', { name: 'keywords', content: 'CLI sharing, terminal sharing, remote terminal, Go CLI, Cloudflare relay, WebSocket terminal, support sessions' }],
    ['meta', { name: 'robots', content: 'index,follow' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:site_name', content: siteName }],
    ['meta', { property: 'og:locale', content: 'en_US' }],
    ['meta', { property: 'og:image', content: ogImage }],
    ['meta', { property: 'og:image:width', content: '1200' }],
    ['meta', { property: 'og:image:height', content: '630' }],
    ['meta', { property: 'og:image:alt', content: 'Controli documentation preview' }],
    ['meta', { name: 'twitter:card', content: 'summary_large_image' }],
    ['meta', { name: 'twitter:image', content: ogImage }],
    ['meta', { name: 'twitter:image:alt', content: 'Controli documentation preview' }],
    ['link', { rel: 'icon', type: 'image/svg+xml', href: `${basePath}favicon.svg` }],
    ['link', { rel: 'icon', href: `${basePath}favicon.ico`, sizes: 'any' }],
    ['link', { rel: 'alternate icon', type: 'image/png', sizes: '32x32', href: `${basePath}favicon-32.png` }],
    ['link', { rel: 'apple-touch-icon', sizes: '180x180', href: `${basePath}apple-touch-icon.png` }],
    ['link', { rel: 'manifest', href: `${basePath}site.webmanifest` }]
  ],
  transformHead({ page, title, description }) {
    const url = pageUrl(page)
    const metaTitle = title || siteTitle
    const metaDescription = description || siteDescription

    return [
      ['link', { rel: 'canonical', href: url }],
      ['meta', { property: 'og:title', content: metaTitle }],
      ['meta', { property: 'og:description', content: metaDescription }],
      ['meta', { property: 'og:url', content: url }],
      ['meta', { name: 'twitter:title', content: metaTitle }],
      ['meta', { name: 'twitter:description', content: metaDescription }]
    ]
  },
  themeConfig: {
    logo: { src: '/logo.svg', alt: 'Controli' },
    siteTitle: 'Controli',
    nav: [
      { text: 'Guide', link: '/overview' },
      { text: 'Install', link: '/install' },
      { text: 'Commands', link: '/commands' },
      { text: 'Tunnel', link: '/tunnel' },
      { text: 'Security', link: '/security' },
      { text: 'GitHub', link: 'https://github.com/rgcsekaraa/controli' }
    ],
    sidebar: [
      {
        text: 'Start',
        items: [
          { text: 'Introduction', link: '/' },
          { text: 'Overview', link: '/overview' },
          { text: 'Install', link: '/install' },
          { text: 'Commands', link: '/commands' },
          { text: 'Tunnel Mode', link: '/tunnel' },
          { text: 'Compatibility', link: '/compatibility' },
          { text: 'Configuration', link: '/configuration' },
          { text: 'State', link: '/state' }
        ]
      },
      {
        text: 'Usage',
        items: [
          { text: 'Host', link: '/host' },
          { text: 'Join', link: '/join' },
          { text: 'Windows', link: '/windows' },
          { text: 'Linux', link: '/linux' },
          { text: 'macOS', link: '/macos' }
        ]
      },
      {
        text: 'Architecture',
        items: [
          { text: 'Relay', link: '/relay' },
          { text: 'Protocol', link: '/protocol' },
          { text: 'Security', link: '/security' },
          { text: 'Troubleshooting', link: '/troubleshooting' }
        ]
      },
      {
        text: 'Project',
        items: [
          { text: 'Build', link: '/build' },
          { text: 'Release', link: '/release' },
          { text: 'Roadmap', link: '/roadmap' }
        ]
      }
    ],
    socialLinks: [
      { icon: 'github', link: 'https://github.com/rgcsekaraa/controli' }
    ],
    search: {
      provider: 'local'
    },
    editLink: {
      pattern: 'https://github.com/rgcsekaraa/controli/edit/main/docs/:path',
      text: 'Edit this page'
    },
    footer: {
      message: 'Released under the MIT license.',
      copyright: 'Copyright © 2026 Chan RG'
    }
  }
})
