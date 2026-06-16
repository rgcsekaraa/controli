import { defineConfig } from 'vitepress'

const siteUrl = 'https://rgcsekaraa.github.io/controli/'

export default defineConfig({
  title: 'Controli',
  description: 'Native Go CLI sharing over an outbound Cloudflare relay.',
  base: '/controli/',
  cleanUrls: true,
  lastUpdated: true,
  sitemap: {
    hostname: siteUrl
  },
  head: [
    ['meta', { name: 'theme-color', content: '#0f766e' }],
    ['meta', { name: 'keywords', content: 'Go CLI sharing, terminal sharing, remote terminal, Cloudflare relay, WebSocket terminal' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:title', content: 'Controli Documentation' }],
    ['meta', { property: 'og:description', content: 'Native Go CLI sharing over an outbound Cloudflare relay.' }],
    ['meta', { property: 'og:url', content: siteUrl }],
    ['meta', { property: 'og:site_name', content: 'Controli' }],
    ['meta', { name: 'twitter:card', content: 'summary_large_image' }],
    ['meta', { name: 'twitter:title', content: 'Controli Documentation' }],
    ['meta', { name: 'twitter:description', content: 'Native Go CLI sharing over an outbound Cloudflare relay.' }],
    ['link', { rel: 'canonical', href: siteUrl }]
  ],
  themeConfig: {
    logo: { src: '/logo.svg', alt: 'Controli' },
    siteTitle: 'Controli',
    nav: [
      { text: 'Guide', link: '/overview' },
      { text: 'Install', link: '/install' },
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
