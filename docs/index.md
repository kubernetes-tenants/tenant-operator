---
layout: home
---

<style scoped>
.custom-hero {
  position: relative;
  width: 100vw;
  margin-left: 50%;
  transform: translateX(-50%);
  padding: 5rem 2rem 4rem;
  overflow: hidden;
}

.custom-hero::before {
  content: '';
  position: fixed;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 100%;
  height: 100%;
  z-index: -1;
  background: radial-gradient(
    ellipse 35% 28% at 50% 25%,
    rgba(100, 108, 255, 0.5) 0%,
    rgba(88, 96, 224, 0.3) 30%,
    rgba(74, 82, 196, 0.18) 50%,
    transparent 70%
  );
  opacity: 1;
  pointer-events: none;
  animation: gradientPulse 8s ease-in-out infinite;
}

/* Light mode - much more visible gradient */
html:not(.dark) .custom-hero::before {
  background: radial-gradient(
    ellipse 40% 32% at 50% 25%,
    rgba(100, 108, 255, 0.35) 0%,
    rgba(88, 96, 224, 0.22) 30%,
    rgba(74, 82, 196, 0.12) 50%,
    rgba(74, 82, 196, 0.04) 65%,
    transparent 80%
  );
}

@keyframes gradientPulse {
  0%, 100% {
    opacity: 0.5;
    transform: translateX(-50%) scale(0.95);
  }
  50% {
    opacity: 1.2;
    transform: translateX(-50%) scale(1.15);
  }
}

.hero-content {
  position: relative;
  max-width: 1200px;
  margin: 0 auto;
  text-align: center;
  z-index: 1;
}

.hero-tagline {
  font-size: clamp(2.5rem, 7vw, 4.5rem);
  font-weight: 800;
  line-height: 1.15;
  margin: 0 0 1.5rem;
  color: var(--vp-c-text-1);
  letter-spacing: -0.02em;
  opacity: 0;
  animation: fadeInUp 1s ease-out 0.1s forwards;
}

.hero-description {
  font-size: clamp(1.15rem, 2.5vw, 1.5rem);
  color: var(--vp-c-text-2);
  margin: 0 0 2.5rem;
  max-width: 800px;
  margin-left: auto;
  margin-right: auto;
  line-height: 1.6;
  font-weight: 500;
  opacity: 0;
  animation: fadeInUp 1s ease-out 0.3s forwards;
}

.hero-actions {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
  margin-bottom: 3rem;
  opacity: 0;
  animation: fadeInUp 1s ease-out 0.5s forwards;
}

.hero-action {
  position: relative;
  display: inline-block;
  padding: 1rem 2.5rem;
  border-radius: 12px;
  font-weight: 600;
  font-size: 1.05rem;
  text-decoration: none;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
  z-index: 1;
}

.hero-action.brand {
  background: linear-gradient(
    135deg,
    rgba(102, 126, 234, 0.15) 0%,
    rgba(118, 75, 162, 0.15) 100%
  );
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.18);
  backdrop-filter: blur(20px) saturate(180%);
  -webkit-backdrop-filter: blur(20px) saturate(180%);
  box-shadow:
    0 8px 32px rgba(102, 126, 234, 0.2),
    inset 0 1px 0 rgba(255, 255, 255, 0.15),
    inset 0 -1px 0 rgba(0, 0, 0, 0.1);
  position: relative;
  isolation: isolate;
}

.hero-action.brand::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 12px;
  background: linear-gradient(
    180deg,
    rgba(255, 255, 255, 0.12) 0%,
    rgba(255, 255, 255, 0.02) 50%,
    rgba(0, 0, 0, 0.05) 100%
  );
  pointer-events: none;
  z-index: 1;
}

.hero-action.brand::after {
  content: '';
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: conic-gradient(
    from var(--gradient-angle),
    transparent 0deg,
    rgba(102, 126, 234, 0.5) 60deg,
    rgba(118, 75, 162, 0.5) 120deg,
    rgba(240, 147, 251, 0.5) 180deg,
    transparent 240deg
  );
  animation: rotateGradient 4s linear infinite;
  opacity: 0;
  transition: opacity 0.4s ease;
  filter: blur(15px);
  z-index: 0;
}

.hero-action.brand:hover::after {
  opacity: 1;
}

.hero-action.brand:hover {
  transform: translateY(-3px);
  border-color: rgba(255, 255, 255, 0.3);
  background: linear-gradient(
    135deg,
    rgba(102, 126, 234, 0.25) 0%,
    rgba(118, 75, 162, 0.25) 100%
  );
  box-shadow:
    0 12px 40px rgba(102, 126, 234, 0.35),
    inset 0 1px 0 rgba(255, 255, 255, 0.2),
    inset 0 -1px 0 rgba(0, 0, 0, 0.1);
}

.hero-action.alt {
  background: rgba(255, 255, 255, 0.03);
  color: var(--vp-c-text-1);
  border: 2px solid rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
}

.hero-action.alt:hover {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(102, 126, 234, 0.5);
  transform: translateY(-3px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);
}

@property --gradient-angle {
  syntax: '<angle>';
  initial-value: 0deg;
  inherits: false;
}

@keyframes rotateGradient {
  0% {
    --gradient-angle: 0deg;
  }
  100% {
    --gradient-angle: 360deg;
  }
}

.hero-diagram {
  opacity: 0;
  animation: fadeInUp 1s ease-out 0.7s forwards;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 768px) {
  .custom-hero {
    padding: 3rem 1.5rem 2.5rem;
  }

  .hero-actions {
    flex-direction: column;
    align-items: stretch;
  }

  .hero-action {
    text-align: center;
  }
}
</style>

<div class="custom-hero">
<div class="hero-content">
<h1 class="hero-tagline">Database-Driven<br/>Kubernetes Automation</h1>
<p class="hero-description">
Turn database rows into production-ready infrastructure.<br/>
Automatically.
</p>

<div class="hero-actions">
<a href="/quickstart" class="hero-action brand">Get Started</a>
<a href="https://github.com/kubernetes-tenants/tenant-operator" class="hero-action alt" target="_blank" rel="noopener noreferrer">View on GitHub</a>
</div>

<div class="hero-diagram">
<AnimatedDiagram />
</div>
</div>
</div>

## Why Tenant Operator?

<p style="font-size: 1.05rem; color: var(--vp-c-text-2); margin: 1.5rem 0">
Managing hundreds or thousands of tenants in Kubernetes shouldn't require custom scripts, manual updates, or rebuilding Helm charts every time your data changes.
</p>

<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); gap: 2rem; margin: 1.5rem 0;">
<div>

### ❌ Traditional Approaches Fall Short

<div style="padding: 1rem; background: var(--vp-c-bg-soft); border-radius: 8px; margin-bottom: 0.75rem; margin-top: 1rem;">
<strong style="display: block; margin-bottom: 0.5rem;">Helm Charts</strong>
<p style="margin: 0; font-size: 0.9rem; color: var(--vp-c-text-2);">
Static values files. Adding 10 new tenants? Manually create 10 new releases and track them separately.
</p>
</div>

<div style="padding: 1rem; background: var(--vp-c-bg-soft); border-radius: 8px; margin-bottom: 0.75rem;">
<strong style="display: block; margin-bottom: 0.5rem;">GitOps Only</strong>
<p style="margin: 0; font-size: 0.9rem; color: var(--vp-c-text-2);">
New customer signs up? Commit YAML, wait for CI/CD, manually sync. Not dynamic enough for real-time provisioning.
</p>
</div>

<div style="padding: 1rem; background: var(--vp-c-bg-soft); border-radius: 8px;">
<strong style="display: block; margin-bottom: 0.5rem;">Custom Scripts</strong>
<p style="margin: 0; font-size: 0.9rem; color: var(--vp-c-text-2);">
kubectl apply in bash loops. Works until you need drift detection, conflict handling, or dependency ordering.
</p>
</div>

</div>
<div>

### ✅ Tenant Operator Solves This

<div style="padding: 1rem; background: var(--vp-c-bg-soft); border-radius: 8px; margin-bottom: 0.75rem; margin-top: 1rem; border-left: 3px solid var(--vp-c-brand);">
<strong style="display: block; margin-bottom: 0.5rem;">Database-Driven Automation</strong>
<p style="margin: 0; font-size: 0.9rem; color: var(--vp-c-text-2);">
Your existing database is the source of truth. No YAML commits, no manual kubectl commands.
</p>
</div>

<div style="padding: 1rem; background: var(--vp-c-bg-soft); border-radius: 8px; margin-bottom: 0.75rem; border-left: 3px solid var(--vp-c-brand);">
<strong style="display: block; margin-bottom: 0.5rem;">Real-Time Synchronization</strong>
<p style="margin: 0; font-size: 0.9rem; color: var(--vp-c-text-2);">
Add a tenant → resources created in 30 seconds. Deactivate a tenant → everything cleaned up automatically.
</p>
</div>

<div style="padding: 1rem; background: var(--vp-c-bg-soft); border-radius: 8px; border-left: 3px solid var(--vp-c-brand);">
<strong style="display: block; margin-bottom: 0.5rem;">Production-Grade Control</strong>
<p style="margin: 0; font-size: 0.9rem; color: var(--vp-c-text-2);">
Built-in policies, drift detection, conflict resolution, dependency management, and comprehensive observability.
</p>
</div>

</div>
</div>

<div style="margin: 2.5rem 0; padding: 2rem; background: var(--vp-c-bg-soft); border-radius: 12px; border: 1px solid var(--vp-c-divider);">

<h3 style="margin: 0">Perfect For</h3>

<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 1.5rem; margin-top: 1.5rem;">
<div style="text-align: center;">
<div style="font-size: 2.5rem; margin-bottom: 0.75rem;">🏢</div>
<strong style="display: block; margin-bottom: 0.5rem;">SaaS Platforms</strong>
<p style="margin: 0; font-size: 0.9rem; color: var(--vp-c-text-2);">
Each customer gets isolated infrastructure provisioned from your user database
</p>
</div>

<div style="text-align: center;">
<div style="font-size: 2.5rem; margin-bottom: 0.75rem;">🌍</div>
<strong style="display: block; margin-bottom: 0.5rem;">Multi-Environment Apps</strong>
<p style="margin: 0; font-size: 0.9rem; color: var(--vp-c-text-2);">
Spin up dev/staging/prod environments dynamically per team or feature branch
</p>
</div>

<div style="text-align: center;">
<div style="font-size: 2.5rem; margin-bottom: 0.75rem;">🔧</div>
<strong style="display: block; margin-bottom: 0.5rem;">Internal Platforms</strong>
<p style="margin: 0; font-size: 0.9rem; color: var(--vp-c-text-2);">
Self-service infrastructure for teams without manual ticket workflows
</p>
</div>
</div>

</div>

<div style="text-align: center; padding: 2.5rem 2rem; background: linear-gradient(135deg, rgba(66, 184, 131, 0.08) 0%, rgba(102, 126, 234, 0.08) 100%); border-radius: 16px; margin: 2.5rem 0; border: 1px solid rgba(102, 126, 234, 0.15); backdrop-filter: blur(10px); -webkit-backdrop-filter: blur(10px); box-shadow: 0 4px 24px rgba(102, 126, 234, 0.08);">
<div>
<div style="font-size: clamp(1.25rem, 3vw, 1.75rem); font-weight: 700; margin-bottom: 0.75rem; color: var(--vp-c-text-1);">
Stop Managing Tenants Manually
</div>
<p style="font-size: clamp(0.9rem, 2vw, 1.05rem); margin: 0 0 1.5rem; line-height: 1.6; color: var(--vp-c-text-2); max-width: 600px; margin-left: auto; margin-right: auto;">
Let your database drive your infrastructure. Focus on your product, not kubectl commands.
</p>
<a href="/quickstart" style="display: inline-block; padding: 0.75rem 2rem; background: linear-gradient(135deg, rgba(102, 126, 234, 0.12) 0%, rgba(118, 75, 162, 0.12) 100%); color: var(--vp-c-brand); border-radius: 10px; text-decoration: none; font-weight: 600; font-size: 1rem; transition: all 0.3s ease; border: 1.5px solid rgba(102, 126, 234, 0.25); box-shadow: 0 2px 12px rgba(102, 126, 234, 0.15);">
Get Started in 5 Minutes →
</a>
</div>
</div>

## How It Works

<HowItWorksDiagram />

::: tip 💡 Interactive
Click **TenantRegistry** and **TenantTemplate** to see the YAML, or click database rows to toggle tenants
:::


<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 1.5rem; margin: 3rem 0">
  <div style="text-align: center; padding: 2rem 1.5rem; background: var(--vp-c-bg-soft); border-radius: 12px">
    <div style="font-size: 2.5rem; margin-bottom: 0.75rem">1️⃣</div>
    <h3 style="margin: 0.5rem 0 0.75rem">Connect Your Data</h3>
    <p style="margin: 0; color: var(--vp-c-text-2); font-size: 0.95rem; line-height: 1.6">
      Point to your MySQL database where tenant information lives. The operator reads active tenants automatically.
    </p>
  </div>

  <div style="text-align: center; padding: 2rem 1.5rem; background: var(--vp-c-bg-soft); border-radius: 12px">
    <div style="font-size: 2.5rem; margin-bottom: 0.75rem">2️⃣</div>
    <h3 style="margin: 0.5rem 0 0.75rem">Define Your Template</h3>
    <p style="margin: 0; color: var(--vp-c-text-2); font-size: 0.95rem; line-height: 1.6">
      Write one template describing what each tenant needs: deployments, services, ingresses, and any custom resources.
    </p>
  </div>

  <div style="text-align: center; padding: 2rem 1.5rem; background: var(--vp-c-bg-soft); border-radius: 12px">
    <div style="font-size: 2.5rem; margin-bottom: 0.75rem">3️⃣</div>
    <h3 style="margin: 0.5rem 0 0.75rem">Deploy Automatically</h3>
    <p style="margin: 0; color: var(--vp-c-text-2); font-size: 0.95rem; line-height: 1.6">
      Every active tenant gets isolated infrastructure. Resources are created, updated, and cleaned up automatically as your data changes.
    </p>
  </div>
</div>

<div style="text-align: center; padding: 3rem 2rem; background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%); border-radius: 16px; margin: 3rem 0; border: 1px solid rgba(102, 126, 234, 0.2); backdrop-filter: blur(10px); -webkit-backdrop-filter: blur(10px); box-shadow: 0 4px 24px rgba(102, 126, 234, 0.1);">
  <div>
    <div style="font-size: clamp(1.5rem, 4vw, 2.25rem); font-weight: 700; margin-bottom: 0.75rem; line-height: 1.2; color: var(--vp-c-text-1);">
      1 Database Row = 1 Complete Stack
    </div>
    <p style="font-size: clamp(0.95rem, 2vw, 1.1rem); margin: 0; line-height: 1.6; color: var(--vp-c-text-2); max-width: 700px; margin-left: auto; margin-right: auto;">
      Add a tenant to your database → Get Deployment + Service + Ingress + DNS + whatever you need
    </p>
  </div>
</div>

::: tip Start in 5 minutes
Follow the [Quick Start Guide](/quickstart) to see this in action with a working MySQL database and sample templates.
:::

## Kubernetes Compatibility

<div style="display: flex; align-items: center; gap: 1rem; padding: 1.25rem; background: var(--vp-c-bg-soft); border-radius: 8px; margin: 2.5rem 0">
  <div style="font-size: 2rem">✅</div>
  <div>
    <strong style="font-size: 1.05rem">Validated on Kubernetes v1.28 – v1.33</strong>
    <p style="margin: 0.5rem 0 0; font-size: 0.9rem; color: var(--vp-c-text-2)">
      Production-tested across multiple versions • See <a href="/installation#kubernetes-compatibility">compatibility details</a>
    </p>
  </div>
</div>

## Documentation

<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 1.5rem; margin: 2.5rem 0">
  <div style="padding: 1.5rem; background: var(--vp-c-bg-soft); border-radius: 8px; border-left: 4px solid #42b883">
    <h3 style="margin: 0 0 1rem">🚀 Getting Started</h3>
    <ul style="list-style: none; padding: 0; margin: 0">
      <li style="margin: 0.75rem 0">
        <a href="/installation"><strong>Installation</strong></a><br/>
        <span style="font-size: 0.9rem; color: var(--vp-c-text-2)">Deploy the operator to your cluster</span>
      </li>
      <li style="margin: 0.75rem 0">
        <a href="/quickstart"><strong>Quick Start</strong></a><br/>
        <span style="font-size: 0.9rem; color: var(--vp-c-text-2)">Complete tutorial in 5 minutes</span>
      </li>
      <li style="margin: 0.75rem 0">
        <a href="/local-development-minikube"><strong>Local Development</strong></a><br/>
        <span style="font-size: 0.9rem; color: var(--vp-c-text-2)">Set up with Minikube</span>
      </li>
    </ul>
  </div>

  <div style="padding: 1.5rem; background: var(--vp-c-bg-soft); border-radius: 8px; border-left: 4px solid #3b82f6">
    <h3 style="margin: 0 0 1rem">📚 Core Concepts</h3>
    <ul style="list-style: none; padding: 0; margin: 0">
      <li style="margin: 0.75rem 0">
        <a href="/architecture"><strong>Architecture</strong></a><br/>
        <span style="font-size: 0.9rem; color: var(--vp-c-text-2)">System design and reconciliation flow</span>
      </li>
      <li style="margin: 0.75rem 0">
        <a href="/api"><strong>API Reference</strong></a><br/>
        <span style="font-size: 0.9rem; color: var(--vp-c-text-2)">Complete CRD specification</span>
      </li>
      <li style="margin: 0.75rem 0">
        <a href="/templates"><strong>Templates</strong></a><br/>
        <span style="font-size: 0.9rem; color: var(--vp-c-text-2)">Go templates with 200+ functions</span>
      </li>
      <li style="margin: 0.75rem 0">
        <a href="/policies"><strong>Policies</strong></a><br/>
        <span style="font-size: 0.9rem; color: var(--vp-c-text-2)">Lifecycle and conflict management</span>
      </li>
      <li style="margin: 0.75rem 0">
        <a href="/datasource"><strong>Datasources</strong></a><br/>
        <span style="font-size: 0.9rem; color: var(--vp-c-text-2)">External data integration (MySQL)</span>
      </li>
    </ul>
  </div>

  <div style="padding: 1.5rem; background: var(--vp-c-bg-soft); border-radius: 8px; border-left: 4px solid #f59e0b">
    <h3 style="margin: 0 0 1rem">⚙️ Operations</h3>
    <ul style="list-style: none; padding: 0; margin: 0">
      <li style="margin: 0.75rem 0">
        <a href="/monitoring"><strong>Monitoring</strong></a><br/>
        <span style="font-size: 0.9rem; color: var(--vp-c-text-2)">Prometheus metrics, alerts, and Grafana</span>
      </li>
      <li style="margin: 0.75rem 0">
        <a href="/performance"><strong>Performance</strong></a><br/>
        <span style="font-size: 0.9rem; color: var(--vp-c-text-2)">Tuning and scalability</span>
      </li>
      <li style="margin: 0.75rem 0">
        <a href="/security"><strong>Security</strong></a><br/>
        <span style="font-size: 0.9rem; color: var(--vp-c-text-2)">RBAC, credentials, and best practices</span>
      </li>
      <li style="margin: 0.75rem 0">
        <a href="/troubleshooting"><strong>Troubleshooting</strong></a><br/>
        <span style="font-size: 0.9rem; color: var(--vp-c-text-2)">Common issues and solutions</span>
      </li>
    </ul>
  </div>

  <div style="padding: 1.5rem; background: var(--vp-c-bg-soft); border-radius: 8px; border-left: 4px solid #8b5cf6">
    <h3 style="margin: 0 0 1rem">🔌 Integrations</h3>
    <ul style="list-style: none; padding: 0; margin: 0">
      <li style="margin: 0.75rem 0">
        <a href="/integration-crossplane"><strong>Crossplane</strong></a><br/>
        <span style="font-size: 0.9rem; color: var(--vp-c-text-2)">K8s-native cloud provisioning</span>
      </li>
      <li style="margin: 0.75rem 0">
        <a href="/integration-external-dns"><strong>External DNS</strong></a><br/>
        <span style="font-size: 0.9rem; color: var(--vp-c-text-2)">Automatic DNS per tenant</span>
      </li>
      <li style="margin: 0.75rem 0">
        <a href="/integration-terraform-operator"><strong>Terraform Operator</strong></a><br/>
        <span style="font-size: 0.9rem; color: var(--vp-c-text-2)">Cloud resource provisioning</span>
      </li>
      <li style="margin: 0.75rem 0">
        <a href="/integration-argocd"><strong>Argo CD</strong></a><br/>
        <span style="font-size: 0.9rem; color: var(--vp-c-text-2)">GitOps delivery pipeline</span>
      </li>
    </ul>
  </div>
</div>

## Resources & Community

<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 1.5rem; margin: 2.5rem 0 0">
  <div style="padding: 1.5rem; background: var(--vp-c-bg-soft); border-radius: 8px">
    <h3 style="margin: 0 0 0.75rem">📦 GitHub Repository</h3>
    <p style="margin: 0 0 0.5rem">
      <a href="https://github.com/kubernetes-tenants/tenant-operator" target="_blank" rel="noopener noreferrer">kubernetes-tenants/tenant-operator</a>
    </p>
    <p style="margin: 0; font-size: 0.9rem; color: var(--vp-c-text-2)">
      Source code, releases, and project roadmap
    </p>
  </div>

  <div style="padding: 1.5rem; background: var(--vp-c-bg-soft); border-radius: 8px">
    <h3 style="margin: 0 0 0.75rem">🐛 Issue Tracker</h3>
    <p style="margin: 0 0 0.5rem">
      <a href="https://github.com/kubernetes-tenants/tenant-operator/issues" target="_blank" rel="noopener noreferrer">Report Issues</a>
    </p>
    <p style="margin: 0; font-size: 0.9rem; color: var(--vp-c-text-2)">
      Bug reports, feature requests, and discussions
    </p>
  </div>

  <div style="padding: 1.5rem; background: var(--vp-c-bg-soft); border-radius: 8px">
    <h3 style="margin: 0 0 0.75rem">📖 Documentation</h3>
    <p style="margin: 0 0 0.5rem">
      <a href="/installation">Get Started →</a>
    </p>
    <p style="margin: 0; font-size: 0.9rem; color: var(--vp-c-text-2)">
      Comprehensive guides and API reference
    </p>
  </div>
</div>
