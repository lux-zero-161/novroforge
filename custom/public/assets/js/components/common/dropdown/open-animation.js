(() => {
  // Menus built on <details> open like Fomantic's dropdowns ("Add file"), sliding down while fading in.
  // A closed <details> keeps its content hidden with content-visibility, so a CSS animation would only play on the
  // first open, and the toggle event fires a frame after the menu shows (it would jump). Start from the summary
  // click instead, in the frame that first paints the open menu.
  const menus = "details.dropdown, details.theme-menu, details.theme-dropdown";
  const keyframes = [
    { opacity: 0, transform: "translateY(-8px)" },
    { opacity: 1, transform: "translateY(0)" },
  ];
  document.addEventListener("click", event => {
    const summary = event.target.closest("summary");
    const menu = summary?.parentElement;
    if (!menu?.matches(menus) || menu.open || matchMedia("(prefers-reduced-motion: reduce)").matches) return;
    const panel = [...menu.children].find(child => child !== summary);
    if (!panel) return;
    requestAnimationFrame(() => {
      if (menu.open) panel.animate(keyframes, { duration: 200, easing: "cubic-bezier(0.33, 1, 0.68, 1)" });
    });
  });
})();
