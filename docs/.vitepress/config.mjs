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
            { text: "Templates", link: "/templates" },
            { text: "Dependencies", link: "/dependencies" },
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
    ],
  })
);
