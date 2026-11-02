(() => {
  // GitHub keeps Comment disabled until there is text or an attachment to send.
  const form = document.querySelector(".github-conversation .comment.form");
  const textarea = form?.querySelector("textarea.markdown-text-editor");
  const submit = form?.querySelector(".button-sequence > .primary.button");
  if (!textarea || !submit) return;
  const files = form.querySelector(".dropzone .files");
  const updateSubmit = () => {
    submit.disabled = !textarea.value.trim() && !files?.children.length;
  };
  textarea.addEventListener("input", updateSubmit);
  if (files) new MutationObserver(updateSubmit).observe(files, { childList: true, subtree: true });
  updateSubmit();
})();
