(() => {
  // The clear button empties the field; a submitted search that is currently applied reloads without it.
  document.addEventListener("click", event => {
    const clear = event.target.closest(".theme-search-clear");
    if (!clear) return;
    event.preventDefault();
    const input = clear.closest(".theme-search-field")?.querySelector('input[type="search"]');
    if (!input) return;
    const applied = Boolean(input.name) && input.defaultValue !== "";
    input.value = "";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    input.focus();
    if (applied && input.form) input.form.requestSubmit();
  });

  // A field with a shortcut takes focus when its key is pressed outside another text field.
  document.addEventListener("keydown", event => {
    if (event.ctrlKey || event.metaKey || event.altKey || event.isComposing) return;
    if (event.target.closest("input, textarea, select, [contenteditable]")) return;
    const input = document.querySelector(`input[data-search-shortcut="${CSS.escape(event.key)}"]`);
    if (!input || !input.checkVisibility()) return;
    event.preventDefault();
    input.focus();
  });

  // Qualifiers Forgejo's issue search understands; their values are highlighted like GitHub's.
  const qualifiers = new Set(["is", "author", "assignee", "review", "mentions", "sort", "modified"]);

  function setup(input) {
    const control = input.closest(".theme-search-syntax-control");
    const highlight = control?.querySelector(".theme-search-highlight");
    const popup = control?.querySelector(".theme-search-syntax");
    if (!highlight || !popup) return;
    const items = [...popup.querySelectorAll(".theme-search-syntax-item")];
    let active = -1;

    function render() {
      highlight.replaceChildren();
      for (const part of input.value.split(/(\s+)/)) {
        const match = /^(-?)([a-z]+):(.*)$/i.exec(part);
        if (!match || !qualifiers.has(match[2].toLowerCase())) {
          highlight.append(part);
          continue;
        }
        const value = document.createElement("span");
        value.className = "theme-search-qualifier-value";
        value.textContent = match[3];
        highlight.append(`${match[1]}${match[2]}:`, value);
      }
      highlight.scrollLeft = input.scrollLeft;
    }

    // The word under the caret decides which qualifiers are offered.
    function currentToken() {
      const end = input.selectionStart ?? input.value.length;
      const start = input.value.lastIndexOf(" ", end - 1) + 1;
      return { start, end, text: input.value.slice(start, end) };
    }
    const visibleItems = () => items.filter(item => !item.hidden);
    function setActive(index) {
      active = index;
      const target = visibleItems()[active];
      items.forEach(item => item.classList.toggle("active", item === target));
      target?.scrollIntoView({ block: "nearest" });
    }
    function update() {
      const text = currentToken().text.toLowerCase();
      let any = false;
      for (const item of items) {
        const name = item.textContent.trim().toLowerCase();
        item.hidden = !(item.dataset.insert.toLowerCase().startsWith(text) || name.startsWith(text));
        any ||= !item.hidden;
      }
      popup.hidden = document.activeElement !== input || text.includes(":") || !any;
      input.setAttribute("aria-expanded", String(!popup.hidden));
      setActive(-1);
    }
    function close() {
      popup.hidden = true;
      input.setAttribute("aria-expanded", "false");
    }
    function insert(item) {
      const { start, end } = currentToken();
      const text = item.dataset.insert;
      input.value = input.value.slice(0, start) + text + input.value.slice(end);
      const caret = start + text.length;
      input.setSelectionRange(caret, caret);
      input.focus();
      render();
      update();
    }

    input.addEventListener("input", () => {
      render();
      update();
    });
    input.addEventListener("scroll", () => (highlight.scrollLeft = input.scrollLeft));
    input.addEventListener("focus", update);
    input.addEventListener("click", update);
    input.addEventListener("keyup", event => {
      if (["ArrowLeft", "ArrowRight", "Home", "End"].includes(event.key)) update();
      highlight.scrollLeft = input.scrollLeft;
    });
    input.addEventListener("blur", close);
    input.addEventListener("keydown", event => {
      if (popup.hidden) return;
      const visible = visibleItems();
      if (event.key === "ArrowDown" || event.key === "ArrowUp") {
        event.preventDefault();
        const step = event.key === "ArrowDown" ? 1 : -1;
        setActive((active + step + visible.length) % visible.length);
      } else if (event.key === "Enter" && active >= 0) {
        event.preventDefault();
        insert(visible[active]);
      } else if (event.key === "Escape") {
        close();
      }
    });
    // Keep the focus in the field while picking a qualifier.
    popup.addEventListener("mousedown", event => {
      if (!event.target.closest(".theme-search-syntax-docs")) event.preventDefault();
    });
    popup.addEventListener("click", event => {
      const item = event.target.closest(".theme-search-syntax-item");
      if (item) insert(item);
    });
    render();
  }

  for (const input of document.querySelectorAll("input[data-search-syntax]")) setup(input);
})();
