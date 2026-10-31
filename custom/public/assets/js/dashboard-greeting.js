(() => {
  // GitHub greets by the visitor's time of day; the server cannot know it, so the heading starts as "Hello".
  const heading = document.querySelector("[data-theme-greeting]");
  if (!heading) return;
  const hour = new Date().getHours();
  const part = hour < 5 ? "evening" : hour < 12 ? "morning" : hour < 18 ? "afternoon" : "evening";
  heading.textContent = `Good ${part}, ${heading.dataset.name}!`;
})();
