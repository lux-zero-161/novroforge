(() => {
  // GitHub's README outline: the rendered headings, indented by level, built when the menu first opens.
  const outline = document.querySelector("[data-readme-outline]");
  const menu = outline?.closest("details");
  const markup = document.querySelector("#readme .markup");
  if (!outline || !menu || !markup) return;
  const headings = [...markup.querySelectorAll("h1[id], h2[id], h3[id], h4[id], h5[id], h6[id]")];
  if (!headings.length) {
    menu.hidden = true;
    return;
  }
  menu.addEventListener(
    "toggle",
    () => {
      for (const heading of headings) {
        const link = document.createElement("a");
        link.className = "theme-dropdown-action";
        link.href = `#${heading.id.replace(/^user-content-/, "")}`;
        link.dataset.level = heading.tagName.slice(1);
        link.textContent = heading.textContent.trim();
        outline.before(link);
      }
    },
    { once: true }
  );
  outline.addEventListener("click", event => {
    if (event.target.closest("a")) menu.open = false;
  });
})();
