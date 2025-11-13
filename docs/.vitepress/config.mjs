import { defineConfig } from "vitepress";
import { withMermaid } from "vitepress-plugin-mermaid";

// https://vitepress.dev/reference/site-config
export default withMermaid(
  defineConfig({
    title: "Lynq",
    description: "Database-Driven Kubernetes Automation Platform",
    base: "/",
    srcDir: ".",
    ignoreDeadLinks: false,
    appearance: 'force-dark', // Force dark mode only

    themeConfig: {
      // https://vitepress.dev/reference/default-theme-config
      logo: "/logo.png",

      nav: [
        { text: "Home", link: "/" },
        { text: "Quick Start", link: "/quickstart" },
        { text: "Documentation", link: "/architecture" },
        { text: "API", link: "/api" },
      ],

      sidebar: [
        {
          text: "Getting Started",
          collapsed: false,
          items: [
            { text: "Installation", link: "/installation" },
            { text: "Quick Start", link: "/quickstart" },
          ],
        },
        {
          text: "Core Concepts",
          collapsed: false,
          items: [
            { text: "Architecture", link: "/architecture" },
            { text: "API Reference", link: "/api" },
            { text: "Configuration", link: "/configuration" },
            { text: "Datasources", link: "/datasource" },
            {
              text: "Templates",
              collapsed: false,
              items: [
                { text: "Overview", link: "/templates" },
                { text: "🛠️ Builder", link: "/template-builder" },
              ],
            },
            {
              text: "Dependencies",
              collapsed: false,
              items: [
                { text: "Overview", link: "/dependencies" },
                { text: "🔍 Visualizer", link: "/dependency-visualizer" },
              ],
            },
            {
              text: "Policies",
              collapsed: false,
              items: [
                { text: "Overview", link: "/policies" },
                { text: "Examples", link: "/policies-examples" },
                { text: "Field-Level Ignore", link: "/field-ignore" },
              ],
            },
          ],
        },
        {
          text: "Advanced Use Cases",
          collapsed: false,
          items: [
            { text: "Overview", link: "/advanced-use-cases" },
            { text: "Custom Domains", link: "/use-case-custom-domains" },
            { text: "Multi-Tier Stack", link: "/use-case-multi-tier" },
            { text: "Blue-Green Deployments", link: "/use-case-blue-green" },
            { text: "Database-per-Tenant", link: "/use-case-database-per-tenant" },
            { text: "Feature Flags", link: "/use-case-feature-flags" },
          ],
        },
        {
          text: "Operations",
          collapsed: false,
          items: [
            {
              text: "Monitoring & Observability",
              link: "/monitoring",
            },
            { text: "Prometheus Queries", link: "/prometheus-queries" },
            { text: "Performance Tuning", link: "/performance" },
            { text: "Security", link: "/security" },
            { text: "Troubleshooting", link: "/troubleshooting" },
            { text: "Alert Runbooks", link: "/alert-runbooks" },
          ],
        },
        {
          text: "Integrations",
          collapsed: false,
          items: [
            { text: "Crossplane (Recommended)", link: "/integration-crossplane" },
            { text: "External DNS (Recommended)", link: "/integration-external-dns" },
            {
              text: "Terraform Operator",
              link: "/integration-terraform-operator",
            },
            {
              text: "Argo CD",
              link: "/integration-argocd",
            },
          ],
        },
        {
          text: "Development",
          collapsed: false,
          items: [
            {
              text: "Local Development",
              link: "/local-development-minikube",
            },
            { text: "Development Guide", link: "/development" },
            {
              text: "Contributing",
              link: "/contributing-datasource",
            },
            { text: "Roadmap", link: "/roadmap" },
          ],
        },
        {
          text: "Glossary",
          link: "/glossary",
        },
      ],

      socialLinks: [
        {
          icon: "github",
          link: "https://github.com/k8s-lynq/lynq",
        },
      ],

      search: {
        provider: "local",
      },

      footer: {
        message:
          '<p style="margin-bottom: 12px">Released under the Apache 2.0 License.<br />Built with ❤️ using Kubebuilder, Controller-Runtime, and VitePress.</p>',
        copyright: "Copyright © 2025 Lynq",
      },

      editLink: {
        pattern:
          "https://github.com/k8s-lynq/lynq/edit/main/docs/:path",
        text: "Edit this page on GitHub",
      },

      lastUpdated: {
        text: "Updated at",
        formatOptions: {
          dateStyle: "full",
          timeStyle: "medium",
        },
      },
    },

    markdown: {
      theme: "github-dark",
      lineNumbers: true,
    },

    mermaid: {
      // Mermaid configuration options
    },

    mermaidPlugin: {
      class: "mermaid my-class", // set additional css classes for parent container
    },

    vue: {
      template: {
        compilerOptions: {
          isCustomElement: () => false,
        },
      },
    },

    vite: {
      optimizeDeps: {
        exclude: [],
      },
    },

    head: [
      ["link", { rel: "icon", type: "image/png", href: "/logo.png" }],
      ["link", { rel: "shortcut icon", href: "/logo.ico" }],
      ["link", { rel: "apple-touch-icon", href: "/logo.ico" }],
      ["meta", { name: "google-site-verification", content: "g7LPr3Wcm6hCm-Lm8iP5KVl11KvPv6Chxpjh3oNKHPw" }],
    ],
  })
);
