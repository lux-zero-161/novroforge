(() => {
  const LIMIT = 3;
  const lang = document.documentElement.lang || undefined;
  const relative = new Intl.RelativeTimeFormat(lang, { numeric: "auto" });
  const day = new Intl.DateTimeFormat(lang, { month: "short", day: "numeric" });

  const ago = time => {
    const seconds = (time - Date.now()) / 1000;
    const steps = [
      ["year", 31536000],
      ["month", 2592000],
      ["week", 604800],
      ["day", 86400],
      ["hour", 3600],
      ["minute", 60],
    ];
    for (const [unit, size] of steps)
      if (Math.abs(seconds) >= size) return relative.format(Math.trunc(seconds / size), unit);
    return relative.format(0, "minute");
  };

  // Only paths on this site; the search answers with absolute links.
  const path = url => {
    try {
      const parsed = new URL(url, location.origin);
      return parsed.origin === location.origin ? parsed.pathname + parsed.search : "#";
    } catch {
      return "#";
    }
  };

  const setup = rail => {
    const icons = rail.querySelector("[data-theme-work-icons]")?.content;
    const icon = name => {
      const source = icons?.querySelector(`[data-icon="${name}"] > svg`);
      return source ? source.cloneNode(true) : document.createElement("span");
    };

    const row = ({ kind, tone, title, link, meta, progress }) => {
      const item = document.createElement("li");
      item.className = "theme-work-row";
      const glyph = icon(kind);
      glyph.classList?.add("theme-work-icon", `theme-work-icon-${tone ?? kind}`);
      item.append(glyph);
      const body = document.createElement("div");
      body.className = "theme-work-body";
      const anchor = document.createElement("a");
      anchor.className = "theme-work-title";
      anchor.href = link;
      anchor.title = title;
      // Titles use Markdown's inline code, as the issue lists render it: `/v0` becomes code.
      title.split("`").forEach((part, index, parts) => {
        if (index % 2 && index < parts.length - 1) {
          const code = document.createElement("code");
          code.className = "inline-code-block";
          code.textContent = part;
          anchor.append(code);
        } else if (part) anchor.append(index % 2 ? `\`${part}` : part);
      });
      const detail = document.createElement("p");
      detail.className = "theme-work-meta";
      detail.textContent = meta;
      body.append(anchor, detail);
      if (progress !== undefined) {
        const bar = document.createElement("progress");
        bar.className = "theme-progress theme-work-progress";
        bar.max = 100;
        bar.value = progress;
        body.append(bar);
      }
      item.append(body);
      return item;
    };

    const filled = new Set();
    const fill = (panel, rows, total) => {
      if (!rows.length) return;
      panel.querySelector("[data-theme-work-rows]").replaceChildren(...rows.slice(0, LIMIT).map(row));
      panel.querySelector("[data-theme-work-count]").textContent = String(total);
      filled.add(panel);
    };

    // Every box appears at once, in its place, when all have answered; an empty rail gives its column to the feed.
    // The outcome is remembered, so the next visit starts with the rail open or closed.
    const reveal = () => {
      rail.removeAttribute("aria-busy");
      for (const panel of filled) {
        panel.hidden = false;
        panel.classList?.add("theme-fade-in");
      }
      rail.hidden = !filled.size;
      try {
        localStorage.setItem(rail.dataset.key, filled.size ? "full" : "empty");
      } catch {}
    };

    const search = async panel => {
      const params = new URLSearchParams(panel.dataset.query);
      params.set("state", "open");
      params.set("limit", String(LIMIT));
      params.set("sort", "recentupdate");
      if (rail.dataset.owner) params.set("owner", rail.dataset.owner);
      if (rail.dataset.team) params.set("team", rail.dataset.team);
      const response = await fetch(`${rail.dataset.search}?${params}`, { headers: { Accept: "application/json" } });
      if (!response.ok) return;
      const issues = await response.json();
      if (!Array.isArray(issues)) return;
      const rows = issues.map(issue => ({
        kind: issue.pull_request ? (issue.pull_request.draft ? "draft" : "pull") : "issue",
        tone: "open",
        title: String(issue.title ?? ""),
        link: path(issue.html_url),
        meta: `${issue.repository?.full_name ?? ""}#${issue.number} · ${ago(Date.parse(issue.updated_at))}`,
      }));
      fill(panel, rows, Number(response.headers.get("X-Total-Count")) || rows.length);
    };

    // The dashboard milestone list, soonest due first; only open milestones with a due date.
    const milestones = async panel => {
      const response = await fetch(panel.dataset.query, { headers: { Accept: "text/html" } });
      if (!response.ok) return;
      const page = new DOMParser().parseFromString(await response.text(), "text/html");
      const due = [...page.querySelectorAll("[data-theme-milestone][data-due]")];
      const rows = due.map(milestone => {
        const anchor = milestone.querySelector("h3 a");
        const repo = milestone.querySelector(".theme-milestone-repo")?.textContent.trim() ?? "";
        const date = new Date(`${milestone.dataset.due}T00:00:00`);
        return {
          kind: "milestone",
          tone: date.getTime() < Date.now() ? "overdue" : "milestone",
          title: anchor?.textContent.trim() ?? "",
          link: path(anchor?.getAttribute("href") ?? "#"),
          meta: [repo, day.format(date), `${Number(milestone.dataset.progress) || 0}%`].filter(Boolean).join(" · "),
          progress: Number(milestone.dataset.progress) || 0,
        };
      });
      fill(panel, rows, rows.length);
    };

    // Releases come from the feed's own rows, newest first, one per repository.
    const releases = panel => {
      const seen = new Set();
      const rows = [];
      for (const item of document.querySelectorAll("[data-theme-feed] .theme-feed-item[data-release]")) {
        if (seen.has(item.dataset.repo)) continue;
        seen.add(item.dataset.repo);
        rows.push({
          kind: "tag",
          title: item.dataset.release,
          link: path(item.dataset.link),
          meta: `${item.dataset.repo} · ${ago(Number(item.dataset.time) * 1000)}`,
        });
      }
      fill(panel, rows, rows.length);
    };

    const loads = [...rail.querySelectorAll("[data-theme-work-panel]")].map(panel => {
      const kind = panel.dataset.themeWorkPanel;
      const load = kind === "search" ? search : kind === "milestones" ? milestones : releases;
      return Promise.resolve()
        .then(() => load(panel))
        .catch(() => {});
    });
    Promise.all(loads).then(reveal);
  };

  for (const rail of document.querySelectorAll("[data-theme-work]")) setup(rail);
})();
