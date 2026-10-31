(() => {
  const pickers = new WeakMap();

  const iconFor = repo => {
    if (repo.private) return "private";
    if (repo.mirror) return "mirror";
    if (repo.fork) return "fork";
    if (repo.template) return "template";
    return "repo";
  };

  const setup = picker => {
    if (pickers.has(picker)) return pickers.get(picker);
    const list = picker.querySelector("[data-theme-repo-picker-list]");
    const query = picker.querySelector(".theme-repo-picker-search input");
    const message = picker.querySelector("[data-theme-repo-picker-message]");
    const more = picker.querySelector("[data-theme-repo-picker-more]");
    const icons = picker.querySelector("[data-theme-repo-picker-icons]")?.content;
    const text = picker.querySelector("[data-theme-repo-picker-text]")?.dataset ?? {};
    const limit = Number(picker.dataset.limit) || 10;
    const current = picker.dataset.current;
    const state = { page: 1, request: undefined, revision: 0, loaded: false, timer: undefined };

    const icon = name => {
      const source = icons?.querySelector(`[data-icon="${name}"] > svg`);
      return source ? source.cloneNode(true) : document.createElement("span");
    };

    const row = repo => {
      const item = document.createElement("li");
      const link = document.createElement("a");
      link.className = "theme-repo-picker-item";
      // Only same-origin paths; the search answers with the repository's link path.
      link.href = String(repo.link || "").startsWith("/") ? repo.link : "#";
      if (current && String(repo.id) === current) link.setAttribute("aria-current", "true");
      link.append(icon(iconFor(repo)));
      const name = document.createElement("span");
      name.className = "theme-repo-picker-name";
      const fullName = String(repo.full_name || "");
      name.textContent = "fullName" in picker.dataset ? fullName : fullName.slice(fullName.indexOf("/") + 1);
      name.title = fullName;
      link.append(name);
      if (current && String(repo.id) === current) {
        const check = icon("check");
        check.classList?.add("theme-repo-picker-check");
        link.append(check);
      }
      item.append(link);
      return item;
    };

    const load = async append => {
      state.request?.abort();
      state.request = new AbortController();
      const revision = ++state.revision;
      const params = new URLSearchParams({
        sort: "updated",
        order: "desc",
        uid: picker.dataset.uid ?? "",
        q: query?.value.trim() ?? "",
        page: String(state.page),
        limit: String(limit),
      });
      if (picker.dataset.team) params.set("team_id", picker.dataset.team);
      if ("exclusive" in picker.dataset) params.set("exclusive", "1");
      list.setAttribute("aria-busy", "true");
      message.hidden = true;
      try {
        const response = await fetch(`${picker.dataset.url}?${params}`, {
          signal: state.request.signal,
          headers: { Accept: "application/json" },
        });
        if (!response.ok) throw new Error("Repository search failed");
        const body = await response.json();
        if (revision !== state.revision) return;
        const repos = (body?.data ?? []).map(result => result.repository).filter(Boolean);
        if (!append) list.replaceChildren();
        list.append(...repos.map(row));
        // The first answer fades in; later filtering and Show more change the list in place.
        if (!state.loaded) list.classList.add("theme-fade-in");
        const total = Number(response.headers.get("X-Total-Count"));
        if (more) more.hidden = !(total > state.page * limit);
        if (!list.children.length) {
          message.textContent = text.empty ?? "";
          message.hidden = false;
        }
        state.loaded = true;
      } catch (error) {
        if (error.name === "AbortError" || revision !== state.revision) return;
        message.textContent = text.error ?? "";
        message.hidden = false;
      } finally {
        if (revision === state.revision) list.removeAttribute("aria-busy");
      }
    };

    query?.addEventListener("input", () => {
      clearTimeout(state.timer);
      state.timer = setTimeout(() => {
        state.page = 1;
        void load(false);
      }, 250);
    });
    query?.addEventListener("keydown", event => {
      // Enter opens the first match, as GitHub's switcher does.
      if (event.key !== "Enter") return;
      event.preventDefault();
      const first = list.querySelector("a.theme-repo-picker-item");
      if (first) location.href = first.href;
    });
    more?.addEventListener("click", () => {
      state.page += 1;
      void load(true);
    });

    const api = {
      open: () => {
        if (!state.loaded) void load(false);
      },
      showSearch: () => {
        const field = query?.closest(".theme-repo-picker-search");
        if (field) field.hidden = false;
        query?.focus();
      },
    };
    pickers.set(picker, api);
    return api;
  };

  for (const picker of document.querySelectorAll("[data-theme-repo-picker]")) {
    const api = setup(picker);
    if (!("lazy" in picker.dataset)) api.open();
  }

  // A lazy picker loads when the dropdown around it opens, then focuses its field.
  document.addEventListener(
    "toggle",
    event => {
      const dropdown = event.target;
      if (!dropdown.open) return;
      const picker = dropdown.querySelector?.("[data-theme-repo-picker][data-lazy]");
      if (!picker) return;
      const api = setup(picker);
      api.open();
      api.showSearch();
    },
    true
  );

  // A button outside the picker reveals its filter field (the search icon beside "Top repositories").
  document.addEventListener("click", event => {
    const toggle = event.target.closest?.("[data-theme-repo-picker-toggle]");
    if (!toggle) return;
    const picker = document.getElementById(toggle.getAttribute("aria-controls"));
    if (picker) setup(picker).showSearch();
  });
})();
