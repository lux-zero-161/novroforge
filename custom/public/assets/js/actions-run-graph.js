// GitHub's job graph for the Actions run summary. Forgejo's run data has no job dependencies, so they come from the
// workflow file itself; the run page builds the cards and uses this to place them in stages and draw the links.
window.themeActionsGraph = (() => {
  const unquote = value => value.trim().replace(/^(["'])(.*)\1$/, "$2");

  // Reads each job's `name` and `needs` from the workflow's `jobs` mapping. This is a small line reader, not a YAML
  // parser: a job whose `needs` it cannot read simply has no dependencies.
  const parseNeeds = text => {
    const jobs = new Map();
    let inJobs = false;
    let jobIndent = -1;
    let propertyIndent = -1;
    let job = null;
    let readingNeeds = false;
    for (const raw of text.split(/\r?\n/)) {
      const line = raw.replace(/\s+#.*$/, "");
      if (!line.trim() || line.trim().startsWith("#")) continue;
      const indent = line.length - line.trimStart().length;
      const content = line.trim();
      if (indent === 0) {
        inJobs = /^jobs:\s*$/.test(content);
        job = null;
        jobIndent = -1;
        continue;
      }
      if (!inJobs) continue;
      if (jobIndent < 0) jobIndent = indent;
      if (indent <= jobIndent) {
        const id = content.match(/^([^:]+):\s*$/)?.[1];
        job = id ? { name: unquote(id), needs: [] } : null;
        if (job) jobs.set(unquote(id), job);
        propertyIndent = -1;
        readingNeeds = false;
        continue;
      }
      if (!job) continue;
      if (propertyIndent < 0) propertyIndent = indent;
      if (indent === propertyIndent) {
        readingNeeds = false;
        const [, key, value = ""] = content.match(/^([\w-]+):\s*(.*)$/) ?? [];
        if (key === "name" && value) job.name = unquote(value);
        if (key !== "needs") continue;
        if (!value) readingNeeds = true;
        else if (value.startsWith("["))
          job.needs = value
            .replace(/^\[|\]$/g, "")
            .split(",")
            .map(unquote)
            .filter(Boolean);
        else job.needs = [unquote(value)];
      } else if (readingNeeds) {
        const item = content.match(/^-\s*(.+)$/)?.[1];
        if (item) job.needs.push(unquote(item));
      }
    }
    return jobs;
  };

  // Places every run job one stage after the deepest job it needs. Within a stage jobs are sorted by name, as on
  // GitHub; jobs with no links at all share one card at the end of their stage.
  const layout = (runJobs, workflow) => {
    const ids = new Map();
    for (const [id, job] of workflow) {
      ids.set(id, id);
      if (!job.name.includes("${{")) ids.set(job.name, id);
    }
    // Matrix jobs carry their values in parentheses after the job's name.
    const resolve = name => ids.get(name) ?? ids.get(name.replace(/\s*\(.*\)$/, ""));
    const nodes = runJobs.map(job => ({ ...job, id: resolve(job.name) }));
    const byId = new Map();
    for (const node of nodes) if (node.id) byId.set(node.id, [...(byId.get(node.id) ?? []), node]);
    const edges = [];
    for (const node of nodes)
      for (const need of node.id ? workflow.get(node.id).needs : [])
        for (const from of byId.get(need) ?? []) if (from !== node) edges.push([from, node]);
    const stage = new Map();
    const depth = (node, path = new Set()) => {
      if (stage.has(node)) return stage.get(node);
      if (path.has(node)) return 0;
      path.add(node);
      const parents = edges.filter(([, to]) => to === node).map(([from]) => depth(from, path));
      const value = parents.length ? Math.max(...parents) + 1 : 0;
      stage.set(node, value);
      return value;
    };
    for (const node of nodes) depth(node);
    const linked = new Set(edges.flat());
    const stages = [];
    for (const node of nodes) (stages[stage.get(node)] ??= []).push(node);
    return {
      edges,
      stages: stages.filter(Boolean).map(list => {
        const sorted = [...list].sort((a, b) => a.name.localeCompare(b.name));
        const cards = sorted.filter(node => linked.has(node)).map(node => [node]);
        const loose = sorted.filter(node => !linked.has(node));
        if (loose.length) cards.push(loose);
        return cards;
      }),
    };
  };

  // GitHub's link: straight across, or along the gap's middle with 12px rounded bends when the rows differ.
  const linkPath = (x1, y1, x2, y2) => {
    const dy = y2 - y1;
    if (Math.abs(dy) < 1) return `M ${x1} ${y1} H ${x2}`;
    const mid = x1 + 28;
    const r = Math.min(12, Math.abs(dy) / 4) * Math.sign(dy);
    const ar = Math.abs(r);
    return [
      `M ${x1} ${y1}`,
      `H ${mid - 2 * ar}`,
      `C ${mid - ar} ${y1} ${mid} ${y1 + r} ${mid} ${y1 + 2 * r}`,
      `V ${y2 - 2 * r}`,
      `C ${mid} ${y2 - r} ${mid + ar} ${y2} ${mid + 2 * ar} ${y2}`,
      `H ${x2}`,
    ].join(" ");
  };

  return { parseNeeds, layout, linkPath };
})();
