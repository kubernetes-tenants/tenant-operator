import{_ as s,c as a,o as e,a2 as l}from"./chunks/framework.9Uv4PgnO.js";const b=JSON.parse('{"title":"Prometheus Query Examples","description":"","frontmatter":{},"headers":[],"relativePath":"prometheus-queries.md","filePath":"prometheus-queries.md","lastUpdated":1761887194000}'),p={name:"prometheus-queries.md"};function r(i,n,t,c,o,u){return e(),a("div",null,[...n[0]||(n[0]=[l(`<h1 id="prometheus-query-examples" tabindex="-1">Prometheus Query Examples <a class="header-anchor" href="#prometheus-query-examples" aria-label="Permalink to &quot;Prometheus Query Examples&quot;">​</a></h1><p>This document provides ready-to-use PromQL queries for monitoring Tenant Operator.</p><nav class="table-of-contents"><ul><li><a href="#tenant-health">Tenant Health</a><ul><li><a href="#check-ready-tenants">Check Ready Tenants</a></li><li><a href="#check-not-ready-tenants">Check Not Ready Tenants</a></li><li><a href="#check-degraded-tenants">Check Degraded Tenants</a></li><li><a href="#resource-health-by-tenant">Resource Health by Tenant</a></li></ul></li><li><a href="#conflict-monitoring">Conflict Monitoring</a><ul><li><a href="#current-conflicts">Current Conflicts</a></li><li><a href="#conflict-rate">Conflict Rate</a></li><li><a href="#historical-conflicts">Historical Conflicts</a></li><li><a href="#conflict-policy-analysis">Conflict Policy Analysis</a></li></ul></li><li><a href="#failure-detection">Failure Detection</a><ul><li><a href="#failed-resources">Failed Resources</a></li><li><a href="#failure-trends">Failure Trends</a></li><li><a href="#critical-failures">Critical Failures</a></li></ul></li><li><a href="#performance-monitoring">Performance Monitoring</a><ul><li><a href="#reconciliation-duration">Reconciliation Duration</a></li><li><a href="#reconciliation-rate">Reconciliation Rate</a></li><li><a href="#apply-performance">Apply Performance</a></li></ul></li><li><a href="#registry-health">Registry Health</a><ul><li><a href="#registry-status">Registry Status</a></li><li><a href="#registry-capacity">Registry Capacity</a></li><li><a href="#registry-trends">Registry Trends</a></li></ul></li><li><a href="#capacity-planning">Capacity Planning</a><ul><li><a href="#resource-counts">Resource Counts</a></li><li><a href="#growth-trends">Growth Trends</a></li><li><a href="#load-distribution">Load Distribution</a></li></ul></li><li><a href="#combined-queries">Combined Queries</a><ul><li><a href="#overall-health-dashboard">Overall Health Dashboard</a></li><li><a href="#problem-detection">Problem Detection</a></li><li><a href="#performance-summary">Performance Summary</a></li></ul></li><li><a href="#alert-conditions">Alert Conditions</a><ul><li><a href="#critical-conditions">Critical Conditions</a></li><li><a href="#warning-conditions">Warning Conditions</a></li></ul></li><li><a href="#tips-for-using-these-queries">Tips for Using These Queries</a><ul><li><a href="#example-filter-by-namespace-and-time">Example: Filter by Namespace and Time</a></li></ul></li><li><a href="#see-also">See Also</a></li></ul></nav><h2 id="tenant-health" tabindex="-1">Tenant Health <a class="header-anchor" href="#tenant-health" aria-label="Permalink to &quot;Tenant Health&quot;">​</a></h2><h3 id="check-ready-tenants" tabindex="-1">Check Ready Tenants <a class="header-anchor" href="#check-ready-tenants" aria-label="Permalink to &quot;Check Ready Tenants&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># All ready tenants</span></span>
<span class="line"><span>tenant_condition_status{type=&quot;Ready&quot;} == 1</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Count ready tenants</span></span>
<span class="line"><span>count(tenant_condition_status{type=&quot;Ready&quot;} == 1)</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Percentage of ready tenants</span></span>
<span class="line"><span>count(tenant_condition_status{type=&quot;Ready&quot;} == 1) / count(tenant_condition_status{type=&quot;Ready&quot;}) * 100</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br></div></div><h3 id="check-not-ready-tenants" tabindex="-1">Check Not Ready Tenants <a class="header-anchor" href="#check-not-ready-tenants" aria-label="Permalink to &quot;Check Not Ready Tenants&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># All not ready tenants</span></span>
<span class="line"><span>tenant_condition_status{type=&quot;Ready&quot;} != 1</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Count not ready tenants</span></span>
<span class="line"><span>count(tenant_condition_status{type=&quot;Ready&quot;} != 1)</span></span>
<span class="line"><span></span></span>
<span class="line"><span># List not ready tenants with details</span></span>
<span class="line"><span>tenant_condition_status{type=&quot;Ready&quot;} != 1</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br></div></div><h3 id="check-degraded-tenants" tabindex="-1">Check Degraded Tenants <a class="header-anchor" href="#check-degraded-tenants" aria-label="Permalink to &quot;Check Degraded Tenants&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># All degraded tenants</span></span>
<span class="line"><span>tenant_degraded_status == 1</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Count degraded tenants</span></span>
<span class="line"><span>count(tenant_degraded_status == 1)</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Degraded tenants by reason</span></span>
<span class="line"><span>sum(tenant_degraded_status) by (reason)</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Top 10 degraded tenants</span></span>
<span class="line"><span>topk(10, tenant_degraded_status)</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br></div></div><h3 id="resource-health-by-tenant" tabindex="-1">Resource Health by Tenant <a class="header-anchor" href="#resource-health-by-tenant" aria-label="Permalink to &quot;Resource Health by Tenant&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># Ready resources per tenant</span></span>
<span class="line"><span>tenant_resources_ready</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Failed resources per tenant</span></span>
<span class="line"><span>tenant_resources_failed</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Resource readiness percentage per tenant</span></span>
<span class="line"><span>(tenant_resources_ready / tenant_resources_desired) * 100</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Tenants with 100% resources ready</span></span>
<span class="line"><span>(tenant_resources_ready / tenant_resources_desired) == 1</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br></div></div><h2 id="conflict-monitoring" tabindex="-1">Conflict Monitoring <a class="header-anchor" href="#conflict-monitoring" aria-label="Permalink to &quot;Conflict Monitoring&quot;">​</a></h2><h3 id="current-conflicts" tabindex="-1">Current Conflicts <a class="header-anchor" href="#current-conflicts" aria-label="Permalink to &quot;Current Conflicts&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># Total resources currently in conflict</span></span>
<span class="line"><span>sum(tenant_resources_conflicted)</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Tenants with conflicts</span></span>
<span class="line"><span>tenant_resources_conflicted &gt; 0</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Top 10 tenants with most conflicts</span></span>
<span class="line"><span>topk(10, tenant_resources_conflicted)</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Conflict percentage per tenant</span></span>
<span class="line"><span>(tenant_resources_conflicted / tenant_resources_desired) * 100</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br></div></div><h3 id="conflict-rate" tabindex="-1">Conflict Rate <a class="header-anchor" href="#conflict-rate" aria-label="Permalink to &quot;Conflict Rate&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># Conflict rate (conflicts per second)</span></span>
<span class="line"><span>rate(tenant_conflicts_total[5m])</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Conflict rate per tenant</span></span>
<span class="line"><span>sum(rate(tenant_conflicts_total[5m])) by (tenant)</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Conflict rate by resource kind</span></span>
<span class="line"><span>sum(rate(tenant_conflicts_total[5m])) by (resource_kind)</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Conflict rate by policy</span></span>
<span class="line"><span>sum(rate(tenant_conflicts_total[5m])) by (conflict_policy)</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br></div></div><h3 id="historical-conflicts" tabindex="-1">Historical Conflicts <a class="header-anchor" href="#historical-conflicts" aria-label="Permalink to &quot;Historical Conflicts&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># Total conflicts in last hour</span></span>
<span class="line"><span>increase(tenant_conflicts_total[1h])</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Total conflicts in last 24 hours</span></span>
<span class="line"><span>increase(tenant_conflicts_total[24h])</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Conflicts over time (5m windows)</span></span>
<span class="line"><span>sum(increase(tenant_conflicts_total[5m])) by (tenant)</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br></div></div><h3 id="conflict-policy-analysis" tabindex="-1">Conflict Policy Analysis <a class="header-anchor" href="#conflict-policy-analysis" aria-label="Permalink to &quot;Conflict Policy Analysis&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># Conflicts by policy type</span></span>
<span class="line"><span>sum(tenant_conflicts_total) by (conflict_policy)</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Force policy usage rate</span></span>
<span class="line"><span>rate(tenant_conflicts_total{conflict_policy=&quot;Force&quot;}[5m])</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Stuck policy conflicts</span></span>
<span class="line"><span>rate(tenant_conflicts_total{conflict_policy=&quot;Stuck&quot;}[5m])</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br></div></div><h2 id="failure-detection" tabindex="-1">Failure Detection <a class="header-anchor" href="#failure-detection" aria-label="Permalink to &quot;Failure Detection&quot;">​</a></h2><h3 id="failed-resources" tabindex="-1">Failed Resources <a class="header-anchor" href="#failed-resources" aria-label="Permalink to &quot;Failed Resources&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># Total failed resources</span></span>
<span class="line"><span>sum(tenant_resources_failed)</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Tenants with failed resources</span></span>
<span class="line"><span>tenant_resources_failed &gt; 0</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Top 10 tenants with most failures</span></span>
<span class="line"><span>topk(10, tenant_resources_failed)</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Failure rate per tenant</span></span>
<span class="line"><span>(tenant_resources_failed / tenant_resources_desired) * 100</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br></div></div><h3 id="failure-trends" tabindex="-1">Failure Trends <a class="header-anchor" href="#failure-trends" aria-label="Permalink to &quot;Failure Trends&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># Failed resources over time</span></span>
<span class="line"><span>tenant_resources_failed</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Increase in failures (last 1h)</span></span>
<span class="line"><span>increase(tenant_resources_failed[1h])</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Average failures per tenant</span></span>
<span class="line"><span>avg(tenant_resources_failed)</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br></div></div><h3 id="critical-failures" tabindex="-1">Critical Failures <a class="header-anchor" href="#critical-failures" aria-label="Permalink to &quot;Critical Failures&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># Tenants with &gt;50% resources failed</span></span>
<span class="line"><span>(tenant_resources_failed / tenant_resources_desired) &gt; 0.5</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Tenants with &gt;5 failed resources</span></span>
<span class="line"><span>tenant_resources_failed &gt; 5</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Tenants that are both degraded and have failures</span></span>
<span class="line"><span>tenant_degraded_status == 1 and tenant_resources_failed &gt; 0</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br></div></div><h2 id="performance-monitoring" tabindex="-1">Performance Monitoring <a class="header-anchor" href="#performance-monitoring" aria-label="Permalink to &quot;Performance Monitoring&quot;">​</a></h2><h3 id="reconciliation-duration" tabindex="-1">Reconciliation Duration <a class="header-anchor" href="#reconciliation-duration" aria-label="Permalink to &quot;Reconciliation Duration&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># P50 reconciliation duration</span></span>
<span class="line"><span>histogram_quantile(0.50, rate(tenant_reconcile_duration_seconds_bucket[5m]))</span></span>
<span class="line"><span></span></span>
<span class="line"><span># P95 reconciliation duration</span></span>
<span class="line"><span>histogram_quantile(0.95, rate(tenant_reconcile_duration_seconds_bucket[5m]))</span></span>
<span class="line"><span></span></span>
<span class="line"><span># P99 reconciliation duration</span></span>
<span class="line"><span>histogram_quantile(0.99, rate(tenant_reconcile_duration_seconds_bucket[5m]))</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Max reconciliation duration</span></span>
<span class="line"><span>histogram_quantile(1.0, rate(tenant_reconcile_duration_seconds_bucket[5m]))</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br></div></div><h3 id="reconciliation-rate" tabindex="-1">Reconciliation Rate <a class="header-anchor" href="#reconciliation-rate" aria-label="Permalink to &quot;Reconciliation Rate&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># Total reconciliation rate</span></span>
<span class="line"><span>rate(tenant_reconcile_duration_seconds_count[5m])</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Success rate</span></span>
<span class="line"><span>rate(tenant_reconcile_duration_seconds_count{result=&quot;success&quot;}[5m])</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Error rate</span></span>
<span class="line"><span>rate(tenant_reconcile_duration_seconds_count{result=&quot;error&quot;}[5m])</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Success percentage</span></span>
<span class="line"><span>(rate(tenant_reconcile_duration_seconds_count{result=&quot;success&quot;}[5m]) / rate(tenant_reconcile_duration_seconds_count[5m])) * 100</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br></div></div><h3 id="apply-performance" tabindex="-1">Apply Performance <a class="header-anchor" href="#apply-performance" aria-label="Permalink to &quot;Apply Performance&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># Apply rate by result</span></span>
<span class="line"><span>sum(rate(apply_attempts_total[5m])) by (result)</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Apply rate by resource kind</span></span>
<span class="line"><span>sum(rate(apply_attempts_total[5m])) by (kind)</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Apply success rate</span></span>
<span class="line"><span>rate(apply_attempts_total{result=&quot;success&quot;}[5m]) / rate(apply_attempts_total[5m])</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Failed applies by kind</span></span>
<span class="line"><span>sum(rate(apply_attempts_total{result=&quot;error&quot;}[5m])) by (kind)</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br></div></div><h2 id="registry-health" tabindex="-1">Registry Health <a class="header-anchor" href="#registry-health" aria-label="Permalink to &quot;Registry Health&quot;">​</a></h2><h3 id="registry-status" tabindex="-1">Registry Status <a class="header-anchor" href="#registry-status" aria-label="Permalink to &quot;Registry Status&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># Desired tenants per registry</span></span>
<span class="line"><span>registry_desired</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Ready tenants per registry</span></span>
<span class="line"><span>registry_ready</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Failed tenants per registry</span></span>
<span class="line"><span>registry_failed</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Registry health percentage</span></span>
<span class="line"><span>(registry_ready / registry_desired) * 100</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br></div></div><h3 id="registry-capacity" tabindex="-1">Registry Capacity <a class="header-anchor" href="#registry-capacity" aria-label="Permalink to &quot;Registry Capacity&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># Total desired tenants across all registries</span></span>
<span class="line"><span>sum(registry_desired)</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Total ready tenants across all registries</span></span>
<span class="line"><span>sum(registry_ready)</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Total failed tenants across all registries</span></span>
<span class="line"><span>sum(registry_failed)</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Overall health percentage</span></span>
<span class="line"><span>(sum(registry_ready) / sum(registry_desired)) * 100</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br></div></div><h3 id="registry-trends" tabindex="-1">Registry Trends <a class="header-anchor" href="#registry-trends" aria-label="Permalink to &quot;Registry Trends&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># Registry health over time</span></span>
<span class="line"><span>(registry_ready / registry_desired) * 100</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Registries with &gt;90% health</span></span>
<span class="line"><span>(registry_ready / registry_desired) &gt; 0.9</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Unhealthy registries (&lt;80% ready)</span></span>
<span class="line"><span>(registry_ready / registry_desired) &lt; 0.8</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br></div></div><h2 id="capacity-planning" tabindex="-1">Capacity Planning <a class="header-anchor" href="#capacity-planning" aria-label="Permalink to &quot;Capacity Planning&quot;">​</a></h2><h3 id="resource-counts" tabindex="-1">Resource Counts <a class="header-anchor" href="#resource-counts" aria-label="Permalink to &quot;Resource Counts&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># Total desired resources across all tenants</span></span>
<span class="line"><span>sum(tenant_resources_desired)</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Total ready resources</span></span>
<span class="line"><span>sum(tenant_resources_ready)</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Total failed resources</span></span>
<span class="line"><span>sum(tenant_resources_failed)</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Total conflicted resources</span></span>
<span class="line"><span>sum(tenant_resources_conflicted)</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br></div></div><h3 id="growth-trends" tabindex="-1">Growth Trends <a class="header-anchor" href="#growth-trends" aria-label="Permalink to &quot;Growth Trends&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># Desired tenant growth rate</span></span>
<span class="line"><span>rate(registry_desired[24h])</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Resource growth per tenant</span></span>
<span class="line"><span>rate(tenant_resources_desired[24h])</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Average resources per tenant</span></span>
<span class="line"><span>avg(tenant_resources_desired)</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br></div></div><h3 id="load-distribution" tabindex="-1">Load Distribution <a class="header-anchor" href="#load-distribution" aria-label="Permalink to &quot;Load Distribution&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># Top 10 tenants by resource count</span></span>
<span class="line"><span>topk(10, tenant_resources_desired)</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Bottom 10 tenants by resource count</span></span>
<span class="line"><span>bottomk(10, tenant_resources_desired)</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Tenants with &gt;100 resources</span></span>
<span class="line"><span>tenant_resources_desired &gt; 100</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Distribution of resources per tenant</span></span>
<span class="line"><span>histogram_quantile(0.50, tenant_resources_desired)</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br></div></div><h2 id="combined-queries" tabindex="-1">Combined Queries <a class="header-anchor" href="#combined-queries" aria-label="Permalink to &quot;Combined Queries&quot;">​</a></h2><h3 id="overall-health-dashboard" tabindex="-1">Overall Health Dashboard <a class="header-anchor" href="#overall-health-dashboard" aria-label="Permalink to &quot;Overall Health Dashboard&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># Total tenants</span></span>
<span class="line"><span>count(tenant_condition_status{type=&quot;Ready&quot;})</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Ready percentage</span></span>
<span class="line"><span>count(tenant_condition_status{type=&quot;Ready&quot;} == 1) / count(tenant_condition_status{type=&quot;Ready&quot;}) * 100</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Total resources</span></span>
<span class="line"><span>sum(tenant_resources_desired)</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Ready resources percentage</span></span>
<span class="line"><span>sum(tenant_resources_ready) / sum(tenant_resources_desired) * 100</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Active conflicts</span></span>
<span class="line"><span>sum(tenant_resources_conflicted)</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Total failures</span></span>
<span class="line"><span>sum(tenant_resources_failed)</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br><span class="line-number">12</span><br><span class="line-number">13</span><br><span class="line-number">14</span><br><span class="line-number">15</span><br><span class="line-number">16</span><br><span class="line-number">17</span><br></div></div><h3 id="problem-detection" tabindex="-1">Problem Detection <a class="header-anchor" href="#problem-detection" aria-label="Permalink to &quot;Problem Detection&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># Tenants with issues (not ready OR degraded OR conflicts OR failures)</span></span>
<span class="line"><span>(tenant_condition_status{type=&quot;Ready&quot;} != 1)</span></span>
<span class="line"><span>or (tenant_degraded_status == 1)</span></span>
<span class="line"><span>or (tenant_resources_conflicted &gt; 0)</span></span>
<span class="line"><span>or (tenant_resources_failed &gt; 0)</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Count problematic tenants</span></span>
<span class="line"><span>count(</span></span>
<span class="line"><span>  (tenant_condition_status{type=&quot;Ready&quot;} != 1)</span></span>
<span class="line"><span>  or (tenant_degraded_status == 1)</span></span>
<span class="line"><span>  or (tenant_resources_conflicted &gt; 0)</span></span>
<span class="line"><span>  or (tenant_resources_failed &gt; 0)</span></span>
<span class="line"><span>)</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br><span class="line-number">12</span><br><span class="line-number">13</span><br></div></div><h3 id="performance-summary" tabindex="-1">Performance Summary <a class="header-anchor" href="#performance-summary" aria-label="Permalink to &quot;Performance Summary&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># P95 latency, error rate, and throughput</span></span>
<span class="line"><span>{</span></span>
<span class="line"><span>  p95_latency: histogram_quantile(0.95, rate(tenant_reconcile_duration_seconds_bucket[5m])),</span></span>
<span class="line"><span>  error_rate: rate(tenant_reconcile_duration_seconds_count{result=&quot;error&quot;}[5m]),</span></span>
<span class="line"><span>  throughput: rate(tenant_reconcile_duration_seconds_count[5m])</span></span>
<span class="line"><span>}</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br></div></div><h2 id="alert-conditions" tabindex="-1">Alert Conditions <a class="header-anchor" href="#alert-conditions" aria-label="Permalink to &quot;Alert Conditions&quot;">​</a></h2><p>These queries are used in the alert rules (<code>config/prometheus/alerts.yaml</code>):</p><h3 id="critical-conditions" tabindex="-1">Critical Conditions <a class="header-anchor" href="#critical-conditions" aria-label="Permalink to &quot;Critical Conditions&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># Tenant has failed resources</span></span>
<span class="line"><span>tenant_resources_failed &gt; 0</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Tenant is degraded</span></span>
<span class="line"><span>tenant_degraded_status &gt; 0</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Tenant not ready</span></span>
<span class="line"><span>tenant_condition_status{type=&quot;Ready&quot;} != 1</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Registry has many failures</span></span>
<span class="line"><span>registry_failed &gt; 5 or (registry_failed / registry_desired &gt; 0.5 and registry_desired &gt; 0)</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br></div></div><h3 id="warning-conditions" tabindex="-1">Warning Conditions <a class="header-anchor" href="#warning-conditions" aria-label="Permalink to &quot;Warning Conditions&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># Resources in conflict</span></span>
<span class="line"><span>tenant_resources_conflicted &gt; 0</span></span>
<span class="line"><span></span></span>
<span class="line"><span># High conflict rate</span></span>
<span class="line"><span>rate(tenant_conflicts_total[5m]) &gt; 0.1</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Resources mismatch</span></span>
<span class="line"><span>tenant_resources_ready != tenant_resources_desired and tenant_resources_desired &gt; 0</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Slow reconciliation</span></span>
<span class="line"><span>histogram_quantile(0.95, rate(tenant_reconcile_duration_seconds_bucket[5m])) &gt; 30</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br></div></div><h2 id="tips-for-using-these-queries" tabindex="-1">Tips for Using These Queries <a class="header-anchor" href="#tips-for-using-these-queries" aria-label="Permalink to &quot;Tips for Using These Queries&quot;">​</a></h2><ol><li><strong>Adjust Time Windows</strong>: Change <code>[5m]</code>, <code>[1h]</code>, <code>[24h]</code> based on your needs</li><li><strong>Filter by Namespace</strong>: Add <code>{namespace=&quot;default&quot;}</code> to filter</li><li><strong>Filter by Tenant</strong>: Add <code>{tenant=&quot;my-tenant&quot;}</code> to focus on specific tenants</li><li><strong>Combine Queries</strong>: Use <code>and</code>, <code>or</code>, <code>unless</code> for complex conditions</li><li><strong>Aggregation</strong>: Use <code>sum</code>, <code>avg</code>, <code>max</code>, <code>min</code> for aggregations</li><li><strong>Top/Bottom N</strong>: Use <code>topk(N, ...)</code> or <code>bottomk(N, ...)</code></li></ol><h3 id="example-filter-by-namespace-and-time" tabindex="-1">Example: Filter by Namespace and Time <a class="header-anchor" href="#example-filter-by-namespace-and-time" aria-label="Permalink to &quot;Example: Filter by Namespace and Time&quot;">​</a></h3><div class="language-promql line-numbers-mode"><button title="Copy Code" class="copy"></button><span class="lang">promql</span><pre class="shiki github-dark vp-code" tabindex="0"><code><span class="line"><span># Failed resources in default namespace, last 1 hour</span></span>
<span class="line"><span>tenant_resources_failed{namespace=&quot;default&quot;}[1h]</span></span>
<span class="line"><span></span></span>
<span class="line"><span># Conflicts for specific tenant in last 5 minutes</span></span>
<span class="line"><span>rate(tenant_conflicts_total{tenant=&quot;acme-prod-template&quot;, namespace=&quot;default&quot;}[5m])</span></span></code></pre><div class="line-numbers-wrapper" aria-hidden="true"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br></div></div><h2 id="see-also" tabindex="-1">See Also <a class="header-anchor" href="#see-also" aria-label="Permalink to &quot;See Also&quot;">​</a></h2><ul><li><a href="./monitoring.html">Monitoring Guide</a> - Complete monitoring documentation</li><li><a href="../config/prometheus/alerts.yaml">Alert Rules</a> - Prometheus alert rules</li><li><a href="./troubleshooting.html">Troubleshooting</a> - Common issues and solutions</li></ul>`,68)])])}const m=s(p,[["render",r]]);export{b as __pageData,m as default};
