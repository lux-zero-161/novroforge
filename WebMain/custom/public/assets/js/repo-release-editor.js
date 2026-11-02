// GitHub's "Choose a tag" menu on the new release form (templates/repo/release/tag_menu.tmpl). NovroForge's tag name field
// sits at the top of the menu: typing filters the repository's tags and offers to create the typed one, a tag row fills
// the field. NovroForge listens for keyup on the field to update the helper text, the target and the title, so a picked tag
// sends one too. The shared dropdown controller closes the menu after a row is clicked.
(() => {
  const input = document.getElementById("tag-name");
  const picker = input?.closest(".theme-release-tag-picker");
  if (!picker) return;
  const details = picker.querySelector(".theme-dropdown");
  const label = picker.querySelector("[data-theme-dropdown-label]");
  const create = picker.querySelector(".theme-release-tag-create");
  const createText = create?.querySelector("span");
  const tags = [...picker.querySelectorAll("[data-tag]")];

  const update = () => {
    const value = input.value;
    const query = value.trim().toLowerCase();
    let exists = false;
    for (const tag of tags) {
      const name = tag.dataset.tag;
      tag.hidden = !name.toLowerCase().includes(query);
      if (name === value) {
        exists = true;
        tag.setAttribute("aria-current", "true");
      } else tag.removeAttribute("aria-current");
    }
    if (create) {
      create.hidden = !value.trim() || exists;
      if (createText) createText.textContent = (create.dataset.template ?? "%s").replace("%s", value);
    }
    if (label) label.textContent = value || picker.dataset.emptyLabel || "";
  };

  const notifyNovroForge = () => input.dispatchEvent(new KeyboardEvent("keyup"));

  // NovroForge only listens for keyup; a pasted or dropped name reaches it too.
  input.addEventListener("input", () => {
    update();
    notifyNovroForge();
  });
  input.addEventListener("keydown", event => {
    // Enter keeps the typed name instead of submitting the form, as GitHub's menu does.
    if (event.key !== "Enter") return;
    event.preventDefault();
    details.open = false;
    details.querySelector("summary")?.focus();
  });
  // A required field inside a closed menu cannot show the browser's message; open the menu first.
  input.addEventListener("invalid", () => {
    details.open = true;
  });
  details.addEventListener("toggle", () => {
    if (details.open) input.focus();
  });
  picker.addEventListener("click", event => {
    const tag = event.target.closest?.("[data-tag]");
    if (!tag) return;
    input.value = tag.dataset.tag;
    update();
    notifyNovroForge();
    details.querySelector("summary")?.focus();
  });
  update();
})();
