import { defineConfig } from "vitepress";
import { withMermaid } from "vitepress-plugin-mermaid";

// https://vitepress.dev/reference/site-config
export default withMermaid(
  defineConfig({
    title: "Tenant Operator",
    description: "Multi-Tenant Kubernetes Automation Platform",
    base: "/",
    srcDir: ".",
    ignoreDeadLinks: true,

    themeConfig: {
      // https://vitepress.dev/reference/default-theme-config
      logo: "/logo.png",

      nav: [
        { text: "Home", link: "/" },
        { text: "Getting Started", link: "/quickstart" },
        { text: "Guide", link: "/api" },
        { text: "Reference", link: "/glossary" },
      ],

      sidebar: [
        {
          text: "Getting Started",
          collapsed: false,
          items: [
            { text: "Installation", link: "/installation" },
            { text: "Quick Start", link: "/quickstart" },
            {
              text: "Local Development (Minikube)",
              link: "/local-development-minikube",
            },
          ],
        },
        {
          text: "Core Concepts",
          collapsed: false,
          items: [
            { text: "API Reference", link: "/api" },
            { text: "Datasources", link: "/datasource" },
            { text: "Templates", link: "/templates" },
            { text: "Dependencies", link: "/dependencies" },
            { text: "Policies", link: "/policies" },
          ],
        },
        {
          text: "Configuration",
          collapsed: false,
          items: [{ text: "Configuration Guide", link: "/configuration" }],
        },
        {
          text: "Operations",
          collapsed: false,
          items: [
            { text: "Monitoring & Observability", link: "/monitoring" },
            { text: "Performance Tuning", link: "/performance" },
            { text: "Security", link: "/security" },
            { text: "Troubleshooting", link: "/troubleshooting" },
          ],
        },
        {
          text: "Integrations",
          collapsed: false,
          items: [
            { text: "External DNS", link: "/integration-external-dns" },
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
            { text: "Development Guide", link: "/development" },
            { text: "Roadmap", link: "/roadmap" },
          ],
        },
        {
          text: "Reference",
          collapsed: false,
          items: [
            { text: "Glossary", link: "/glossary" },
            { text: "Prometheus Queries", link: "/prometheus-queries" },
          ],
        },
      ],

      socialLinks: [
        {
          icon: "github",
          link: "https://github.com/kubernetes-tenants/tenant-operator",
        },
      ],

      search: {
        provider: "local",
      },

      footer: {
        message:
          '<p style="margin-bottom: 12px">Released under the Apache 2.0 License.<br />Built with ❤️ using Kubebuilder, Controller-Runtime, and VitePress.</p>',
        copyright: "Copyright © 2025 Tenant Operator",
      },

      editLink: {
        pattern:
          "https://github.com/kubernetes-tenants/tenant-operator/edit/main/docs/:path",
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
          isCustomElement: (tag) => false,
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
    ],
  })
);
