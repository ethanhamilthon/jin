const query = typeof matchMedia === "undefined" ? null : matchMedia("(max-width: 700px)");

// mobile is true on a phone-size window: one pane, the sidebar as a drawer.
export const mobile = $state({ on: query?.matches ?? false });
query?.addEventListener("change", (e) => (mobile.on = e.matches));

// followViewport keeps the page as high as the visible part, so the on-screen
// keyboard of iOS does not cover the composer.
export function followViewport() {
  const view = window.visualViewport;
  if (!view) return;
  const fit = () => {
    document.documentElement.style.setProperty("--app-h", view.height + "px");
    if (mobile.on && view.offsetTop) window.scrollTo(0, 0);
  };
  fit();
  view.addEventListener("resize", fit);
  view.addEventListener("scroll", fit);
}
